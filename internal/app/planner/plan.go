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
var ErrPlanMarkdownRequired = errors.New("plan markdown is required")
var ErrPlanAnswerRequired = errors.New("plan answer is required")

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

type RefinePlanRequest struct {
	RootDir  string
	Git      string
	Markdown string
	Answers  []QuestionAnswer
}

type QuestionAnswer struct {
	Question string
	Answer   string
}

func (r RefinePlanRequest) validate() error {
	if strings.TrimSpace(r.RootDir) == "" {
		return ErrPlanRootDirRequired
	}

	if strings.TrimSpace(r.Git) == "" {
		return ErrPlanGitRequired
	}

	if strings.TrimSpace(r.Markdown) == "" {
		return ErrPlanMarkdownRequired
	}

	if len(r.Answers) == 0 {
		return ErrPlanAnswerRequired
	}

	for _, answer := range r.Answers {
		if strings.TrimSpace(answer.Answer) == "" {
			return ErrPlanAnswerRequired
		}
	}

	return nil
}

type RefinePlanResponse struct {
	Title    string
	Git      string
	Items    []decider.Item
	Markdown string
}
