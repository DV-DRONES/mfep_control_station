package cache

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCollectGarbage(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"remove.json":        `{"garbage": true}`,
		"nested/remove.yaml": "garbage: true\n",
		"keep-false.json":    `{"garbage": false}`,
		"keep-nested.json":   `{"metadata": {"garbage": true}}`,
		"keep.txt":           "garbage: true",
		"keep-string.yaml":   "garbage: \"true\"\n",
	}
	for name, contents := range files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	deleted, err := CollectGarbage(root)
	if err == nil {
		t.Fatal("expected an error for the non-boolean YAML garbage value")
	}
	if len(deleted) != 2 {
		t.Fatalf("expected 2 deleted files, got %d: %v", len(deleted), deleted)
	}
	for _, name := range []string{"remove.json", filepath.Join("nested", "remove.yaml")} {
		if _, err := os.Stat(filepath.Join(root, name)); !os.IsNotExist(err) {
			t.Errorf("expected %q to be deleted, stat error: %v", name, err)
		}
	}
	for _, name := range []string{"keep-false.json", "keep-nested.json", "keep.txt", "keep-string.yaml"} {
		if _, err := os.Stat(filepath.Join(root, name)); err != nil {
			t.Errorf("expected %q to remain, stat error: %v", name, err)
		}
	}
}

func TestCollectGarbageReportsInvalidJSONAndKeepsFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "broken.json")
	if err := os.WriteFile(path, []byte(`{"garbage":`), 0o600); err != nil {
		t.Fatal(err)
	}

	deleted, err := CollectGarbage(root)
	if err == nil {
		t.Fatal("expected an error for invalid JSON")
	}
	if len(deleted) != 0 {
		t.Fatalf("expected no deleted files, got %v", deleted)
	}
	if _, statErr := os.Stat(path); statErr != nil {
		t.Fatalf("expected invalid file to remain: %v", statErr)
	}
}

func TestCommandHandlerCollectsOnlyConfiguredDirectory(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "remove.json")
	if err := os.WriteFile(path, []byte(`{"garbage":true}`), 0o600); err != nil {
		t.Fatal(err)
	}
	handler, err := NewCommandHandler(root)
	if err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodPost, "/collect", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body)
	}
	var response collectResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Deleted) != 1 || filepath.Base(response.Deleted[0]) != filepath.Base(path) {
		t.Fatalf("unexpected deleted paths: %v", response.Deleted)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected configured file to be deleted, stat error: %v", err)
	}
}

func TestCommandHandlerRejectsNonPostAndRequestBody(t *testing.T) {
	handler, err := NewCommandHandler(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		name   string
		method string
		body   string
		want   int
	}{
		{name: "get", method: http.MethodGet, want: http.StatusMethodNotAllowed},
		{name: "body", method: http.MethodPost, body: "unexpected", want: http.StatusBadRequest},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, "/collect", strings.NewReader(test.body))
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)
			if recorder.Code != test.want {
				t.Fatalf("expected status %d, got %d", test.want, recorder.Code)
			}
		})
	}
}
