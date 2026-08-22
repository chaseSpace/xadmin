package organization

import (
	"strings"
	"testing"
)

func TestNormalizeUsername(t *testing.T) {
	for _, input := range []string{"Admin", "a1", "user_name", "user-name", "  Alice_01  "} {
		t.Run("valid_"+input, func(t *testing.T) {
			got, err := normalizeUsername(input)
			if err != nil {
				t.Fatalf("expected valid username, got %v", err)
			}
			if got != strings.TrimSpace(input) {
				t.Fatalf("expected trimmed username %q, got %q", strings.TrimSpace(input), got)
			}
		})
	}

	for _, input := range []string{"", "1admin", "_admin", "-admin", "user.name", "用户", "a b", "a@b", "a" + strings.Repeat("b", 64)} {
		t.Run("invalid_"+input, func(t *testing.T) {
			if _, err := normalizeUsername(input); err == nil {
				t.Fatalf("expected username %q to be rejected", input)
			}
		})
	}
}
