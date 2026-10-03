package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
	"unicode"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	run(ctx, os.Stdout)
}

func run(ctx context.Context, stdout io.Writer) {
	fmt.Fprintln(stdout, lookupEmail(ctx))
}

// This helper uses only the standard library so installed bundles need no module files.
func lookupEmail(ctx context.Context) string {
	if ctx.Err() != nil {
		return "unknown"
	}
	command := exec.Command("codex", "app-server", "--listen", "stdio://")
	command.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	stdin, err := command.StdinPipe()
	if err != nil {
		return "unknown"
	}
	defer stdin.Close()
	stdout, err := command.StdoutPipe()
	if err != nil {
		return "unknown"
	}
	defer stdout.Close()
	// Nil stderr connects to the null device, avoiding diagnostics in the result.
	if err := command.Start(); err != nil {
		return "unknown"
	}
	defer stopServer(command)
	result := make(chan string, 1)
	go func() {
		result <- readAccount(stdin, stdout)
	}()
	select {
	case email := <-result:
		return email
	case <-ctx.Done():
		return "unknown"
	}
}

func stopServer(command *exec.Cmd) {
	group := -command.Process.Pid
	_ = syscall.Kill(group, syscall.SIGTERM)
	done := make(chan struct{})
	go func() {
		_ = command.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		_ = syscall.Kill(group, syscall.SIGKILL)
		<-done
	}
	// Also stop descendants when the server exits before they do.
	_ = syscall.Kill(group, syscall.SIGKILL)
}

func readAccount(stdin io.Writer, stdout io.Reader) string {
	encoder := json.NewEncoder(stdin)
	decoder := json.NewDecoder(stdout)
	if err := encoder.Encode(map[string]any{
		"method": "initialize", "id": 0,
		"params": map[string]any{"clientInfo": map[string]string{
			"name": "whoami", "version": "1.0.0",
		}},
	}); err != nil {
		return "unknown"
	}
	if _, err := readResponse(decoder, 0); err != nil {
		return "unknown"
	}
	if err := encoder.Encode(map[string]string{"method": "initialized"}); err != nil {
		return "unknown"
	}
	if err := encoder.Encode(map[string]any{
		"method": "account/read", "id": 1,
		"params": map[string]bool{"refreshToken": false},
	}); err != nil {
		return "unknown"
	}
	response, err := readResponse(decoder, 1)
	if err != nil {
		return "unknown"
	}
	var result struct {
		Account *struct {
			Type  string  `json:"type"`
			Email *string `json:"email"`
		} `json:"account"`
	}
	if err := json.Unmarshal(response, &result); err != nil || result.Account == nil || result.Account.Type != "chatgpt" || result.Account.Email == nil {
		return "unknown"
	}
	email := strings.TrimSpace(*result.Account.Email)
	if email == "" || strings.IndexFunc(email, func(character rune) bool {
		return unicode.IsControl(character) || character == '\u2028' || character == '\u2029'
	}) >= 0 {
		return "unknown"
	}
	return email
}

func readResponse(decoder *json.Decoder, requestID int) (json.RawMessage, error) {
	for {
		var message map[string]json.RawMessage
		if err := decoder.Decode(&message); err != nil {
			return nil, err
		}
		if message == nil {
			return nil, errors.New("invalid response")
		}
		var id int
		if raw, exists := message["id"]; !exists || string(raw) == "null" || json.Unmarshal(raw, &id) != nil || id != requestID {
			continue
		}
		if _, exists := message["error"]; exists {
			return nil, errors.New("account lookup failed")
		}
		return message["result"], nil
	}
}
