package lsp

import "encoding/json"

// PulledConfiguration is implemented by a Flavor whose server pulls its settings with
// workspace/configuration and only reads them again when told they changed. rust-analyzer is one:
// a loose .rs file is served only once it is listed in "detachedFiles", so each new loose file must
// reach it (M3 R3). Other servers keep the plain push of Configuration after initialized.
type PulledConfiguration interface {
	PullsConfiguration() bool
}

func pullsConfiguration(flavor Flavor) bool {
	pulled, ok := flavor.(PulledConfiguration)
	return ok && pulled.PullsConfiguration()
}

// answer replies to a request of the server: workspace/configuration gets the flavor's settings
// once per item asked; anything else gets null.
func (s *Server) answer(method string, params json.RawMessage) any {
	if method != "workspace/configuration" || !pullsConfiguration(s.flavor) {
		return nil
	}
	var request struct {
		Items []json.RawMessage `json:"items"`
	}
	_ = json.Unmarshal(params, &request)
	settings := s.flavor.Configuration()
	answers := make([]any, len(request.Items))
	for i := range answers {
		answers[i] = settings
	}
	return answers
}

// refreshConfigurationLocked tells a pulling server that its settings changed (a new loose file),
// so it asks for them again. Nothing is sent when they did not change.
func (s *Server) refreshConfigurationLocked() {
	if !pullsConfiguration(s.flavor) {
		return
	}
	current, err := json.Marshal(s.flavor.Configuration())
	if err != nil || string(current) == s.settingsSent {
		return
	}
	s.settingsSent = string(current)
	s.notifyLocked("workspace/didChangeConfiguration", map[string]any{"settings": nil})
}
