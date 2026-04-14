package codex

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	deciderpkg "github.com/neatflowcv/worker/internal/pkg/decider"
)

//go:embed decision_output_schema.json
var decisionOutputSchema []byte

var (
	errDecisionMarkdownEmpty     = errors.New("decision markdown is empty")
	errDecisionDirectoriesEmpty  = errors.New("decision directories are required")
	errDecisionItemQuestionEmpty = errors.New("decision item question is empty")
)

var _ deciderpkg.Decider = (*Decider)(nil)

type Decider struct {
	runner     *Runner
	schemaPath string
}

func NewDecider() (*Decider, error) {
	schemaPath, err := writeSchemaFile()
	if err != nil {
		return nil, fmt.Errorf("write schema file: %w", err)
	}

	return &Decider{
		runner:     &Runner{},
		schemaPath: schemaPath,
	}, nil
}

func (d *Decider) Decide(request deciderpkg.DecideRequest) (*deciderpkg.Decision, error) {
	directories, err := directories(request.Directories)
	if err != nil {
		return nil, err
	}

	workDir := directories[0]

	output, err := d.runner.Run(
		context.Background(),
		workDir,
		"--output-schema",
		d.schemaPath,
		buildPrompt(request.Title, directories),
	)
	if err != nil {
		return nil, fmt.Errorf("execute codex decider: %w", err)
	}

	decision, err := readDecision(output)
	if err != nil {
		return nil, err
	}

	return decision, nil
}

func directories(raw []string) ([]string, error) {
	directories := make([]string, 0, len(raw))

	for _, dir := range raw {
		dir = strings.TrimSpace(dir)
		if dir == "" {
			continue
		}

		directories = append(directories, dir)
	}

	if len(directories) == 0 {
		return nil, errDecisionDirectoriesEmpty
	}

	return directories, nil
}

func buildPrompt(title string, directories []string) string {
	return fmt.Sprintf(
		"%q 작업에 대한 계획을 세워. "+
			"다음 디렉토리들을 검토해서 현재 코드 구조와 제약을 반영한 실행 가능한 계획을 작성해: %s. "+
			"출력은 반드시 제공된 output schema를 따르는 JSON 객체 하나만 반환해. "+
			"JSON 바깥의 설명, 코드 블록, 마크다운 fence는 절대 출력하지 마. "+
			"`markdown` 필드에는 한국어 마크다운 계획 문서를 넣어. "+
			"이 문서는 목표, 범위, 구현 방향, 작업 단계, 검증 방법이 드러나야 한다. "+
			"`items` 필드에는 계획을 진행하기 전에 사용자가 답해야 하는 결정 사항만 넣어. "+
			"각 item은 `question`과 `expected_answers`를 반드시 포함해야 하며, "+
			"`question`은 바로 답할 수 있게 구체적으로 쓰고 `expected_answers`는 짧은 선택지 후보 배열로 작성해. "+
			"추가로 결정할 사항이 없으면 `items`는 빈 배열로 반환해.",
		strings.TrimSpace(title),
		quotedDirectories(directories),
	)
}

func quotedDirectories(directories []string) string {
	quoted := make([]string, 0, len(directories))

	for _, dir := range directories {
		quoted = append(quoted, fmt.Sprintf("%q", dir))
	}

	return strings.Join(quoted, ", ")
}

func writeSchemaFile() (string, error) {
	file, err := os.CreateTemp("", "decider-schema-*.json")
	if err != nil {
		return "", fmt.Errorf("create schema temp file: %w", err)
	}

	path := file.Name()
	_, err = file.Write(decisionOutputSchema)
	closeErr := file.Close()

	if err != nil {
		_ = os.Remove(path)

		return "", fmt.Errorf("write schema temp file: %w", err)
	}

	if closeErr != nil {
		_ = os.Remove(path)

		return "", fmt.Errorf("close schema temp file: %w", closeErr)
	}

	return path, nil
}

func readDecision(content []byte) (*deciderpkg.Decision, error) {
	var decision deciderpkg.Decision

	err := json.Unmarshal(content, &decision)
	if err != nil {
		return nil, fmt.Errorf("unmarshal decision file: %w", err)
	}

	if strings.TrimSpace(decision.Markdown) == "" {
		return nil, errDecisionMarkdownEmpty
	}

	for _, item := range decision.Items {
		if strings.TrimSpace(item.Question) == "" {
			return nil, errDecisionItemQuestionEmpty
		}
	}

	return &decision, nil
}
