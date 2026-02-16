// Package runrecord handles persistence of GoGrid agent run results.
// Records are stored as JSON files under .gogrid/runs/.
package runrecord

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/lonestarx1/gogrid/pkg/cost"
	"github.com/lonestarx1/gogrid/pkg/llm"
	"github.com/lonestarx1/gogrid/pkg/trace"
)

const runsDir = ".gogrid/runs"

// Record captures the complete result of a single agent run.
type Record struct {
	RunID       string        `json:"run_id"`
	Agent       string        `json:"agent"`
	Model       string        `json:"model"`
	Provider    string        `json:"provider"`
	Input       string        `json:"input"`
	Output      string        `json:"output"`
	Turns       int           `json:"turns"`
	Usage       llm.Usage     `json:"usage"`
	Cost        float64       `json:"cost"`
	Spans       []*trace.Span `json:"spans"`
	CostRecords []cost.Record `json:"cost_records"`
	StartTime   time.Time     `json:"start_time"`
	Duration    time.Duration `json:"duration"`
	Error       string        `json:"error,omitempty"`
}

// Save persists a record to .gogrid/runs/<run-id>.json relative to baseDir.
func Save(baseDir string, rec *Record) error {
	if rec.RunID == "" {
		return fmt.Errorf("runrecord: run ID is required")
	}

	dir := filepath.Join(baseDir, runsDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("runrecord: create dir: %w", err)
	}

	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return fmt.Errorf("runrecord: marshal: %w", err)
	}

	path := filepath.Join(dir, rec.RunID+".json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("runrecord: write: %w", err)
	}

	return nil
}

// Load reads a record from .gogrid/runs/<runID>.json relative to baseDir.
func Load(baseDir, runID string) (*Record, error) {
	path := filepath.Join(baseDir, runsDir, runID+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("runrecord: read %s: %w", runID, err)
	}

	var rec Record
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, fmt.Errorf("runrecord: unmarshal %s: %w", runID, err)
	}

	return &rec, nil
}

// List returns all run IDs sorted by descending order (newest first).
// IDs are time-sortable, so lexicographic descending order gives newest first.
func List(baseDir string) ([]string, error) {
	dir := filepath.Join(baseDir, runsDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("runrecord: list: %w", err)
	}

	var ids []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if ext := filepath.Ext(name); ext == ".json" {
			ids = append(ids, name[:len(name)-len(ext)])
		}
	}

	sort.Sort(sort.Reverse(sort.StringSlice(ids)))
	return ids, nil
}
