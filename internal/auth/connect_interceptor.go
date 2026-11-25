package auth

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"yaba/internal/ctxutil"
)

// ErrUnauthenticated is returned to Connect callers whose request has no session user.
var ErrUnauthenticated = errors.New("unauthenticated")

// ConnectAuthInterceptor rejects Connect RPCs that have no user in the context
// (set by SessionInterceptor) with connect.CodeUnauthenticated.
type ConnectAuthInterceptor struct{}

var _ connect.Interceptor = (*ConnectAuthInterceptor)(nil)

// NewConnectAuthInterceptor returns an interceptor that requires an authenticated user.
func NewConnectAuthInterceptor() *ConnectAuthInterceptor {
	return &ConnectAuthInterceptor{}
}

// WrapUnary rejects unary handler calls that have no authenticated user.
func (*ConnectAuthInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		if req.Spec().IsClient {
			return next(ctx, req)
		}

		if err := requireUser(ctx); err != nil {
			return nil, err
		}

		return next(ctx, req)
	}
}

// WrapStreamingClient passes client streams through unchanged.
func (*ConnectAuthInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

// WrapStreamingHandler rejects streaming handler calls that have no authenticated user.
func (*ConnectAuthInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		if err := requireUser(ctx); err != nil {
			return err
		}

		return next(ctx, conn)
	}
}

func requireUser(ctx context.Context) error {
	if ctxutil.GetUser(ctx) == uuid.Nil {
		return connect.NewError(connect.CodeUnauthenticated, ErrUnauthenticated)
	}

	return nil
}
