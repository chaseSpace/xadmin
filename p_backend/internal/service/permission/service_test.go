package permission

import "testing"

func TestMenuIDsWithinScope(t *testing.T) {
	tests := []struct {
		name     string
		menuIDs  []int64
		scopeIDs []int64
		want     bool
	}{
		{name: "empty role permissions", menuIDs: nil, scopeIDs: []int64{1, 2}, want: true},
		{name: "same permissions", menuIDs: []int64{1, 2}, scopeIDs: []int64{1, 2}, want: true},
		{name: "strict subset", menuIDs: []int64{2}, scopeIDs: []int64{1, 2, 3}, want: true},
		{name: "permission outside scope", menuIDs: []int64{1, 4}, scopeIDs: []int64{1, 2, 3}, want: false},
		{name: "ignores invalid ids", menuIDs: []int64{0, -1, 2}, scopeIDs: []int64{2}, want: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := menuIDsWithinScope(test.menuIDs, test.scopeIDs); got != test.want {
				t.Fatalf("menuIDsWithinScope(%v, %v) = %v, want %v", test.menuIDs, test.scopeIDs, got, test.want)
			}
		})
	}
}
