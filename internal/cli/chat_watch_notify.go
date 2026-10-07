package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"

	"github.com/erikkubica/gws/internal/auth"
	"github.com/erikkubica/gws/internal/services/chat"
)

func sendOSNotification(title, message string) {
	if runtime.GOOS != "darwin" {
		return
	}
	_ = exec.Command(
		"osascript",
		"-e", "on run argv",
		"-e", "display notification (item 1 of argv) with title (item 2 of argv)",
		"-e", "end run",
		message,
		title,
	).Run()
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
		spaceID = chat.ExtractSpaceID(latest.Name)
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
