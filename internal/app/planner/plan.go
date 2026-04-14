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
var ErrPlanSourcesRequired = errors.New("plan sources are required")
var ErrPlanSourceReferenceRequired = errors.New("plan source reference is required")
var ErrInvalidSourceKind = errors.New("invalid source kind")

type SourceKind string

const (
	SourceKindGit SourceKind = "Git"
	SourceKindURL SourceKind = "URL"
)

type Source struct {
	Kind      SourceKind
	Reference string
}

type CreatePlanRequest struct {
	RootDir string
	Title   string
	Sources []Source
}

func (r CreatePlanRequest) validate() error {
	if strings.TrimSpace(r.Title) == "" {
		return ErrPlanTitleRequired
	}

	if len(r.Sources) == 0 {
		return ErrPlanSourcesRequired
	}

	for _, source := range r.Sources {
		if strings.TrimSpace(source.Reference) == "" {
			return ErrPlanSourceReferenceRequired
		}

		if !source.Kind.isValid() {
			return ErrInvalidSourceKind
		}

		if source.Kind == SourceKindGit && strings.TrimSpace(r.RootDir) == "" {
			return ErrPlanRootDirRequired
		}
	}

	return nil
}

func normalizeRootDir(rootDir string) string {
	return filepath.Clean(strings.TrimSpace(rootDir))
}

type CreatePlanResponse struct {
	Title    string
	Sources  []Source
	Items    []decider.Item
	Markdown string
}
