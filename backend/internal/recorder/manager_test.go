package recorder

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"backend/internal/config"
)

func TestSafeDirBlocksTraversal(t *testing.T) {
	m := NewManager(config.Config{RecordingPath: "/tmp/visioncore-test-rec"}, nil, nil)
	if _, err := m.safeDir("../evil", time.Now()); err == nil {
		t.Fatal("expected traversal blocked for ../evil")
	}
	if _, err := m.safeDir("cam 1/..", time.Now()); err == nil {
		t.Fatal("expected invalid id rejected")
	}
	dir, err := m.safeDir("cam1", time.Date(2026, 9, 28, 15, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(filepath.ToSlash(dir), "cam1/2026/09/28/15") {
		t.Fatal("unexpected layout:", dir)
	}
}
