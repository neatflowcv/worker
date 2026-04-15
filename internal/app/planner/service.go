package planner

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	deciderpkg "github.com/neatflowcv/worker/internal/pkg/decider"
	codexdecider "github.com/neatflowcv/worker/internal/pkg/decider/codex"
)

type Service struct {
	decider deciderpkg.Decider
}

func NewService(decider deciderpkg.Decider) *Service {
	return &Service{
		decider: decider,
	}
}

func NewCodexService() (*Service, error) {
	decider, err := codexdecider.NewDecider()
	if err != nil {
		return nil, fmt.Errorf("create codex decider: %w", err)
	}

	return NewService(decider), nil
}

func (s *Service) CreatePlan(request CreatePlanRequest) (*CreatePlanResponse, error) {
	err := request.validate()
	if err != nil {
		return nil, err
	}

	title := strings.TrimSpace(request.Title)

	gitDir, err := fetchGitSource(normalizeRootDir(request.RootDir), request.Git)
	if err != nil {
		return nil, err
	}

	decision, err := s.decider.Decide(deciderpkg.DecideRequest{
		Title:       title,
		Directories: []string{gitDir},
	})
	if err != nil {
		return nil, fmt.Errorf("decide plan: %w", err)
	}

	return &CreatePlanResponse{
		Title:    title,
		Git:      gitDir,
		Items:    decision.Items,
		Markdown: decision.Markdown,
	}, nil
}

func fetchGitSource(rootDir, reference string) (string, error) {
	reference = strings.TrimSpace(reference)
	localDir := filepath.Join(rootDir, repositoryDirName(reference))

	err := syncGitRepository(localDir, reference)
	if err != nil {
		return "", err
	}

	return localDir, nil
}

func repositoryDirName(reference string) string {
	name := filepath.Base(reference)
	name = strings.TrimSuffix(name, ".git")

	if name == "." || name == string(filepath.Separator) || name == "" {
		return "repository"
	}

	return name
}

func syncGitRepository(localDir, repositoryURL string) error {
	err := os.MkdirAll(filepath.Dir(localDir), sourceDirMode)
	if err != nil {
		return fmt.Errorf("create planner root directory: %w", err)
	}

	_, err = git.PlainClone(localDir, false, newCloneOptions(repositoryURL))
	if err == nil {
		return nil
	}

	if !errors.Is(err, git.ErrRepositoryAlreadyExists) {
		return fmt.Errorf("clone repository: %w", err)
	}

	repository, openErr := git.PlainOpen(localDir)
	if openErr != nil {
		return fmt.Errorf("open repository: %w", openErr)
	}

	worktree, worktreeErr := repository.Worktree()
	if worktreeErr != nil {
		return fmt.Errorf("worktree: %w", worktreeErr)
	}

	pullErr := worktree.Pull(newPullOptions())
	if pullErr != nil && !errors.Is(pullErr, git.NoErrAlreadyUpToDate) {
		return fmt.Errorf("pull repository: %w", pullErr)
	}

	return nil
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
		Depth:             0,
		Auth:              nil,
		RecurseSubmodules: git.NoRecurseSubmodules,
		Progress:          nil,
		Force:             false,
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
