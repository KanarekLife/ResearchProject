package openaicompatible

import (
	"errors"
	"net/http"
	"strings"

	"mcbench/integrations/inference"
)

// rejected turns a provider's refusal of max_tokens or of reasoning_content
// on messages into errTokenParam or errReasoningField; other errors pass.
func rejected(err error, msgs []inference.Message, useMaxCompletion bool) error {
	var he *inference.HTTPError
	if !errors.As(err, &he) {
		return err
	}
	if he.Status == http.StatusBadRequest && !useMaxCompletion && strings.Contains(he.Body, fieldMaxCompletionTokens) {
		return errTokenParam
	}
	if he.Status == http.StatusUnprocessableEntity && hasReasoning(msgs) {
		return errReasoningField
	}
	return err
}
