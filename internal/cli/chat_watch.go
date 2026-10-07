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
	"syscall"
	"time"

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
	Use:   "watch [space_id]",
	Short: "Continuously monitor a Google Chat space for new messages with debouncing",
	Args:  cobra.ExactArgs(1),
	RunE:  runChatWatch,
}

func runChatWatch(cmd *cobra.Command, args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	svc, err := chat.NewService(ctx)
	if err != nil {
		return err
	}
	space := chat.NormalizeSpaceName(args[0])
	buffer := chat.NewBurstBuffer(watchDebounce)

	if err := seedInitialMessages(svc, space, buffer); err != nil {
		return err
	}
	fmt.Printf("Watching %s (interval: %s, debounce: %s)...\n", space, watchInterval, watchDebounce)
	return watchLoop(ctx, svc, space, buffer)
}

func seedInitialMessages(svc *chat.Service, space string, buffer *chat.BurstBuffer) error {
	initial, err := svc.ListMessagesWithOrder(space, 20, "DESC")
	if err != nil {
		return fmt.Errorf("seed initial messages: %w", err)
	}
	buffer.SeedIDs(initial)
	return nil
}

func watchLoop(ctx context.Context, svc *chat.Service, space string, buffer *chat.BurstBuffer) error {
	pollInterval := watchInterval
	if pollInterval <= 0 {
		pollInterval = 10 * time.Second
	}
	ticker := time.NewTicker(pollInterval)
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
	title := fmt.Sprintf("Google Chat (%s)", space)
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
	return []string{
		"GWS_SPACE_ID=" + space,
		"GWS_MESSAGE_COUNT=" + strconv.Itoa(len(msgs)),
		"GWS_MESSAGE_ID=" + latest.Name,
		"GWS_SENDER=" + latest.SenderName,
		"GWS_SENDER_ID=" + latest.SenderID,
		"GWS_THREAD=" + latest.ThreadName,
		"GWS_TEXT=" + chat.FormatBurstText(msgs),
		"GWS_PAYLOAD=" + string(payload),
	}
}

func init() {
	chatWatchCmd.Flags().DurationVar(&watchInterval, "interval", 10*time.Second, "Polling interval (e.g. 5s, 10s, 30s)")
	chatWatchCmd.Flags().DurationVar(&watchDebounce, "debounce", 0, "Debounce duration to coalesce message bursts (e.g. 30s)")
	chatWatchCmd.Flags().BoolVar(&watchNotify, "notify", false, "Deliver native desktop notification on new message/burst")
	chatWatchCmd.Flags().StringVar(&watchExec, "exec", "", "Shell command to execute on each message or debounced burst")
	chatWatchCmd.Flags().BoolVar(&watchJSON, "json", false, "Stream JSON payloads to stdout")

	chatCmd.AddCommand(chatWatchCmd)
}
