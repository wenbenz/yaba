package handlers

import (
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"os"
	"strings"
	"yaba/errors"
	"yaba/internal/ctxutil"
)

type SingleUserModeInterceptor struct {
	UserID      uuid.UUID
	Intercepted http.Handler
}

func (interceptor SingleUserModeInterceptor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := ctxutil.WithUser(r.Context(), interceptor.UserID)
	interceptor.Intercepted.ServeHTTP(w, r.WithContext(ctx))
}

func InterceptSingleUserMode(handler http.Handler) (SingleUserModeInterceptor, error) {
	enabled := os.Getenv("SINGLE_USER_MODE")
	if strings.ToLower(enabled) != "true" {
		return SingleUserModeInterceptor{}, errors.InvalidStateError{
			Message: "Single user mode interceptor called, but SINGLE_USER_MODE is not enabled!",
		}
	}

	id := os.Getenv("SINGLE_USER_UUID")
	userID, err := uuid.Parse(id)

	if err != nil {
		return SingleUserModeInterceptor{},
			fmt.Errorf("could not parse UUID from SINGLE_USER_UUID in single user mode: %w", err)
	}

	return SingleUserModeInterceptor{
		UserID:      userID,
		Intercepted: handler,
	}, nil
}
