package rpc

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/validate"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"

	yabav1 "yaba/gen/yaba/v1"
	"yaba/gen/yaba/v1/yabav1connect"
	"yaba/internal/ctxutil"
	"yaba/internal/database"
	"yaba/internal/model"
)

var (
	errUnauthenticated = errors.New("unauthenticated")
	errUnknownCardType = errors.New("unknown card type")
	errNotFound        = errors.New("payment method not found")
)

// NewHandler serves the Connect services under /rpc/.
func NewHandler(pool *pgxpool.Pool) http.Handler {
	interceptors := connect.WithInterceptors(connect.UnaryInterceptorFunc(requireUser), validate.NewInterceptor())
	mux := http.NewServeMux()
	mux.Handle(yabav1connect.NewPaymentMethodServiceHandler(&PaymentMethodServer{Pool: pool}, interceptors))
	mux.Handle(yabav1connect.NewRewardCardServiceHandler(&RewardCardServer{Pool: pool}, interceptors))

	return http.StripPrefix("/rpc", mux)
}

// requireUser rejects calls without the user that auth.SessionInterceptor puts in the context.
func requireUser(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		if ctxutil.GetUser(ctx) == uuid.Nil {
			return nil, connect.NewError(connect.CodeUnauthenticated, errUnauthenticated)
		}

		return next(ctx, req)
	}
}

// PaymentMethodServer implements PaymentMethodService for the user in the request context.
type PaymentMethodServer struct {
	Pool *pgxpool.Pool
}

func (s *PaymentMethodServer) ListPaymentMethods(
	ctx context.Context,
	_ *connect.Request[yabav1.ListPaymentMethodsRequest],
) (*connect.Response[yabav1.ListPaymentMethodsResponse], error) {
	methods, err := database.ListPaymentMethods(ctx, s.Pool)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	out := make([]*yabav1.PaymentMethod, 0, len(methods))
	for _, method := range methods {
		out = append(out, toProto(method))
	}

	return connect.NewResponse(&yabav1.ListPaymentMethodsResponse{PaymentMethods: out}), nil
}

func (s *PaymentMethodServer) CreatePaymentMethod(
	ctx context.Context,
	req *connect.Request[yabav1.CreatePaymentMethodRequest],
) (*connect.Response[yabav1.CreatePaymentMethodResponse], error) {
	method, err := s.fromProto(ctx, uuid.New(), req.Msg)
	if err != nil {
		return nil, err
	}

	if err = database.CreatePaymentMethod(ctx, s.Pool, method); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&yabav1.CreatePaymentMethodResponse{PaymentMethod: toProto(method)}), nil
}

func (s *PaymentMethodServer) UpdatePaymentMethod(
	ctx context.Context,
	req *connect.Request[yabav1.UpdatePaymentMethodRequest],
) (*connect.Response[yabav1.UpdatePaymentMethodResponse], error) {
	id := uuid.MustParse(req.Msg.GetId()) // validated by protovalidate

	// The DAO's update doesn't report missing rows, so check ownership first.
	if _, err := database.GetPaymentMethod(ctx, s.Pool, id); pgxscan.NotFound(err) {
		return nil, connect.NewError(connect.CodeNotFound, errNotFound)
	} else if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	method, err := s.fromProto(ctx, id, req.Msg)
	if err != nil {
		return nil, err
	}

	if err = database.UpdatePaymentMethod(ctx, s.Pool, method); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&yabav1.UpdatePaymentMethodResponse{PaymentMethod: toProto(method)}), nil
}

func (s *PaymentMethodServer) DeletePaymentMethod(
	ctx context.Context,
	req *connect.Request[yabav1.DeletePaymentMethodRequest],
) (*connect.Response[yabav1.DeletePaymentMethodResponse], error) {
	id := uuid.MustParse(req.Msg.GetId()) // validated by protovalidate

	deleted, err := database.DeletePaymentMethod(ctx, s.Pool, id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if !deleted {
		return nil, connect.NewError(connect.CodeNotFound, errNotFound)
	}

	return connect.NewResponse(&yabav1.DeletePaymentMethodResponse{}), nil
}

// paymentMethodFields is implemented by both the create and update requests.
type paymentMethodFields interface {
	GetCardTypeId() string
	GetDisplayName() string
	GetAcquiredDate() *timestamppb.Timestamp
	GetCancelByDate() *timestamppb.Timestamp
}

// fromProto builds a payment method; the display name defaults to the reward card's name.
func (s *PaymentMethodServer) fromProto(
	ctx context.Context,
	id uuid.UUID,
	msg paymentMethodFields,
) (*model.PaymentMethod, error) {
	cardType := uuid.MustParse(msg.GetCardTypeId()) // validated by protovalidate

	card, err := database.GetRewardCard(ctx, s.Pool, cardType)
	if pgxscan.NotFound(err) {
		return nil, connect.NewError(connect.CodeInvalidArgument, errUnknownCardType)
	} else if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// protovalidate rejects an explicitly empty name, so "" means unset.
	name := msg.GetDisplayName()
	if name == "" {
		name = card.Name
	}

	return &model.PaymentMethod{
		ID:           id,
		DisplayName:  name,
		CardType:     cardType,
		AcquiredDate: fromTimestamp(msg.GetAcquiredDate()),
		CancelByDate: fromTimestamp(msg.GetCancelByDate()),
	}, nil
}

func toProto(pm *model.PaymentMethod) *yabav1.PaymentMethod {
	out := &yabav1.PaymentMethod{
		Id:           pm.ID.String(),
		DisplayName:  pm.DisplayName,
		AcquiredDate: toTimestamp(pm.AcquiredDate),
		CancelByDate: toTimestamp(pm.CancelByDate),
	}

	// Older rows have no card type.
	if pm.CardType != uuid.Nil {
		out.CardTypeId = pm.CardType.String()
	}

	return out
}

func fromTimestamp(ts *timestamppb.Timestamp) sql.NullTime {
	if ts == nil {
		return sql.NullTime{}
	}

	return sql.NullTime{Time: ts.AsTime(), Valid: true}
}

func toTimestamp(t sql.NullTime) *timestamppb.Timestamp {
	if !t.Valid {
		return nil
	}

	return timestamppb.New(t.Time)
}
