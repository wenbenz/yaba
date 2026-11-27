package rpc

import (
	"context"
	"database/sql"
	"errors"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	yabaerrors "yaba/errors"
	yabav1 "yaba/gen/yaba/v1"
	"yaba/internal/ctxutil"
	"yaba/internal/database"
	"yaba/internal/model"
	"yaba/internal/user"
)

var errEmailTaken = errors.New("email already in use")

// UserServer implements UserService for the user in the request context.
type UserServer struct {
	Pool *pgxpool.Pool
}

func (s *UserServer) GetProfile(
	ctx context.Context,
	_ *connect.Request[yabav1.GetProfileRequest],
) (*connect.Response[yabav1.GetProfileResponse], error) {
	u, err := database.GetUserByID(ctx, s.Pool, ctxutil.GetUser(ctx))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&yabav1.GetProfileResponse{Profile: toProfile(u)}), nil
}

func (s *UserServer) UpdateProfile(
	ctx context.Context,
	req *connect.Request[yabav1.UpdateProfileRequest],
) (*connect.Response[yabav1.UpdateProfileResponse], error) {
	u, err := database.GetUserByID(ctx, s.Pool, ctxutil.GetUser(ctx))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	u.Email = sql.NullString{String: req.Msg.GetEmail(), Valid: true}

	if err = database.UpdateUser(ctx, s.Pool, u); err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
			return nil, connect.NewError(connect.CodeAlreadyExists, errEmailTaken)
		}

		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&yabav1.UpdateProfileResponse{Profile: toProfile(u)}), nil
}

func (s *UserServer) ChangePassword(
	ctx context.Context,
	req *connect.Request[yabav1.ChangePasswordRequest],
) (*connect.Response[yabav1.ChangePasswordResponse], error) {
	msg := req.Msg

	err := user.ChangePassword(ctx, s.Pool, ctxutil.GetUser(ctx), msg.GetCurrentPassword(), msg.GetNewPassword())
	if _, ok := errors.AsType[yabaerrors.InvalidInputError](err); ok {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	} else if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&yabav1.ChangePasswordResponse{}), nil
}

func toProfile(u *model.User) *yabav1.Profile {
	return &yabav1.Profile{
		Id:       u.ID.String(),
		Username: u.Username,
		Email:    u.Email.String,
	}
}
