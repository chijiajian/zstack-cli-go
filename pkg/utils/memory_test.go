package utils

import "testing"

func TestParseMemorySizeOverflow(t *testing.T) {
	for _, size := range []string{"8388608T", "16777217T", "9223372036854775808", "9007199254740992K"} {
		if value, err := ParseMemorySize(size); err == nil {
			t.Fatalf("accepted overflowing %s as %d", size, value)
		}
	}
	for input, want := range map[string]int64{"1G": 1 << 30, " 2gb ": 2 << 30, "9223372036854775807": 1<<63 - 1, "0": 0, "1024M": 1 << 30} {
		got, err := ParseMemorySize(input)
		if err != nil || got != want {
			t.Fatalf("%s: %d %v", input, got, err)
		}
	}
}
