package openaicompatible

import (
	"net/http"
	"strings"

	"mcbench/integrations/inference"
)

func statusError(resp *http.Response, raw []byte, useMaxCompletion bool) error {
	if resp.StatusCode == http.StatusBadRequest && !useMaxCompletion && strings.Contains(string(raw), fieldMaxCompletionTokens) {
		return errTokenParam
	}
	return inference.NewHTTPError(resp, raw)
}
