package exec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGitCommand(t *testing.T) {
	tests := []struct {
		name     string
		repoPath string
		expected string
	}{
		{
			name:     "creates git command executor",
			repoPath: "/home/user/repo",
			expected: "/home/user/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewGitCommand(tt.repoPath)
			assert.NotNil(t, cmd)
			assert.Equal(t, tt.expected, cmd.repoPath)
		})
	}
}
