package audit

import (
	"context"
	"os"
	"testing"
)

func TestStore_RecordAndList(t *testing.T) {
	tmpFile := "/tmp/test_audit.db"
	defer os.Remove(tmpFile)

	store, err := NewStore(tmpFile)
	if err != nil {
		t.Fatalf("create store failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Record some entries
	if err := store.Record(ctx, "start", "container-1", "user1", "test detail"); err != nil {
		t.Fatalf("record failed: %v", err)
	}
	if err := store.Record(ctx, "stop", "container-2", "user2", ""); err != nil {
		t.Fatalf("record failed: %v", err)
	}

	// List records
	records, err := store.List(ctx, 10)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}

	if len(records) != 2 {
		t.Errorf("expected 2 records, got %d", len(records))
	}

	// Check order (most recent first)
	if records[0].Action != "stop" {
		t.Errorf("expected first record action 'stop', got %s", records[0].Action)
	}
}

func TestStore_ListLimit(t *testing.T) {
	tmpFile := "/tmp/test_audit_limit.db"
	defer os.Remove(tmpFile)

	store, err := NewStore(tmpFile)
	if err != nil {
		t.Fatalf("create store failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Record 5 entries
	for i := 0; i < 5; i++ {
		_ = store.Record(ctx, "action", "target", "user", "")
	}

	// List with limit
	records, err := store.List(ctx, 3)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}

	if len(records) != 3 {
		t.Errorf("expected 3 records, got %d", len(records))
	}
}
