package rpc

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/validate"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"

	yabav1 "yaba/gen/yaba/v1"
	"yaba/gen/yaba/v1/yabav1connect"
	"yaba/internal/auth"
	"yaba/internal/database"
	"yaba/internal/model"
)

var (
	errUnknownCardType       = errors.New("unknown card type")
	errPaymentMethodNotFound = errors.New("payment method not found")
)

// PaymentMethodServer implements yabav1connect.PaymentMethodServiceHandler on top of
// the payment method DAO. All calls are scoped to the user in the request context.
type PaymentMethodServer struct {
	Pool *pgxpool.Pool
}

var _ yabav1connect.PaymentMethodServiceHandler = (*PaymentMethodServer)(nil)

// paymentMethodInput holds the fields shared by create and update requests.
type paymentMethodInput struct {
	cardTypeID     string
	displayName    string
	hasDisplayName bool
	acquiredDate   *timestamppb.Timestamp
	cancelByDate   *timestamppb.Timestamp
}

// NewHandler builds the PaymentMethodService Connect handler. Session auth runs before
// protovalidate, so unauthenticated calls get CodeUnauthenticated even if invalid.
// It returns the path to mount the handler on (without any prefix) and the handler.
func NewHandler(pool *pgxpool.Pool, opts ...connect.HandlerOption) (string, http.Handler) {
	return yabav1connect.NewPaymentMethodServiceHandler(
		&PaymentMethodServer{Pool: pool},
		connect.WithInterceptors(auth.NewConnectAuthInterceptor(), validate.NewInterceptor()),
		connect.WithHandlerOptions(opts...),
	)
}

// ListPaymentMethods returns the caller's payment methods ordered by display name.
func (s *PaymentMethodServer) ListPaymentMethods(
	ctx context.Context,
	_ *connect.Request[yabav1.ListPaymentMethodsRequest],
) (*connect.Response[yabav1.ListPaymentMethodsResponse], error) {
	methods, err := database.ListPaymentMethods(ctx, s.Pool)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	out := make([]*yabav1.PaymentMethod, len(methods))
	for i, method := range methods {
		out[i] = PaymentMethodToProto(method)
	}

	return connect.NewResponse(&yabav1.ListPaymentMethodsResponse{PaymentMethods: out}), nil
}

// CreatePaymentMethod creates a payment method for the caller. The display name
// defaults to the reward card's name when not supplied.
func (s *PaymentMethodServer) CreatePaymentMethod(
	ctx context.Context,
	req *connect.Request[yabav1.CreatePaymentMethodRequest],
) (*connect.Response[yabav1.CreatePaymentMethodResponse], error) {
	method, err := s.paymentMethodFromInput(ctx, createInput(req.Msg))
	if err != nil {
		return nil, err
	}

	method.ID = uuid.New()

	if err = database.CreatePaymentMethod(ctx, s.Pool, method); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	created, err := s.getPaymentMethod(ctx, method.ID)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&yabav1.CreatePaymentMethodResponse{
		PaymentMethod: PaymentMethodToProto(created),
	}), nil
}

// UpdatePaymentMethod fully replaces one of the caller's payment methods. It returns
// CodeNotFound if the id does not exist or belongs to another user.
func (s *PaymentMethodServer) UpdatePaymentMethod(
	ctx context.Context,
	req *connect.Request[yabav1.UpdatePaymentMethodRequest],
) (*connect.Response[yabav1.UpdatePaymentMethodResponse], error) {
	id, err := parsePaymentMethodID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}

	if _, err = s.getPaymentMethod(ctx, id); err != nil {
		return nil, err
	}

	method, err := s.paymentMethodFromInput(ctx, updateInput(req.Msg))
	if err != nil {
		return nil, err
	}

	method.ID = id

	if err = database.UpdatePaymentMethod(ctx, s.Pool, method); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	updated, err := s.getPaymentMethod(ctx, id)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&yabav1.UpdatePaymentMethodResponse{
		PaymentMethod: PaymentMethodToProto(updated),
	}), nil
}

// DeletePaymentMethod deletes one of the caller's payment methods. It returns
// CodeNotFound if nothing was deleted.
func (s *PaymentMethodServer) DeletePaymentMethod(
	ctx context.Context,
	req *connect.Request[yabav1.DeletePaymentMethodRequest],
) (*connect.Response[yabav1.DeletePaymentMethodResponse], error) {
	id, err := parsePaymentMethodID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}

	deleted, err := database.DeletePaymentMethod(ctx, s.Pool, id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if !deleted {
		return nil, connect.NewError(connect.CodeNotFound, errPaymentMethodNotFound)
	}

	return connect.NewResponse(&yabav1.DeletePaymentMethodResponse{}), nil
}

// getPaymentMethod reads one of the caller's payment methods, mapping a missing row
// to CodeNotFound.
func (s *PaymentMethodServer) getPaymentMethod(ctx context.Context, id uuid.UUID) (*model.PaymentMethod, error) {
	method, err := database.GetPaymentMethod(ctx, s.Pool, id)
	if err != nil {
		if pgxscan.NotFound(err) {
			return nil, connect.NewError(connect.CodeNotFound, errPaymentMethodNotFound)
		}

		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return method, nil
}

// paymentMethodFromInput checks that the card type is a known reward card and builds
// a payment method without an ID. The display name defaults to the reward card's name.
func (s *PaymentMethodServer) paymentMethodFromInput(
	ctx context.Context,
	in paymentMethodInput,
) (*model.PaymentMethod, error) {
	cardType, err := uuid.Parse(in.cardTypeID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("%w: %w", errUnknownCardType, err))
	}

	rewardCard, err := database.GetRewardCard(ctx, s.Pool, cardType)
	if err != nil {
		if pgxscan.NotFound(err) {
			return nil, connect.NewError(connect.CodeInvalidArgument, errUnknownCardType)
		}

		return nil, connect.NewError(connect.CodeInternal, err)
	}

	displayName := rewardCard.Name
	if in.hasDisplayName {
		displayName = in.displayName
	}

	return &model.PaymentMethod{
		DisplayName:  displayName,
		CardType:     cardType,
		AcquiredDate: nullTimeFromProto(in.acquiredDate),
		CancelByDate: nullTimeFromProto(in.cancelByDate),
	}, nil
}

// createInput extracts the shared payment method fields from a create request.
func createInput(msg *yabav1.CreatePaymentMethodRequest) paymentMethodInput {
	return paymentMethodInput{
		cardTypeID:     msg.GetCardTypeId(),
		displayName:    msg.GetDisplayName(),
		hasDisplayName: msg.DisplayName != nil,
		acquiredDate:   msg.GetAcquiredDate(),
		cancelByDate:   msg.GetCancelByDate(),
	}
}

// updateInput extracts the shared payment method fields from an update request.
func updateInput(msg *yabav1.UpdatePaymentMethodRequest) paymentMethodInput {
	return paymentMethodInput{
		cardTypeID:     msg.GetCardTypeId(),
		displayName:    msg.GetDisplayName(),
		hasDisplayName: msg.DisplayName != nil,
		acquiredDate:   msg.GetAcquiredDate(),
		cancelByDate:   msg.GetCancelByDate(),
	}
}

// parsePaymentMethodID parses a payment method id, mapping a malformed id to
// CodeInvalidArgument.
func parsePaymentMethodID(raw string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	return id, nil
}
