package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewApp(t *testing.T) {
	tests := []struct {
		name      string
		wantReady bool
	}{
		{
			name:      "creates new app with ready=false",
			wantReady: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApp()
			assert.NotNil(t, app)
			assert.Equal(t, tt.wantReady, app.ready)
		})
	}
}

func TestAppInit(t *testing.T) {
	tests := []struct {
		name      string
		wantReady bool
	}{
		{
			name:      "init sets ready to true",
			wantReady: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApp()
			app.Init()
			assert.Equal(t, tt.wantReady, app.ready)
		})
	}
}
