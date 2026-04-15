package plannercli

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/neatflowcv/worker/internal/app/planner"
	"github.com/neatflowcv/worker/internal/pkg/decider"
	"github.com/stretchr/testify/require"
)

func TestModelUpdateInputBackspaceRemovesLastRune(t *testing.T) {
	t.Parallel()

	// Arrange
	currentModel := model{
		rootDir: "",
		service: nil,
		output:  "",
		git:     "",
		title:   "",
		phase:   phaseInput,
		response: &planner.CreatePlanResponse{
			Title: "",
			Git:   "",
			Items: []decider.Item{
				{
					Question:        "무엇을 먼저 할까?",
					ExpectedAnswers: []string{"Git"},
				},
			},
			Markdown: "",
		},
		selected:    0,
		customInput: "한글a",
		err:         nil,
	}

	// Act
	updatedModel, command := currentModel.updateInput(tea.KeyMsg{
		Type:  tea.KeyBackspace,
		Runes: nil,
		Alt:   false,
		Paste: false,
	})
	resultModel, ok := updatedModel.(model)

	// Assert
	require.True(t, ok)
	require.Nil(t, command)
	require.Equal(t, "한글", resultModel.customInput)
}

func TestModelUpdateInputAppendsSpace(t *testing.T) {
	t.Parallel()

	// Arrange
	currentModel := model{
		rootDir: "",
		service: nil,
		output:  "",
		git:     "",
		title:   "",
		phase:   phaseInput,
		response: &planner.CreatePlanResponse{
			Title: "",
			Git:   "",
			Items: []decider.Item{
				{
					Question:        "무엇을 먼저 할까?",
					ExpectedAnswers: []string{"Git"},
				},
			},
			Markdown: "",
		},
		selected:    0,
		customInput: "직접",
		err:         nil,
	}

	// Act
	updatedModel, command := currentModel.updateInput(tea.KeyMsg{
		Type:  tea.KeySpace,
		Runes: []rune{' '},
		Alt:   false,
		Paste: false,
	})
	resultModel, ok := updatedModel.(model)

	// Assert
	require.True(t, ok)
	require.Nil(t, command)
	require.Equal(t, "직접 ", resultModel.customInput)
}

func TestTrimLastRuneRemovesSingleKoreanSyllable(t *testing.T) {
	t.Parallel()

	// Arrange
	value := "한글"

	// Act
	trimmed := trimLastRune(value)

	// Assert
	require.Equal(t, "한", trimmed)
}
