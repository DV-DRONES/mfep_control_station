package wifi

import (
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSendAndReceiveOne(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "settings.yaml")
	output := filepath.Join(root, "received")
	contents := []byte("device:\n  mode: test\n")
	if err := os.WriteFile(source, contents, 0o600); err != nil {
		t.Fatal(err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	type result struct {
		path string
		err  error
	}
	received := make(chan result, 1)
	go func() {
		path, err := ReceiveOne(listener, output, 5*time.Second)
		received <- result{path: path, err: err}
	}()

	if err := Send(listener.Addr().String(), source); err != nil {
		t.Fatal(err)
	}
	got := <-received
	if got.err != nil {
		t.Fatal(got.err)
	}
	actual, err := os.ReadFile(got.path)
	if err != nil {
		t.Fatal(err)
	}
	if string(actual) != string(contents) {
		t.Fatalf("received %q, want %q", actual, contents)
	}
}

func TestRejectsUnsupportedOrUnsafeNames(t *testing.T) {
	for _, name := range []string{"../escape.json", "notes.txt", `folder\file.yaml`} {
		if err := validateFileName(name); err == nil {
			t.Errorf("validateFileName(%q) succeeded; want error", name)
		}
	}
	if err := validateFileName("settings.YML"); err != nil {
		t.Errorf("validateFileName accepted YAML extension with uppercase: %v", err)
	}
}

func TestServeListenerAcceptsRepeatedTransfers(t *testing.T) {
	root := t.TempDir()
	output := filepath.Join(root, "received")
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	serveResult := make(chan error, 1)
	go func() {
		serveResult <- serveListener(listener, output)
	}()

	for index, contents := range []string{"first", "second"} {
		source := filepath.Join(root, []string{"settings1.json", "settings2.json"}[index])
		if err := os.WriteFile(source, []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := Send(listener.Addr().String(), source); err != nil {
			t.Fatal(err)
		}
		destination := filepath.Join(output, filepath.Base(source))
		deadline := time.Now().Add(5 * time.Second)
		for {
			received, err := os.ReadFile(destination)
			if err == nil && string(received) == contents {
				break
			}
			if time.Now().After(deadline) {
				t.Fatalf("received file %q did not contain %q; last read error: %v", destination, contents, err)
			}
			time.Sleep(time.Millisecond)
		}
	}

	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	if err := <-serveResult; err != nil {
		t.Fatal(err)
	}
}
