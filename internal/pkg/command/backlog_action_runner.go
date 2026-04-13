package command

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/neatflowcv/worker/internal/pkg/domain"
)

const backlogFileMode = 0o600

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

func (r *backlogActionRunner) StartBacklogItem(
	ctx context.Context,
	projectDir string,
	item *domain.BacklogItem,
) error {
	fileName := item.ID() + ".md"
	filePath := filepath.Join(projectDir, fileName)

	err := os.WriteFile(filePath, []byte(item.Description()), backlogFileMode)
	if err != nil {
		return fmt.Errorf("write start file: %w", err)
	}

	defer func() {
		_ = os.Remove(filePath)
	}()

	err = r.runner.Run(ctx, projectDir, buildBacklogStartPrompt(fileName, item))
	if err != nil {
		return fmt.Errorf("execute codex start: %w", err)
	}

	return nil
}

func (r *backlogActionRunner) RefineBacklogItem(
	ctx context.Context,
	projectDir string,
	item *domain.BacklogItem,
) (*domain.BacklogItem, error) {
	fileName := item.ID() + ".md"
	filePath := filepath.Join(projectDir, fileName)

	err := os.WriteFile(filePath, []byte(item.Description()), backlogFileMode)
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
			"%q 파일에 %q 에 대한 목표와 범위, 그리고 커밋 계획, 결정해야 할 사항 등을 포함한 계획 초안을 작성해. "+
				"특히, 결정해야 할 사항은 답변하기 쉬운 형식으로 만들어.",
			fileName,
			item.Title(),
		)
	}

	return fmt.Sprintf(
		"%q 파일의 계획을 더 명확하고 간결하게 정리해. 만약, 추가적으로 결정해야 할 사항이 있다면, 추가해.",
		fileName,
	)
}

func buildBacklogStartPrompt(fileName string, item *domain.BacklogItem) string {
	if item.Description() == "" {
		return fmt.Sprintf(
			"%q 작업을 수행해.",
			item.Title(),
		)
	}

	return fmt.Sprintf(
		"%q 파일에 명시된 작업을 수행해.",
		fileName,
	)
}
