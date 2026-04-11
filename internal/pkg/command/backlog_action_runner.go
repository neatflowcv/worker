package command

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/neatflowcv/worker/internal/pkg/domain"
)

const refineFileMode = 0o600

type backlogActionRunner struct {
	runner Runner
}

func NewBacklogActionRunner() *backlogActionRunner {
	return NewBacklogActionRunnerWithRunner(&ExecRunner{})
}

func NewBacklogActionRunnerWithRunner(runner Runner) *backlogActionRunner {
	return &backlogActionRunner{
		runner: runner,
	}
}

func (r *backlogActionRunner) RefineBacklogItem(
	ctx context.Context,
	projectDir string,
	item *domain.BacklogItem,
) (*domain.BacklogItem, error) {
	fileName := item.ID() + ".md"
	filePath := filepath.Join(projectDir, fileName)

	err := os.WriteFile(filePath, []byte(item.Description()), refineFileMode)
	if err != nil {
		return nil, fmt.Errorf("write refine file: %w", err)
	}

	defer func() {
		_ = os.Remove(filePath)
	}()

	prompt := buildBacklogRefinePrompt(fileName, item)

	err = r.runner.Run(ctx, projectDir, prompt)
	if err != nil {
		return nil, fmt.Errorf("execute codex refine: %w", err)
	}

	//nolint:gosec // The path is confined to projectDir and uses a sanitized filename.
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read refined file: %w", err)
	}

	return item.SetDescription(string(content)), nil
}

func buildBacklogRefinePrompt(fileName string, item *domain.BacklogItem) string {
	if item.Description() == "" {
		return fmt.Sprintf(
			"Read the backlog item markdown file %q. "+
				"The file is empty, so write an initial draft for the backlog item based on its title %q. "+
				"Create a concise, actionable markdown draft with clear structure. "+
				"Modify the file in place. Do not rename the file.",
			fileName,
			item.Title(),
		)
	}

	return fmt.Sprintf(
		"Read and refine the backlog item markdown file %q. "+
			"Improve clarity and structure while preserving intent. "+
			"Modify the file in place. Do not rename the file.",
		fileName,
	)
}
