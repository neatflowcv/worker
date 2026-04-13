package local

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/neatflowcv/worker/internal/pkg/domain"
	workspacerpkg "github.com/neatflowcv/worker/internal/pkg/workspacer"
)

var _ workspacerpkg.Workspacer = (*Workspacer)(nil)

const projectDirMode = 0o750

type Workspacer struct {
	rootDir string
}

func NewWorkspacer(rootDir string) *Workspacer {
	return &Workspacer{
		rootDir: rootDir,
	}
}

func (w *Workspacer) PrepareWorkspace(ctx context.Context, project *domain.Project) (*domain.Workspace, error) {
	projectDir := filepath.Join(w.rootDir, project.ID())

	err := os.MkdirAll(projectDir, projectDirMode)
	if err != nil {
		return nil, fmt.Errorf("create project directory: %w", err)
	}

	mainDir := filepath.Join(projectDir, "main")

	err = w.ensureRepository(ctx, mainDir, project)
	if err != nil {
		return nil, err
	}

	return domain.NewWorkspace(projectDir, mainDir, nil), nil
}

func (w *Workspacer) ensureRepository(ctx context.Context, mainDir string, project *domain.Project) error {
	_, err := git.PlainCloneContext(ctx, mainDir, false, newCloneOptions(project))
	if err == nil {
		return nil
	}

	if !errors.Is(err, git.ErrRepositoryAlreadyExists) {
		return fmt.Errorf("clone repository: %w", err)
	}

	repository, openErr := git.PlainOpen(mainDir)
	if openErr != nil {
		return fmt.Errorf("open repository: %w", openErr)
	}

	worktree, worktreeErr := repository.Worktree()
	if worktreeErr != nil {
		return fmt.Errorf("worktree: %w", worktreeErr)
	}

	pullErr := worktree.PullContext(ctx, newPullOptions())
	if pullErr != nil && !errors.Is(pullErr, git.NoErrAlreadyUpToDate) {
		return fmt.Errorf("pull repository: %w", pullErr)
	}

	return nil
}

func newCloneOptions(project *domain.Project) *git.CloneOptions {
	return &git.CloneOptions{
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
	}
}

func newPullOptions() *git.PullOptions {
	return &git.PullOptions{
		RemoteName:        git.DefaultRemoteName,
		RemoteURL:         "",
		ReferenceName:     plumbing.ReferenceName(""),
		SingleBranch:      false,
		Force:             false,
		Depth:             0,
		RecurseSubmodules: git.NoRecurseSubmodules,
		Auth:              nil,
		Progress:          nil,
		InsecureSkipTLS:   false,
		ClientCert:        nil,
		ClientKey:         nil,
		CABundle:          nil,
		ProxyOptions: transport.ProxyOptions{
			URL:      "",
			Username: "",
			Password: "",
		},
	}
}
