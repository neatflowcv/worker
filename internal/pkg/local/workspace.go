package local

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/neatflowcv/worker/internal/pkg/domain"
	"github.com/neatflowcv/worker/internal/pkg/workspace"
)

var _ workspace.Workspace = (*Workspace)(nil)

const projectDirMode = 0o750

type Workspace struct {
	rootDir string
}

func NewWorkspace(rootDir string) *Workspace {
	return &Workspace{
		rootDir: rootDir,
	}
}

func (w *Workspace) CreateWorkspace(ctx context.Context, project *domain.Project) error {
	projectDir := filepath.Join(w.rootDir, project.ID())

	err := os.MkdirAll(projectDir, projectDirMode)
	if err != nil {
		return fmt.Errorf("create project directory: %w", err)
	}

	mainDir := filepath.Join(projectDir, "main")

	_, err = git.PlainCloneContext(ctx, mainDir, false, &git.CloneOptions{
		URL:               project.RepositoryURL(),
		Auth:              nil,
		RemoteName:        git.DefaultRemoteName,
		ReferenceName:     plumbing.ReferenceName(""),
		SingleBranch:      false,
		Mirror:            false,
		NoCheckout:        false,
		Depth:             0,
		RecurseSubmodules: git.NoRecurseSubmodules,
		ShallowSubmodules: false,
		Progress:          nil,
		Tags:              git.AllTags,
		InsecureSkipTLS:   false,
		ClientCert:        nil,
		ClientKey:         nil,
		CABundle:          nil,
		ProxyOptions: transport.ProxyOptions{
			URL:      "",
			Username: "",
			Password: "",
		},
		Shared: false,
	})
	if err != nil {
		return fmt.Errorf("clone repository: %w", err)
	}

	return nil
}
