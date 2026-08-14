package domain

import (
	"testing"
	"time"
)

func TestBaseVersionLifecycle(t *testing.T) {
	now := time.Date(2026, time.August, 13, 12, 0, 0, 0, time.UTC)
	var base Base

	if err := base.InitForCreate(now); err != nil {
		t.Fatalf("InitForCreate() error = %v", err)
	}
	if base.Version != 1 {
		t.Fatalf("created version = %d, want 1", base.Version)
	}

	base.TouchForUpdate(now.Add(time.Minute))
	if base.Version != 2 {
		t.Fatalf("updated version = %d, want 2", base.Version)
	}
}
