package gogit

import (
	"context"
	"errors"
	"fmt"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/neatflowcv/worker/internal/pkg/gitter"
)

var _ gitter.Gitter = (*Gitter)(nil)

type Gitter struct{}

func New() *Gitter {
	return &Gitter{}
}

func (g *Gitter) CloneRepository(ctx context.Context, repositoryURL string, destinationDir string) error {
	_, err := git.PlainCloneContext(ctx, destinationDir, false, newCloneOptions(repositoryURL))
	if err == nil {
		return nil
	}

	if errors.Is(err, git.ErrRepositoryAlreadyExists) {
		return nil
	}

	return fmt.Errorf("clone repository: %w", err)
}

func (g *Gitter) PullRepository(ctx context.Context, repositoryDir string) error {
	repository, err := git.PlainOpen(repositoryDir)
	if err != nil {
		return fmt.Errorf("open repository: %w", err)
	}

	worktree, err := repository.Worktree()
	if err != nil {
		return fmt.Errorf("worktree: %w", err)
	}

	err = worktree.PullContext(ctx, newPullOptions())
	if err == nil || errors.Is(err, git.NoErrAlreadyUpToDate) {
		return nil
	}

	return fmt.Errorf("pull repository: %w", err)
}

func newCloneOptions(repositoryURL string) *git.CloneOptions {
	return &git.CloneOptions{
		URL:               repositoryURL,
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
