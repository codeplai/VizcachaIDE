package dap

import (
	"context"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/google/go-dap"
)

// frameVariables reads the arguments and locals of any frame of the paused stack.
// It answers empty lists when the program is running or the request fails.
func (s *Session) frameVariables(frameID int) domain.FrameVariables {
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
	arguments, locals := scopeReferences(response.(*dap.ScopesResponse).Body.Scopes, s.flavor)
	return domain.FrameVariables{
		Arguments: s.childrenOrEmpty(ctx, arguments),
		Locals:    s.childrenOrEmpty(ctx, locals),
	}
}

func (s *Session) childrenOrEmpty(ctx context.Context, reference int) []domain.Variable {
	if reference == 0 {
		return []domain.Variable{}
	}
	return s.children(ctx, reference)
}
