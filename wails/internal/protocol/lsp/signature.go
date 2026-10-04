package lsp

import (
	"encoding/json"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

type signatureWire struct {
	Label           string          `json:"label"`
	Documentation   json.RawMessage `json:"documentation"`
	ActiveParameter *int            `json:"activeParameter"`
	Parameters      []struct {
		Label json.RawMessage `json:"label"`
	} `json:"parameters"`
}

// toSignatureHelp maps a signatureHelp result; nil when there is no signature.
func toSignatureHelp(raw json.RawMessage) *domain.SignatureHelp {
	var help struct {
		Signatures      []signatureWire `json:"signatures"`
		ActiveSignature int             `json:"activeSignature"`
		ActiveParameter int             `json:"activeParameter"`
	}
	if json.Unmarshal(raw, &help) != nil || len(help.Signatures) == 0 {
		return nil
	}
	signature := help.Signatures[min(max(help.ActiveSignature, 0), len(help.Signatures)-1)]
	parameters := make([]string, 0, len(signature.Parameters))
	for _, parameter := range signature.Parameters {
		parameters = append(parameters, parameterLabel(signature.Label, parameter.Label))
	}
	active := help.ActiveParameter
	if signature.ActiveParameter != nil {
		active = *signature.ActiveParameter
	}
	return &domain.SignatureHelp{
		Label: signature.Label, Documentation: markupText(signature.Documentation),
		Parameters: parameters, ActiveParameter: active,
	}
}

// parameterLabel is a string, or [start, end) UTF-16 offsets inside the signature label.
func parameterLabel(signature string, raw json.RawMessage) string {
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return text
	}
	var offsets [2]int
	if json.Unmarshal(raw, &offsets) != nil {
		return ""
	}
	start, end := runeIndex(signature, offsets[0]), runeIndex(signature, offsets[1])
	runes := []rune(signature)
	if start > end || end > len(runes) {
		return ""
	}
	return string(runes[start:end])
}
