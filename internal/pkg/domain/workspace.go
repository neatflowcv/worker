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

func (w *Worktree) clone() *Worktree {
	if w == nil {
		return nil
	}

	return &Worktree{
		branch: w.branch,
		dir:    w.dir,
	}
}

type Workspace struct {
	root      string
	main      string
	worktrees []*Worktree
}

func NewWorkspace(root, main string, worktrees []*Worktree) *Workspace {
	clonedWorktrees := make([]*Worktree, 0, len(worktrees))
	for _, worktree := range worktrees {
		clonedWorktrees = append(clonedWorktrees, worktree.clone())
	}

	return &Workspace{
		root:      root,
		main:      main,
		worktrees: clonedWorktrees,
	}
}

func (w *Workspace) Root() string {
	return w.root
}

func (w *Workspace) Main() string {
	return w.main
}

func (w *Workspace) Worktrees() []*Worktree {
	worktrees := make([]*Worktree, 0, len(w.worktrees))
	for _, worktree := range w.worktrees {
		worktrees = append(worktrees, worktree.clone())
	}

	return worktrees
}
