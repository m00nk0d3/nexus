-- context_snapshots: add diff_summary column to persist git diff --stat output
ALTER TABLE context_snapshots ADD COLUMN diff_summary TEXT;
