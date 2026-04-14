package decider

type DecideRequest struct {
	Title       string
	Directories []string
}

//nolint:tagliatelle // JSON field names are fixed by the codex output schema.
type Item struct {
	Question        string   `json:"question"`
	ExpectedAnswers []string `json:"expected_answers"`
}

type Decision struct {
	Markdown string `json:"markdown"`
	Items    []Item `json:"items"`
}

type Decider interface {
	Decide(request DecideRequest) (*Decision, error)
}
