package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDatabase(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "creates database with path",
			path:     ":memory:",
			expected: ":memory:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := NewDatabase(tt.path)
			assert.NotNil(t, db)
			assert.Equal(t, tt.expected, db.path)
		})
	}
}
