package system

import "testing"

func TestResolveIPBlacklistStatus(t *testing.T) {
	cases := []struct {
		name      string
		rawStatus string
		rawEndAt  string
		want      string
	}{
		{name: "active", rawStatus: "active", rawEndAt: "2099-12-31 23:59:59", want: "active"},
		{name: "expired", rawStatus: "active", rawEndAt: "2000-01-01 00:00:00", want: "expired"},
		{name: "manual inactive", rawStatus: "inactive", rawEndAt: "2099-12-31 23:59:59", want: "manual_inactive"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveIPBlacklistStatus(tc.rawStatus, tc.rawEndAt); got != tc.want {
				t.Fatalf("resolveIPBlacklistStatus() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestNormalizeBlacklistIPs(t *testing.T) {
	ips, err := normalizeBlacklistIPs([]string{" 1.2.3.4 ", "", "2001:db8::1"})
	if err != nil {
		t.Fatalf("normalizeBlacklistIPs() unexpected error: %v", err)
	}
	if len(ips) != 2 || ips[0] != "1.2.3.4" || ips[1] != "2001:db8::1" {
		t.Fatalf("normalizeBlacklistIPs() = %#v", ips)
	}

	if _, err := normalizeBlacklistIPs([]string{"1.2.3.0/24"}); err == nil {
		t.Fatal("normalizeBlacklistIPs() expected cidr error")
	}
	if _, err := normalizeBlacklistIPs([]string{"999.1.1.1"}); err == nil {
		t.Fatal("normalizeBlacklistIPs() expected invalid ip error")
	}
	if _, err := normalizeBlacklistIPs([]string{" "}); err == nil {
		t.Fatal("normalizeBlacklistIPs() expected empty list error")
	}
}
