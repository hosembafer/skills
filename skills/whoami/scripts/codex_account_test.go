package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func fakeServer(t *testing.T, mode, reply string) string {
	t.Helper()
	directory := filepath.Join(t.TempDir(), "tools with spaces")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	testBinary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	quotedBinary := "'" + strings.ReplaceAll(testBinary, "'", "'\\''") + "'"
	script := "#!/bin/sh\nexec " + quotedBinary + " -test.run=^TestFakeAppServer$ -- \"$@\"\n"
	if err := os.WriteFile(filepath.Join(directory, "codex"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", directory)
	t.Setenv("SKILLS_WHOAMI_TEST_SERVER", mode)
	t.Setenv("SKILLS_WHOAMI_TEST_REPLY", reply)
	t.Setenv("SKILLS_WHOAMI_TEST_STATE", directory)
	return directory
}

// This subprocess speaks the real stdio protocol; it never calls the installed CLI.
func TestFakeAppServer(t *testing.T) {
	mode := os.Getenv("SKILLS_WHOAMI_TEST_SERVER")
	if mode == "" {
		return
	}
	if mode == "hang" {
		signal.Ignore(syscall.SIGTERM)
	}
	directory := os.Getenv("SKILLS_WHOAMI_TEST_STATE")
	write := func(name, content string) {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(content), 0o644); err != nil {
			os.Exit(2)
		}
	}
	write("pid", strconv.Itoa(os.Getpid()))
	group, err := syscall.Getpgid(0)
	if err != nil {
		os.Exit(2)
	}
	write("group", strconv.Itoa(group))
	args := os.Args[len(os.Args)-3:]
	write("args", strings.Join(args, "\n"))
	log, err := os.OpenFile(filepath.Join(directory, "requests"), os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		os.Exit(2)
	}
	decoder := json.NewDecoder(os.Stdin)
	for step := 0; ; step++ {
		var message map[string]any
		if err := decoder.Decode(&message); err != nil {
			os.Exit(0)
		}
		if err := json.NewEncoder(log).Encode(message); err != nil {
			os.Exit(2)
		}
		switch step {
		case 0:
			if mode == "init_error" {
				fmt.Fprintln(os.Stdout, `{"id":0,"error":{"message":"Server error details"}}`)
				continue
			}
			fmt.Fprintln(os.Stdout, `{"method":"account/updated","params":{"authMode":"chatgpt"}}`)
			fmt.Fprintln(os.Stdout, `{"id":99,"result":{"unused":true}}`)
			fmt.Fprintln(os.Stdout, `{"id":0,"result":{"userAgent":"fixture","codexHome":"/tmp/fixture-codex","platformFamily":"unix","platformOs":"macos"}}`)
		case 2:
			write("ready", "ready")
			fmt.Fprintln(os.Stderr, "Server debug details must not appear in helper output.")
			if mode == "hang" {
				time.Sleep(time.Hour)
			}
			if mode == "closed" {
				os.Exit(0)
			}
			fmt.Fprintln(os.Stdout, `{"method":"account/updated","params":{"authMode":"chatgpt"}}`)
			fmt.Fprintln(os.Stdout, os.Getenv("SKILLS_WHOAMI_TEST_REPLY"))
		}
	}
}

func assertProtocol(t *testing.T, directory string, wantRequests int) {
	t.Helper()
	args, err := os.ReadFile(filepath.Join(directory, "args"))
	if err != nil || string(args) != "app-server\n--listen\nstdio://" {
		t.Fatalf("app-server arguments = %q; error = %v", args, err)
	}
	data, err := os.ReadFile(filepath.Join(directory, "requests"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != wantRequests {
		t.Fatalf("received %d requests, want %d: %s", len(lines), wantRequests, data)
	}
	methods := []string{"initialize", "initialized", "account/read"}
	for i, line := range lines {
		var request map[string]any
		if err := json.Unmarshal([]byte(line), &request); err != nil {
			t.Fatal(err)
		}
		if request["method"] != methods[i] {
			t.Fatalf("request %d method = %v, want %s", i, request["method"], methods[i])
		}
		switch i {
		case 0:
			params, ok := request["params"].(map[string]any)
			if !ok {
				t.Fatal("initialize params missing")
			}
			info, ok := params["clientInfo"].(map[string]any)
			name, _ := info["name"].(string)
			version, _ := info["version"].(string)
			if !ok || request["id"] != float64(0) || name == "" || version == "" {
				t.Fatalf("invalid initialize request: %s", line)
			}
		case 1:
			if _, exists := request["id"]; exists {
				t.Fatalf("initialized must be a notification: %s", line)
			}
		case 2:
			params, ok := request["params"].(map[string]any)
			if !ok || request["id"] != float64(1) || params["refreshToken"] != false || len(params) != 1 {
				t.Fatalf("account/read must disable token refresh: %s", line)
			}
		}
	}
	pidText, err := os.ReadFile(filepath.Join(directory, "pid"))
	if err != nil {
		t.Fatal(err)
	}
	pid, err := strconv.Atoi(string(pidText))
	if err != nil {
		t.Fatal(err)
	}
	groupText, err := os.ReadFile(filepath.Join(directory, "group"))
	if err != nil || string(groupText) != string(pidText) {
		t.Fatalf("server must have its own process group: pid=%s, group=%s; error=%v", pidText, groupText, err)
	}
	if err := syscall.Kill(pid, 0); err != syscall.ESRCH {
		t.Fatalf("server process %d was not reaped: %v", pid, err)
	}
}

// An omitted handshake, extra method, exposed diagnostic, or unsafe email must fail these cases.
func TestAccountLookup(t *testing.T) {
	cases := []struct {
		name, mode, reply, output string
	}{
		{"chatgpt", "reply", `{"id":1,"result":{"account":{"type":"chatgpt","email":"developer@example.com","planType":"pro"},"requiresOpenaiAuth":true}}`, "developer@example.com\n"},
		{"trimmed_email", "reply", `{"id":1,"result":{"account":{"type":"chatgpt","email":"  developer@example.com  ","planType":"pro"},"requiresOpenaiAuth":true}}`, "developer@example.com\n"},
		{"api_key", "reply", `{"id":1,"result":{"account":{"type":"apiKey","email":"developer@example.com"},"requiresOpenaiAuth":true}}`, "unknown\n"},
		{"logged_out", "reply", `{"id":1,"result":{"account":null,"requiresOpenaiAuth":true}}`, "unknown\n"},
		{"null_email", "reply", `{"id":1,"result":{"account":{"type":"chatgpt","email":null,"planType":"pro"},"requiresOpenaiAuth":true}}`, "unknown\n"},
		{"missing_email", "reply", `{"id":1,"result":{"account":{"type":"chatgpt"}}}`, "unknown\n"},
		{"numeric_email", "reply", `{"id":1,"result":{"account":{"type":"chatgpt","email":42}}}`, "unknown\n"},
		{"empty_email", "reply", `{"id":1,"result":{"account":{"type":"chatgpt","email":"  "}}}`, "unknown\n"},
		{"multiline_email", "reply", `{"id":1,"result":{"account":{"type":"chatgpt","email":"developer@example.com\nextra"}}}`, "unknown\n"},
		{"unicode_line_separator", "reply", `{"id":1,"result":{"account":{"type":"chatgpt","email":"developer@example.com\u2028extra"}}}`, "unknown\n"},
		{"control_character", "reply", `{"id":1,"result":{"account":{"type":"chatgpt","email":"developer@example.com\u001bextra"}}}`, "unknown\n"},
		{"missing_account", "reply", `{"id":1,"result":{}}`, "unknown\n"},
		{"malformed_account", "reply", `{"id":1,"result":{"account":"bad"}}`, "unknown\n"},
		{"malformed_result", "reply", `{"id":1,"result":[]}`, "unknown\n"},
		{"rpc_error", "reply", `{"id":1,"error":{"message":"Server error details"},"result":{"account":{"type":"chatgpt","email":"developer@example.com"}}}`, "unknown\n"},
		{"malformed_json", "reply", "not JSON", "unknown\n"},
		{"non_object", "reply", "[]", "unknown\n"},
		{"null_response", "reply", "null", "unknown\n"},
		{"transport_closed", "closed", "", "unknown\n"},
		{"initialization_error", "init_error", "", "unknown\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			directory := fakeServer(t, tc.mode, tc.reply)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			var output bytes.Buffer
			run(ctx, &output)
			if output.String() != tc.output {
				t.Fatalf("output = %q, want %q", output.String(), tc.output)
			}
			requests := 3
			if tc.mode == "init_error" {
				requests = 1
			}
			assertProtocol(t, directory, requests)
		})
	}
}

func TestMissingCodex(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	var output bytes.Buffer
	run(context.Background(), &output)
	if output.String() != "unknown\n" {
		t.Fatalf("output = %q, want unknown", output.String())
	}
}

func TestAccountDeadline(t *testing.T) {
	directory := fakeServer(t, "hang", "")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var output bytes.Buffer
	done := make(chan struct{})
	go func() {
		run(ctx, &output)
		close(done)
	}()
	deadline := time.Now().Add(3 * time.Second)
	for {
		if _, err := os.Stat(filepath.Join(directory, "ready")); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("server never reached account/read")
		}
		time.Sleep(5 * time.Millisecond)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("helper did not stop a server that ignores SIGTERM")
	}
	if output.String() != "unknown\n" {
		t.Fatalf("output = %q, want unknown", output.String())
	}
	assertProtocol(t, directory, 3)
}
