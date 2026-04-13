package command

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/neatflowcv/worker/internal/pkg/domain"
)

const backlogFileMode = 0o600

//go:embed recommend_worktree_output_schema.json
var recommendWorktreeOutputSchema []byte

var (
	errRecommendedWorktreeBranchNameEmpty    = errors.New("recommended worktree branch name is empty")
	errRecommendedWorktreeDirectoryNameEmpty = errors.New("recommended worktree directory name is empty")
)

//nolint:tagliatelle // Schema field names are fixed by the codex output schema.
type worktreeRecommendation struct {
	BranchName    string `json:"branch_name"`
	DirectoryName string `json:"directory_name"`
}

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
	fileName, cleanup, err := writeBacklogItemFile(projectDir, item)
	if err != nil {
		return fmt.Errorf("write start file: %w", err)
	}

	defer cleanup()

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
	fileName, cleanup, err := writeBacklogItemFile(projectDir, item)
	if err != nil {
		return nil, fmt.Errorf("write refine file: %w", err)
	}

	defer cleanup()

	prompt := buildBacklogRefinePrompt(fileName, item)

	err = r.runner.Run(ctx, projectDir, prompt)
	if err != nil {
		return nil, fmt.Errorf("execute codex refine: %w", err)
	}

	filePath := filepath.Join(projectDir, fileName)

	//nolint:gosec // The path is confined to projectDir and uses a sanitized filename.
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read refined file: %w", err)
	}

	return item.SetDescription(string(content)), nil
}

func (r *backlogActionRunner) RecommendWorktree(
	ctx context.Context,
	projectDir string,
	item *domain.BacklogItem,
) (*domain.Worktree, error) {
	fileName, cleanupBacklogFile, err := writeBacklogItemFile(projectDir, item)
	if err != nil {
		return nil, fmt.Errorf("write recommend file: %w", err)
	}

	defer cleanupBacklogFile()

	schemaPath, cleanupSchemaFile, err := writeSchemaTempFile()
	if err != nil {
		return nil, fmt.Errorf("write schema file: %w", err)
	}

	defer cleanupSchemaFile()

	outputFileName := item.ID() + ".json"
	outputPath := filepath.Join(projectDir, outputFileName)

	defer func() {
		_ = os.Remove(outputPath)
	}()

	err = r.runner.Run(
		ctx,
		projectDir,
		"--output-schema",
		schemaPath,
		buildRecommendWorktreePrompt(fileName),
		"-o",
		outputFileName,
	)
	if err != nil {
		return nil, fmt.Errorf("execute codex recommend worktree: %w", err)
	}

	recommendation, err := readWorktreeRecommendation(outputPath)
	if err != nil {
		return nil, err
	}

	return domain.NewWorktree(
		recommendation.BranchName,
		recommendation.DirectoryName,
	), nil
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

func buildRecommendWorktreePrompt(fileName string) string {
	return fmt.Sprintf(
		`Based on the contents of the %q file, suggest a branch name and directory name for a git worktree.`,
		fileName,
	)
}

func writeBacklogItemFile(
	projectDir string,
	item *domain.BacklogItem,
) (string, func(), error) {
	fileName := item.ID() + ".md"
	filePath := filepath.Join(projectDir, fileName)

	err := os.WriteFile(filePath, []byte(item.Description()), backlogFileMode)
	if err != nil {
		return "", nil, fmt.Errorf("write backlog item file: %w", err)
	}

	return fileName, func() {
		_ = os.Remove(filePath)
	}, nil
}

func writeSchemaTempFile() (string, func(), error) {
	file, err := os.CreateTemp("", "backlog-action-runner-schema-*.json")
	if err != nil {
		return "", nil, fmt.Errorf("create schema temp file: %w", err)
	}

	path := file.Name()
	cleanup := func() {
		_ = os.Remove(path)
	}

	_, err = file.Write(recommendWorktreeOutputSchema)
	closeErr := file.Close()

	if err != nil {
		cleanup()

		return "", nil, fmt.Errorf("write schema temp file: %w", err)
	}

	if closeErr != nil {
		cleanup()

		return "", nil, fmt.Errorf("close schema temp file: %w", closeErr)
	}

	return path, cleanup, nil
}

func readWorktreeRecommendation(outputPath string) (*worktreeRecommendation, error) {
	//nolint:gosec // The path is confined to projectDir and uses a sanitized filename.
	content, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, fmt.Errorf("read recommended worktree file: %w", err)
	}

	var recommendation worktreeRecommendation

	err = json.Unmarshal(content, &recommendation)
	if err != nil {
		return nil, fmt.Errorf("unmarshal recommended worktree file: %w", err)
	}

	if recommendation.BranchName == "" {
		return nil, errRecommendedWorktreeBranchNameEmpty
	}

	if recommendation.DirectoryName == "" {
		return nil, errRecommendedWorktreeDirectoryNameEmpty
	}

	return &recommendation, nil
}
