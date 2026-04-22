package ai

import "errors"

var (
	ErrEmptyText          = errors.New("text cannot be empty")
	ErrServiceUnavailable = errors.New("ai service is not available")
)

type ExplanationRequest struct {
	Text string `json:"text"`
}

type ExplanationResponse struct {
	OriginalText string   `json:"originalText"`
	Explanation  string   `json:"explanation"`
	Tone         string   `json:"tone"`
	Examples     []string `json:"examples"`
}
