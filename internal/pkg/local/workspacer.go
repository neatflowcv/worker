package local

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
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

	return domain.NewWorkspace(projectDir, mainDir), nil
}

func (w *Workspacer) CreateWorktree(
	ctx context.Context,
	project *domain.Project,
	workspace *domain.Workspace,
	item *domain.BacklogItem,
) (*domain.Worktree, error) {
	_ = project

	branch := item.ID()
	worktreeDir := filepath.Join(workspace.Root(), branch)
	worktreePathArg := filepath.Join("..", branch)

	//nolint:gosec // Git worktree command uses repository-owned paths and backlog IDs.
	command := exec.CommandContext(
		ctx,
		"git",
		"worktree",
		"add",
		"-b",
		branch,
		worktreePathArg,
	)
	command.Dir = workspace.Main()

	output, err := command.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git worktree add: %w: %s", err, output)
	}

	return domain.NewWorktree(branch, worktreeDir), nil
}

func (w *Workspacer) CloseWorktree(
	ctx context.Context,
	project *domain.Project,
	worktree *domain.Worktree,
) error {
	mainDir := filepath.Join(filepath.Dir(worktree.Dir()), "main")

	repository, err := git.PlainOpen(mainDir)
	if err != nil {
		return fmt.Errorf("open repository: %w", err)
	}

	err = repository.PushContext(ctx, newPushOptions(project, worktree))
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return fmt.Errorf("push branch: %w", err)
	}

	return nil
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

func newPushOptions(project *domain.Project, worktree *domain.Worktree) *git.PushOptions {
	refSpec := config.RefSpec(
		plumbing.NewBranchReferenceName(worktree.Branch()) +
			":" +
			plumbing.NewBranchReferenceName(worktree.Branch()),
	)

	return &git.PushOptions{
		RemoteName:        git.DefaultRemoteName,
		RemoteURL:         "",
		RefSpecs:          []config.RefSpec{refSpec},
		Auth:              newAuthMethod(project.Auth()),
		Progress:          nil,
		Prune:             false,
		Force:             false,
		InsecureSkipTLS:   false,
		ClientCert:        nil,
		ClientKey:         nil,
		CABundle:          nil,
		RequireRemoteRefs: nil,
		FollowTags:        false,
		ForceWithLease:    nil,
		Options:           nil,
		Atomic:            false,
		ProxyOptions: transport.ProxyOptions{
			URL:      "",
			Username: "",
			Password: "",
		},
	}
}

func newAuthMethod(auth *domain.Auth) *githttp.BasicAuth {
	if auth == nil {
		return nil
	}

	return &githttp.BasicAuth{
		Username: auth.Username(),
		Password: auth.Password(),
	}
}
