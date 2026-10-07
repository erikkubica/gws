package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/erikkubica/gws/internal/auth"
	"github.com/erikkubica/gws/internal/services/chat"
	"github.com/spf13/cobra"
)

var (
	watchInterval time.Duration
	watchDebounce time.Duration
	watchNotify   bool
	watchExec     string
	watchJSON     bool
)

var chatWatchCmd = &cobra.Command{
	Use:   "watch [space_id|all]",
	Short: "Continuously monitor Google Chat space(s) for new messages with debouncing",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runChatWatch,
}

func runChatWatch(cmd *cobra.Command, args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	svc, err := chat.NewService(ctx)
	if err != nil {
		return err
	}
	target := "all"
	if len(args) > 0 && args[0] != "" && args[0] != "all" {
		target = chat.NormalizeSpaceName(args[0])
	}
	buffer := chat.NewBurstBuffer(watchDebounce)
	if target == "all" {
		return runWatchAll(ctx, svc, buffer)
	}
	return runWatchSingle(ctx, svc, target, buffer)
}

func runWatchSingle(ctx context.Context, svc *chat.Service, space string, buffer *chat.BurstBuffer) error {
	if err := seedInitialMessages(svc, space, buffer); err != nil {
		return err
	}
	fmt.Printf("Watching %s (interval: %s, debounce: %s)...\n", space, watchInterval, watchDebounce)
	return watchLoop(ctx, svc, space, buffer)
}

func runWatchAll(ctx context.Context, svc *chat.Service, buffer *chat.BurstBuffer) error {
	activities := make(map[string]string)
	if err := seedAllSpaces(svc, buffer, activities); err != nil {
		return err
	}
	fmt.Printf("Watching ALL conversations (interval: %s, debounce: %s)...\n", watchInterval, watchDebounce)
	return watchAllLoop(ctx, svc, buffer, activities)
}

func seedInitialMessages(svc *chat.Service, space string, buffer *chat.BurstBuffer) error {
	initial, err := svc.ListMessagesWithOrder(space, 20, "DESC")
	if err != nil {
		return fmt.Errorf("seed initial messages: %w", err)
	}
	buffer.SeedIDs(initial)
	return nil
}

func seedAllSpaces(svc *chat.Service, buffer *chat.BurstBuffer, activities map[string]string) error {
	spaces, err := svc.ListSpaceActivities(50)
	if err != nil {
		return fmt.Errorf("seed all spaces: %w", err)
	}
	for _, sp := range spaces {
		activities[sp.Name] = sp.LastActiveTime
	}
	for i := 0; i < len(spaces) && i < 10; i++ {
		msgs, _ := svc.ListMessagesWithOrder(spaces[i].Name, 5, "DESC")
		buffer.SeedIDs(msgs)
	}
	return nil
}

func watchLoop(ctx context.Context, svc *chat.Service, space string, buffer *chat.BurstBuffer) error {
	ticker := time.NewTicker(resolvePollInterval())
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("\nStopping watcher...")
			return nil
		case <-ticker.C:
			pollAndFlush(svc, space, buffer)
		}
	}
}

func watchAllLoop(ctx context.Context, svc *chat.Service, buf *chat.BurstBuffer, act map[string]string) error {
	ticker := time.NewTicker(resolvePollInterval())
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("\nStopping watcher...")
			return nil
		case <-ticker.C:
			pollAndFlushAll(svc, buf, act)
		}
	}
}

func resolvePollInterval() time.Duration {
	if watchInterval <= 0 {
		return 10 * time.Second
	}
	return watchInterval
}

func pollAndFlush(svc *chat.Service, space string, buffer *chat.BurstBuffer) {
	recent, err := svc.ListMessagesWithOrder(space, 10, "DESC")
	if err == nil {
		for i := len(recent) - 1; i >= 0; i-- {
			buffer.Add(recent[i], time.Now())
		}
	}
	if buffer.ShouldFlush(time.Now()) {
		if flushed := buffer.Flush(); len(flushed) > 0 {
			dispatchBurst(space, flushed)
		}
	}
}

func pollAndFlushAll(svc *chat.Service, buffer *chat.BurstBuffer, activities map[string]string) {
	current, err := svc.ListSpaceActivities(50)
	if err == nil {
		for _, sp := range current {
			if prevTime, exists := activities[sp.Name]; !exists || sp.LastActiveTime != prevTime {
				activities[sp.Name] = sp.LastActiveTime
				msgs, _ := svc.ListMessagesWithOrder(sp.Name, 5, "DESC")
				for i := len(msgs) - 1; i >= 0; i-- {
					buffer.Add(msgs[i], time.Now())
				}
			}
		}
	}
	if buffer.ShouldFlush(time.Now()) {
		if flushed := buffer.Flush(); len(flushed) > 0 {
			dispatchBurst("all", flushed)
		}
	}
}

func dispatchBurst(space string, msgs []*chat.MessageInfo) {
	payload, _ := json.Marshal(msgs)
	if watchJSON {
		fmt.Println(string(payload))
	} else {
		printFlushedBurst(msgs)
	}
	sendNotificationIfNeeded(space, msgs)
	executeCommandIfNeeded(space, msgs, payload)
}

func printFlushedBurst(msgs []*chat.MessageInfo) {
	fmt.Printf("\n--- Flushed Burst (%d message(s)) ---\n", len(msgs))
	for _, m := range msgs {
		printMessageItem(m)
	}
}

func sendNotificationIfNeeded(space string, msgs []*chat.MessageInfo) {
	if !watchNotify || len(msgs) == 0 {
		return
	}
	latest := msgs[len(msgs)-1]
	displaySpace := space
	if space == "all" {
		displaySpace = extractSpaceID(latest.Name)
	}
	title := fmt.Sprintf("Google Chat (%s)", displaySpace)
	body := fmt.Sprintf("%s: %s", latest.SenderName, latest.Text)
	if len(msgs) > 1 {
		body = fmt.Sprintf("[%d messages] Latest from %s: %s", len(msgs), latest.SenderName, latest.Text)
	}
	sendOSNotification(title, body)
}

func sendOSNotification(title, message string) {
	if runtime.GOOS != "darwin" {
		return
	}
	script := fmt.Sprintf("display notification %q with title %q", message, title)
	_ = exec.Command("osascript", "-e", script).Run()
}

func executeCommandIfNeeded(space string, msgs []*chat.MessageInfo, payload []byte) {
	if watchExec == "" || len(msgs) == 0 {
		return
	}
	cmd := exec.Command("sh", "-c", watchExec)
	cmd.Env = append(os.Environ(), buildWatchEnv(space, msgs, payload)...)
	cmd.Stdin = bytes.NewReader(payload)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing watch command: %v\n", err)
	}
}

func buildWatchEnv(space string, msgs []*chat.MessageInfo, payload []byte) []string {
	latest := msgs[len(msgs)-1]
	currentAcc, _ := auth.GetActiveAccount()
	if auth.SelectedAccount != "" {
		currentAcc = auth.SelectedAccount
	}
	spaceID := space
	if spaceID == "all" && latest != nil {
		spaceID = extractSpaceID(latest.Name)
	}
	return []string{
		"GWS_ACCOUNT=" + currentAcc,
		"GWS_SPACE_ID=" + spaceID,
		"GWS_MESSAGE_COUNT=" + strconv.Itoa(len(msgs)),
		"GWS_MESSAGE_ID=" + latest.Name,
		"GWS_SENDER=" + latest.SenderName,
		"GWS_SENDER_ID=" + latest.SenderID,
		"GWS_THREAD=" + latest.ThreadName,
		"GWS_TEXT=" + chat.FormatBurstText(msgs),
		"GWS_PAYLOAD=" + string(payload),
	}
}

func extractSpaceID(messageResourceName string) string {
	parts := strings.Split(messageResourceName, "/")
	if len(parts) >= 2 {
		return parts[0] + "/" + parts[1]
	}
	return messageResourceName
}

func init() {
	chatWatchCmd.Flags().DurationVar(&watchInterval, "interval", 10*time.Second, "Polling interval (e.g. 5s, 10s, 30s)")
	chatWatchCmd.Flags().DurationVar(&watchDebounce, "debounce", 0, "Debounce duration to coalesce message bursts (e.g. 30s)")
	chatWatchCmd.Flags().BoolVar(&watchNotify, "notify", false, "Deliver native desktop notification on new message/burst")
	chatWatchCmd.Flags().StringVar(&watchExec, "exec", "", "Shell command to execute on each message or debounced burst")
	chatWatchCmd.Flags().BoolVar(&watchJSON, "json", false, "Stream JSON payloads to stdout")

	chatCmd.AddCommand(chatWatchCmd)
}
