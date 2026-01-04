package session

import "testing"

func TestIsReservedName(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"Moderator", true},
		{"moderator", false}, // case-sensitive
		{"Alice", false},
		{"Bob", false},
		{"", false},
		{"MODERATOR", false},
		{"Mod", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsReservedName(tt.name)
			if result != tt.expected {
				t.Errorf("IsReservedName(%q) = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}
