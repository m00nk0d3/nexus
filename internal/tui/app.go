package tui

// App represents the main TUI application
type App struct {
	ready bool
}

// NewApp creates a new TUI application
func NewApp() *App {
	return &App{ready: false}
}

// Init initializes the application
func (a *App) Init() {
	a.ready = true
}
