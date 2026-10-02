package delve

import (
	"context"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/google/go-dap"
)

const argumentsScope = "Arguments"

// scopeReferences returns the references of the Arguments and Locals scopes (0 when
// a scope is missing).
func scopeReferences(scopes []dap.Scope) (arguments, locals int) {
	for _, scope := range scopes {
		switch {
		case strings.HasPrefix(scope.Name, argumentsScope):
			arguments = scope.VariablesReference
		case strings.HasPrefix(scope.Name, localsScope):
			locals = scope.VariablesReference
		}
	}
	return arguments, locals
}

// frameVariables reads the arguments and locals of any frame of the paused stack.
// It answers empty lists when the program is running or the request fails.
func (s *session) frameVariables(frameID int) domain.FrameVariables {
	empty := domain.FrameVariables{Arguments: []domain.Variable{}, Locals: []domain.Variable{}}
	ctx := s.pausedContext()
	if ctx == nil {
		return empty
	}
	request := &dap.ScopesRequest{Request: newRequest("scopes"), Arguments: dap.ScopesArguments{FrameId: frameID}}
	response, err := s.call(ctx, request)
	if err != nil {
		return empty
	}
	arguments, locals := scopeReferences(response.(*dap.ScopesResponse).Body.Scopes)
	return domain.FrameVariables{
		Arguments: s.childrenOrEmpty(ctx, arguments),
		Locals:    s.childrenOrEmpty(ctx, locals),
	}
}

func (s *session) childrenOrEmpty(ctx context.Context, reference int) []domain.Variable {
	if reference == 0 {
		return []domain.Variable{}
	}
	return s.children(ctx, reference)
}
