package domain

type Worktree struct {
	branch string
	dir    string
}

func NewWorktree(branch, dir string) *Worktree {
	return &Worktree{
		branch: branch,
		dir:    dir,
	}
}

func (w *Worktree) Branch() string {
	if w == nil {
		return ""
	}

	return w.branch
}

func (w *Worktree) Dir() string {
	if w == nil {
		return ""
	}

	return w.dir
}
