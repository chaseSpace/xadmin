package ipblacklist

import (
	"net/netip"
	"strings"
	"sync"
	"time"
)

type Entry struct {
	ID      int64
	IP      string
	StartAt time.Time
	EndAt   time.Time
}

type Match struct {
	ID int64
	IP string
}

type Store struct {
	mu        sync.RWMutex
	entries   map[string]map[int64]Entry
	idIndex   map[int64]string
	hitDeltas map[int64]int32
}

func NewStore() *Store {
	return &Store{
		entries:   make(map[string]map[int64]Entry),
		idIndex:   make(map[int64]string),
		hitDeltas: make(map[int64]int32),
	}
}

var defaultStore = NewStore()

func DefaultStore() *Store {
	return defaultStore
}

func (s *Store) Replace(entries []Entry) {
	next := make(map[string]map[int64]Entry, len(entries))
	nextIndex := make(map[int64]string, len(entries))
	for _, entry := range entries {
		normalized, ok := normalizeEntry(entry)
		if !ok {
			continue
		}
		putEntry(next, normalized)
		nextIndex[normalized.ID] = normalized.IP
	}

	s.mu.Lock()
	s.entries = next
	s.idIndex = nextIndex
	s.mu.Unlock()
}

func (s *Store) Add(entry Entry) bool {
	normalized, ok := normalizeEntry(entry)
	if !ok {
		return false
	}
	s.mu.Lock()
	s.putEntryLocked(normalized)
	s.mu.Unlock()
	return true
}

func (s *Store) AddMany(entries []Entry) int {
	if len(entries) == 0 {
		return 0
	}
	next := make(map[string]map[int64]Entry, len(entries))
	for _, entry := range entries {
		normalized, ok := normalizeEntry(entry)
		if !ok {
			continue
		}
		putEntry(next, normalized)
	}
	if len(next) == 0 {
		return 0
	}
	s.mu.Lock()
	count := 0
	for _, byID := range next {
		for _, entry := range byID {
			s.putEntryLocked(entry)
			count++
		}
	}
	s.mu.Unlock()
	return count
}

func (s *Store) Delete(rawIPs ...string) int {
	if len(rawIPs) == 0 {
		return 0
	}
	keys := make([]string, 0, len(rawIPs))
	for _, rawIP := range rawIPs {
		key := normalizeIP(rawIP)
		if key == "" {
			continue
		}
		keys = appendIfMissing(keys, key)
	}
	if len(keys) == 0 {
		return 0
	}
	deleted := 0
	s.mu.Lock()
	for _, key := range keys {
		if byID, ok := s.entries[key]; ok {
			deleted += len(byID)
			for id := range byID {
				delete(s.idIndex, id)
			}
			delete(s.entries, key)
		}
	}
	s.mu.Unlock()
	return deleted
}

func (s *Store) DeleteByID(ids ...int64) int {
	if len(ids) == 0 {
		return 0
	}
	wanted := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		if id > 0 {
			wanted[id] = struct{}{}
		}
	}
	if len(wanted) == 0 {
		return 0
	}
	deleted := 0
	s.mu.Lock()
	for id := range wanted {
		if s.deleteByIDLocked(id) {
			deleted++
		}
	}
	s.mu.Unlock()
	return deleted
}

func (s *Store) Match(rawIP string, now time.Time) (Match, bool) {
	candidates := CandidateIPs(rawIP)
	if len(candidates) == 0 {
		return Match{}, false
	}

	var match Match
	s.mu.RLock()
	for _, ip := range candidates {
		byID, ok := s.entries[ip]
		if !ok {
			continue
		}
		entry, ok := mergedActiveEntry(byID, now)
		if ok {
			match = Match{ID: entry.ID, IP: entry.IP}
			break
		}
	}
	s.mu.RUnlock()

	if match.ID <= 0 {
		return Match{}, false
	}

	s.mu.Lock()
	s.hitDeltas[match.ID]++
	s.mu.Unlock()
	return match, true
}

func (s *Store) HitDelta(id int64) int32 {
	if id <= 0 {
		return 0
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.hitDeltas[id]
}

func (s *Store) HitDeltas(ids []int64) map[int64]int32 {
	deltas := make(map[int64]int32, len(ids))
	if len(ids) == 0 {
		return deltas
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if delta := s.hitDeltas[id]; delta > 0 {
			deltas[id] = delta
		}
	}
	return deltas
}

func (s *Store) FlushHitDeltas() map[int64]int32 {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.hitDeltas) == 0 {
		return nil
	}
	flushed := s.hitDeltas
	s.hitDeltas = make(map[int64]int32)
	return flushed
}

func (e Entry) activeAt(now time.Time) bool {
	if e.StartAt.IsZero() || e.EndAt.IsZero() {
		return false
	}
	return !now.Before(e.StartAt) && !now.After(e.EndAt)
}

func normalizeEntry(entry Entry) (Entry, bool) {
	key := normalizeIP(entry.IP)
	if key == "" || entry.ID <= 0 {
		return Entry{}, false
	}
	entry.IP = key
	return entry, true
}

func putEntry(entries map[string]map[int64]Entry, entry Entry) {
	if entries[entry.IP] == nil {
		entries[entry.IP] = make(map[int64]Entry)
	}
	entries[entry.IP][entry.ID] = entry
}

func (s *Store) putEntryLocked(entry Entry) {
	if previousIP := s.idIndex[entry.ID]; previousIP != "" && previousIP != entry.IP {
		if byID := s.entries[previousIP]; byID != nil {
			delete(byID, entry.ID)
			if len(byID) == 0 {
				delete(s.entries, previousIP)
			}
		}
	}
	putEntry(s.entries, entry)
	s.idIndex[entry.ID] = entry.IP
}

func (s *Store) deleteByIDLocked(id int64) bool {
	ip := s.idIndex[id]
	if ip == "" {
		return false
	}
	byID := s.entries[ip]
	if byID == nil {
		delete(s.idIndex, id)
		return false
	}
	if _, ok := byID[id]; !ok {
		delete(s.idIndex, id)
		return false
	}
	delete(byID, id)
	delete(s.idIndex, id)
	if len(byID) == 0 {
		delete(s.entries, ip)
	}
	return true
}

func mergedActiveEntry(entries map[int64]Entry, now time.Time) (Entry, bool) {
	var selected Entry
	for _, entry := range entries {
		if !entry.activeAt(now) {
			continue
		}
		if selected.ID == 0 || entry.EndAt.After(selected.EndAt) || (entry.EndAt.Equal(selected.EndAt) && entry.ID > selected.ID) {
			selected = entry
		}
	}
	return selected, selected.ID > 0
}

func CandidateIPs(raw string) []string {
	ip := normalizeIP(raw)
	if ip == "" {
		return nil
	}
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return nil
	}
	ips := []string{addr.String()}
	if addr.IsLoopback() {
		ips = appendIfMissing(ips, "127.0.0.1")
		ips = appendIfMissing(ips, "::1")
	}
	return ips
}

func normalizeIP(raw string) string {
	input := strings.TrimSpace(raw)
	if input == "" {
		return ""
	}
	addr, err := netip.ParseAddr(input)
	if err != nil {
		return ""
	}
	return addr.Unmap().String()
}

func appendIfMissing(items []string, value string) []string {
	for _, item := range items {
		if item == value {
			return items
		}
	}
	return append(items, value)
}
