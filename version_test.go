package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFullVersion(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		commit   string
		expected string
	}{
		{
			name:     "default",
			version:  "0.0.1",
			commit:   "HEAD",
			expected: "0.0.1@HEAD",
		},
		{
			name:     "custom",
			version:  "1.2.3",
			commit:   "abc123",
			expected: "1.2.3@abc123",
		},
		{
			name:     "empty",
			version:  "",
			commit:   "",
			expected: "@",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Version = tt.version
			Commit = tt.commit
			actual := FullVersion()
			assert.Equal(t, tt.expected, actual)
		})
	}
}
