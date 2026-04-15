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
	errDecisionDirectoryEmpty    = errors.New("decision directory is required")
	errDecisionItemQuestionEmpty = errors.New("decision item question is empty")
	errRefineAnswerEmpty         = errors.New("refine answer is required")
	errRefineMarkdownEmpty       = errors.New("refine markdown is required")
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
	directory := strings.TrimSpace(request.Directory)
	if directory == "" {
		return nil, errDecisionDirectoryEmpty
	}

	output, err := d.runner.Run(
		context.Background(),
		directory,
		"--output-schema",
		d.schemaPath,
		buildPrompt(request.Title),
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

func (d *Decider) RefinePlan(request deciderpkg.RefinePlanRequest) (*deciderpkg.Decision, error) {
	directory := strings.TrimSpace(request.Directory)
	if directory == "" {
		return nil, errDecisionDirectoryEmpty
	}

	markdown := strings.TrimSpace(request.Markdown)
	if markdown == "" {
		return nil, errRefineMarkdownEmpty
	}

	if len(request.Answers) == 0 {
		return nil, errRefineAnswerEmpty
	}

	for _, answer := range request.Answers {
		if strings.TrimSpace(answer.Answer) == "" {
			return nil, errRefineAnswerEmpty
		}
	}

	output, err := d.runner.Run(
		context.Background(),
		directory,
		"--output-schema",
		d.schemaPath,
		buildRefinePrompt(markdown, request.Answers),
	)
	if err != nil {
		return nil, fmt.Errorf("execute codex refine plan: %w", err)
	}

	decision, err := readDecision(output)
	if err != nil {
		return nil, err
	}

	return decision, nil
}

func buildPrompt(title string) string {
	return fmt.Sprintf(
		"%q 작업에 대한 계획을 세워. "+
			"출력은 반드시 제공된 output schema를 따르는 JSON 객체 하나만 반환해. "+
			"JSON 바깥의 설명, 코드 블록, 마크다운 fence는 절대 출력하지 마. "+
			"`markdown` 필드에는 한국어 마크다운 계획 문서를 넣어. "+
			"이 문서는 목표, 범위, 구현 방향, 작업 단계, 검증 방법이 드러나야 한다. "+
			"`items` 필드에는 계획을 진행하기 전에 사용자가 답해야 하는 결정 사항만 넣어. "+
			"각 item은 `question`과 `expected_answers`를 반드시 포함해야 하며, "+
			"`question`은 바로 답할 수 있게 구체적으로 쓰고 `expected_answers`는 짧은 선택지 후보 배열로 작성해. "+
			"추가로 결정할 사항이 없으면 `items`는 빈 배열로 반환해.",
		strings.TrimSpace(title),
	)
}

func buildRefinePrompt(markdown string, answers []deciderpkg.QuestionAnswer) string {
	instructions := make([]string, 0, len(answers))
	for _, answer := range answers {
		question := strings.TrimSpace(answer.Question)

		resolvedAnswer := strings.TrimSpace(answer.Answer)
		if question == "" {
			instructions = append(
				instructions,
				fmt.Sprintf(
					"사용자가 기존 질문을 무시하고 %q 라는 추가 지시를 남겼다. 이 지시를 반영해 계획을 수정해.",
					resolvedAnswer,
				),
			)

			continue
		}

		instructions = append(
			instructions,
			fmt.Sprintf(
				"사용자가 답한 질문은 %q 이고, 답변은 %q 이다. 이 질문과 답변을 반영해 계획을 수정해.",
				question,
				resolvedAnswer,
			),
		)
	}

	return fmt.Sprintf(
		"다음은 현재 계획 문서다.\n\n%s\n\n"+
			"%s "+
			"출력은 반드시 제공된 output schema를 따르는 JSON 객체 하나만 반환해. "+
			"JSON 바깥의 설명, 코드 블록, 마크다운 fence는 절대 출력하지 마. "+
			"`markdown` 필드에는 수정된 한국어 마크다운 계획 문서를 넣어. "+
			"수정된 문서는 기존 계획의 구조와 의도를 유지하되, 새 답변을 반영해서 목표, 범위, 구현 방향, 작업 단계, 검증 방법이 더 명확해야 한다. "+
			"`items` 필드에는 아직 사용자가 답해야 하는 결정 사항만 남겨. "+
			"각 item은 `question`과 `expected_answers`를 반드시 포함해야 하며, "+
			"`question`은 바로 답할 수 있게 구체적으로 쓰고 `expected_answers`는 짧은 선택지 후보 배열로 작성해. "+
			"더 물어볼 사항이 없으면 `items`는 빈 배열로 반환해.",
		markdown,
		strings.Join(instructions, " "),
	)
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
