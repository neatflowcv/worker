package planner

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/neatflowcv/worker/internal/pkg/decider"
)

const sourceDirMode = 0o750

var ErrPlanTitleRequired = errors.New("plan title is required")
var ErrPlanRootDirRequired = errors.New("plan root directory is required")
var ErrPlanGitRequired = errors.New("plan git is required")

type CreatePlanRequest struct {
	RootDir string
	Git     string
	Title   string
}

func (r CreatePlanRequest) validate() error {
	if strings.TrimSpace(r.Title) == "" {
		return ErrPlanTitleRequired
	}

	if strings.TrimSpace(r.RootDir) == "" {
		return ErrPlanRootDirRequired
	}

	if strings.TrimSpace(r.Git) == "" {
		return ErrPlanGitRequired
	}

	return nil
}

func normalizeRootDir(rootDir string) string {
	return filepath.Clean(strings.TrimSpace(rootDir))
}

type CreatePlanResponse struct {
	Title    string
	Git      string
	Items    []decider.Item
	Markdown string
}
