package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"yaba/internal/constants"
	"yaba/internal/errors"

	"github.com/google/uuid"
)

type SingleUserModeInterceptor struct {
	UserID      uuid.UUID
	Intercepted http.Handler
}

func (interceptor SingleUserModeInterceptor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), constants.CTXUser, interceptor.UserID)
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
