package modal

// WorktreeCreateConfirmedMsg is sent when the user confirms creating a new worktree.
type WorktreeCreateConfirmedMsg struct {
	Branch string
	Path   string
}

// WorktreeDeleteConfirmedMsg is sent when the user confirms deleting a worktree.
type WorktreeDeleteConfirmedMsg struct {
	Path string
}

// ModalCancelledMsg is sent when the user cancels a modal (Esc or 'n').
type ModalCancelledMsg struct{}
