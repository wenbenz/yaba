package rpc

import (
	"database/sql"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	yabav1 "yaba/gen/yaba/v1"
	"yaba/internal/model"
)

// PaymentMethodToProto converts a domain payment method into its protobuf form.
// A nil input yields nil. A missing card type maps to an empty card_type_id and
// NULL dates map to unset timestamps.
func PaymentMethodToProto(pm *model.PaymentMethod) *yabav1.PaymentMethod {
	if pm == nil {
		return nil
	}

	cardTypeID := ""
	if pm.CardType != uuid.Nil {
		cardTypeID = pm.CardType.String()
	}

	return &yabav1.PaymentMethod{
		Id:           pm.ID.String(),
		DisplayName:  pm.DisplayName,
		CardTypeId:   cardTypeID,
		AcquiredDate: nullTimeToProto(pm.AcquiredDate),
		CancelByDate: nullTimeToProto(pm.CancelByDate),
	}
}

// nullTimeFromProto maps an unset timestamp to NULL and a set one to its UTC time.
func nullTimeFromProto(ts *timestamppb.Timestamp) sql.NullTime {
	if ts == nil {
		return sql.NullTime{}
	}

	return sql.NullTime{Time: ts.AsTime().UTC(), Valid: true}
}

// nullTimeToProto maps NULL to an unset timestamp.
func nullTimeToProto(t sql.NullTime) *timestamppb.Timestamp {
	if !t.Valid {
		return nil
	}

	return timestamppb.New(t.Time)
}
