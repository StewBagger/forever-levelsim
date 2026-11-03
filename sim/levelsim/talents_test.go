package levelsim

import "testing"

func TestTrimTalents(t *testing.T) {
	for _, tc := range []struct {
		talents string
		points  int32
		want    string
	}{
		// The deepest tree is spent first, row by row; trees that tie go in string order.
		{"05032-055-50532310253101", 11, "--5051"},
		{"05032-055-50532310253101", 0, ""},
		{"05032-055-50532310253101", 51, "05032-055-50532310253101"},
		{"05032-055-50532310253101", 31, "--50532310253101"},
		{"05032-055-50532310253101", 41, "05032--50532310253101"},
	} {
		if got := TrimTalents(tc.talents, tc.points); got != tc.want {
			t.Errorf("TrimTalents(%q, %d) = %q, want %q", tc.talents, tc.points, got, tc.want)
		}
	}
}
