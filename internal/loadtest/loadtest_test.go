package loadtest

import "testing"

func TestParseMemoryBytes(t *testing.T) {
	tests := map[string]uint64{
		"512m": 512000000,
		"1g":   1000000000,
		"2gb":  2000000000,
		"1GiB": 1073741824,
	}
	for input, want := range tests {
		got, err := ParseMemoryBytes(input)
		if err != nil {
			t.Fatalf("ParseMemoryBytes(%q): %v", input, err)
		}
		if got != want {
			t.Fatalf("ParseMemoryBytes(%q) = %d, want %d", input, got, want)
		}
	}
}

func TestParsePercent(t *testing.T) {
	got, err := ParsePercent("80%")
	if err != nil {
		t.Fatal(err)
	}
	if got != 0.8 {
		t.Fatalf("ParsePercent(80%%) = %f, want 0.8", got)
	}
}
