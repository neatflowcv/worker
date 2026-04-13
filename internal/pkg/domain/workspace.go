package domain

type Workspace struct {
	root string
	main string
}

func NewWorkspace(root, main string) *Workspace {
	return &Workspace{
		root: root,
		main: main,
	}
}

func (w *Workspace) Root() string {
	return w.root
}

func (w *Workspace) Main() string {
	return w.main
}
