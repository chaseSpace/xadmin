package ipblacklist

import (
	"testing"
	"time"
)

func TestStoreMatchUsesMemoryEntries(t *testing.T) {
	store := NewStore()
	now := time.Now()
	if ok := store.Add(Entry{ID: 1, IP: "127.0.0.1", StartAt: now.Add(-time.Minute), EndAt: now.Add(time.Minute)}); !ok {
		t.Fatal("expected add to succeed")
	}

	match, ok := store.Match("127.0.0.1", now)
	if !ok {
		t.Fatal("expected blacklist match")
	}
	if match.ID != 1 || match.IP != "127.0.0.1" {
		t.Fatalf("unexpected match: %#v", match)
	}
	if delta := store.HitDelta(1); delta != 1 {
		t.Fatalf("unexpected hit delta: %d", delta)
	}
}

func TestStoreDeleteRemovesMemoryEntry(t *testing.T) {
	store := NewStore()
	now := time.Now()
	store.Add(Entry{ID: 1, IP: "127.0.0.1", StartAt: now.Add(-time.Minute), EndAt: now.Add(time.Minute)})
	if deleted := store.Delete("127.0.0.1"); deleted != 1 {
		t.Fatalf("unexpected deleted count: %d", deleted)
	}
	if _, ok := store.Match("127.0.0.1", now); ok {
		t.Fatal("expected deleted entry to miss")
	}
}

func TestStoreMergesSameIPEntries(t *testing.T) {
	store := NewStore()
	now := time.Now()
	store.Replace([]Entry{
		{ID: 1, IP: "127.0.0.1", StartAt: now.Add(-time.Minute), EndAt: now.Add(time.Minute)},
		{ID: 2, IP: "127.0.0.1", StartAt: now.Add(-time.Minute), EndAt: now.Add(2 * time.Minute)},
	})

	match, ok := store.Match("127.0.0.1", now)
	if !ok {
		t.Fatal("expected merged blacklist match")
	}
	if match.ID != 2 {
		t.Fatalf("expected latest effective entry, got %#v", match)
	}
	if delta := store.HitDelta(2); delta != 1 {
		t.Fatalf("unexpected hit delta: %d", delta)
	}
	if deleted := store.DeleteByID(2); deleted != 1 {
		t.Fatalf("unexpected deleted count: %d", deleted)
	}
	match, ok = store.Match("127.0.0.1", now)
	if !ok || match.ID != 1 {
		t.Fatalf("expected fallback entry after delete, got %#v ok=%v", match, ok)
	}
}

func TestStoreAddMovesExistingIDToNewIP(t *testing.T) {
	store := NewStore()
	now := time.Now()
	store.Add(Entry{ID: 1, IP: "127.0.0.1", StartAt: now.Add(-time.Minute), EndAt: now.Add(time.Minute)})
	store.Add(Entry{ID: 1, IP: "127.0.0.2", StartAt: now.Add(-time.Minute), EndAt: now.Add(time.Minute)})

	if _, ok := store.Match("127.0.0.1", now); ok {
		t.Fatal("expected old ip to miss after moving existing id")
	}
	if match, ok := store.Match("127.0.0.2", now); !ok || match.ID != 1 {
		t.Fatalf("expected new ip to match, got %#v ok=%v", match, ok)
	}
}

func TestStoreMatchRejectsExpiredEntries(t *testing.T) {
	store := NewStore()
	now := time.Now()
	store.Replace([]Entry{
		{ID: 1, IP: "127.0.0.1", StartAt: now.Add(-time.Hour), EndAt: now.Add(-time.Minute)},
	})

	if _, ok := store.Match("127.0.0.1", now); ok {
		t.Fatal("expected expired entry to miss")
	}
}

func TestCandidateIPsIncludesLoopbackAliases(t *testing.T) {
	ips := CandidateIPs("::ffff:127.0.0.1")
	want := map[string]bool{"127.0.0.1": true, "::1": true}
	for _, ip := range ips {
		delete(want, ip)
	}
	if len(want) > 0 {
		t.Fatalf("missing loopback aliases: %#v in %#v", want, ips)
	}
}
