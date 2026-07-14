package service

import (
	"context"
	"errors"
	"strings"

	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/dto"
	"github.com/rayeemomayeer/SpotSync/internal/models"
	"github.com/rayeemomayeer/SpotSync/internal/repository"
	"gorm.io/gorm"
)

type PaymentService struct {
	payments     *repository.PaymentRepository
	reservations *repository.ReservationRepository
}

func NewPaymentService(payments *repository.PaymentRepository, reservations *repository.ReservationRepository) *PaymentService {
	return &PaymentService{payments: payments, reservations: reservations}
}

func (s *PaymentService) Record(ctx context.Context, userID uint, req dto.CreatePaymentRequest) (*models.Payment, error) {
	res, err := s.reservations.FindByID(ctx, req.ReservationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	if res.UserID != userID {
		return nil, domain.ErrNotOwner
	}

	piID := strings.TrimSpace(req.StripePaymentIntentID)
	if existing, err := s.payments.GetByPaymentIntent(ctx, piID); err == nil && existing != nil {
		return existing, nil
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	currency := strings.ToLower(strings.TrimSpace(req.Currency))
	if currency == "" {
		currency = "usd"
	}

	resID := req.ReservationID
	payment := &models.Payment{
		ReservationID:         &resID,
		UserID:                userID,
		ZoneID:                res.ZoneID,
		StripePaymentIntentID: piID,
		AmountCents:           req.AmountCents,
		Currency:              currency,
		Status:                models.PaymentStatusSucceeded,
	}
	if err := s.payments.Create(ctx, payment); err != nil {
		return nil, err
	}
	return payment, nil
}

func (s *PaymentService) Get(ctx context.Context, userID, paymentID uint, role string) (*models.Payment, error) {
	p, err := s.payments.GetByID(ctx, paymentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	if role != models.RoleAdmin && role != models.RoleSaaSAdmin && p.UserID != userID {
		return nil, domain.ErrForbidden
	}
	return p, nil
}

func (s *PaymentService) ListMine(ctx context.Context, userID uint) ([]models.Payment, error) {
	return s.payments.ListByUser(ctx, userID)
}

func (s *PaymentService) RecordRefund(ctx context.Context, userID, paymentID uint, role string, req dto.CreateRefundRequest) (*models.Refund, error) {
	p, err := s.Get(ctx, userID, paymentID, role)
	if err != nil {
		return nil, err
	}
	if p.Status == models.PaymentStatusRefunded {
		return nil, domain.ErrConflict
	}

	ref := &models.Refund{
		PaymentID:      p.ID,
		StripeRefundID: strPtr(strings.TrimSpace(req.StripeRefundID)),
		AmountCents:    req.AmountCents,
		Status:         models.RefundStatusSucceeded,
	}
	if err := s.payments.CreateRefund(ctx, ref); err != nil {
		return nil, err
	}
	if err := s.payments.UpdateStatus(ctx, p.ID, models.PaymentStatusRefunded); err != nil {
		return nil, err
	}
	return ref, nil
}

func strPtr(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func ToPaymentResponse(p *models.Payment) dto.PaymentResponse {
	return dto.PaymentResponse{
		ID:                    p.ID,
		ReservationID:         p.ReservationID,
		UserID:                p.UserID,
		ZoneID:                p.ZoneID,
		StripePaymentIntentID: p.StripePaymentIntentID,
		AmountCents:           p.AmountCents,
		Currency:              p.Currency,
		Status:                p.Status,
		CreatedAt:             p.CreatedAt,
		UpdatedAt:             p.UpdatedAt,
	}
}

func ToRefundResponse(r *models.Refund) dto.RefundResponse {
	return dto.RefundResponse{
		ID:             r.ID,
		PaymentID:      r.PaymentID,
		StripeRefundID: r.StripeRefundID,
		AmountCents:    r.AmountCents,
		Status:         r.Status,
		CreatedAt:      r.CreatedAt,
	}
}
