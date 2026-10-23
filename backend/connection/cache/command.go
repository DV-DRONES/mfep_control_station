package cache

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

type collectResponse struct {
	Deleted []string `json:"deleted"`
	Error   string   `json:"error,omitempty"`
}

func NewCommandHandler(root string) (http.Handler, error) {
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve collector root: %w", err)
	}
	absoluteRoot, err = filepath.EvalSymlinks(absoluteRoot)
	if err != nil {
		return nil, fmt.Errorf("resolve collector root: %w", err)
	}
	info, err := os.Stat(absoluteRoot)
	if err != nil {
		return nil, fmt.Errorf("inspect collector root: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("collector root %q is not a directory", absoluteRoot)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/collect", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.ContentLength != 0 {
			http.Error(w, "request body must be empty", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		deleted, err := CollectGarbage(absoluteRoot)
		response := collectResponse{Deleted: deleted}
		if err != nil {
			response.Error = err.Error()
			w.WriteHeader(http.StatusInternalServerError)
		}
		if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
			return
		}
	})
	return mux, nil
}
