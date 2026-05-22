package service

import (
	"github.com/nusaroute/pkg/logger"
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nusaroute/pkg/events"
	"github.com/nusaroute/pkg/kafka"
	"github.com/nusaroute/services/resolution-service/internal/model"
	"github.com/nusaroute/services/resolution-service/internal/repository"
)

type ResolutionService interface {
	AutoCreateDeliveryFailedTicket(ctx context.Context, orderID, awb, reason string, attemptNum, maxAttempts int) error
	AutoCreateLostTicketAndClaim(ctx context.Context, orderID, awb string, hoursSinceUpdate int) error
	AutoCreateDamagedTicket(ctx context.Context, orderID, awb, reportedBy, damageDesc string, evidence []string) error

	CreateTicket(ctx context.Context, ticket *model.Ticket) error
	GetTicketByID(ctx context.Context, id string) (*model.Ticket, error)
	ListTickets(ctx context.Context, status string, page, perPage int) ([]model.Ticket, int64, error)
	UpdateTicket(ctx context.Context, id string, req model.UpdateTicketRequest) error

	CreateClaim(ctx context.Context, claim *model.Claim) error
	GetClaimByID(ctx context.Context, id string) (*model.Claim, error)
}

type resolutionService struct {
	repo     repository.ResolutionRepository
	producer *kafka.Producer
}

func NewResolutionService(repo repository.ResolutionRepository, producer *kafka.Producer) ResolutionService {
	return &resolutionService{repo: repo, producer: producer}
}

// Removed publishResolutionCreated as outbox handles this now.

func (s *resolutionService) AutoCreateDeliveryFailedTicket(ctx context.Context, orderID, awb, reason string, attemptNum, maxAttempts int) error {
	if attemptNum >= maxAttempts {
		ticket := &model.Ticket{
			OrderID: orderID, AWB: awb, Type: events.ResolutionTypeDeliveryFailed,
			Description: "Pengiriman gagal setelah " + strconv.Itoa(maxAttempts) + " percobaan. Alasan: " + reason,
		}
		evt := events.ResolutionCreatedEvent{
			BaseEvent: events.BaseEvent{EventID: uuid.New().String(), EventType: events.TopicResolutionCreated, Timestamp: time.Now(), Source: "resolution-service", TraceID: logger.GetTraceID(ctx)},
			TicketID: ticket.ID, OrderID: ticket.OrderID, AWB: ticket.AWB, Type: ticket.Type, Priority: ticket.Priority,
		}
		if err := s.repo.CreateTicket(ctx, ticket, events.TopicResolutionCreated, evt); err != nil {
			return err
		}
	}
	return nil
}

func (s *resolutionService) AutoCreateLostTicketAndClaim(ctx context.Context, orderID, awb string, hoursSinceUpdate int) error {
	ticket := &model.Ticket{
		OrderID: orderID, AWB: awb, Type: events.ResolutionTypeLost,
		Description: "Paket tidak menunjukkan aktivitas selama " + strconv.Itoa(hoursSinceUpdate) + " jam.",
	}
	evt := events.ResolutionCreatedEvent{
		BaseEvent: events.BaseEvent{EventID: uuid.New().String(), EventType: events.TopicResolutionCreated, Timestamp: time.Now(), Source: "resolution-service", TraceID: logger.GetTraceID(ctx)},
		TicketID: ticket.ID, OrderID: ticket.OrderID, AWB: ticket.AWB, Type: ticket.Type, Priority: ticket.Priority,
	}
	if err := s.repo.CreateTicket(ctx, ticket, events.TopicResolutionCreated, evt); err != nil {
		return err
	}
	
	claim := &model.Claim{TicketID: ticket.ID, OrderID: orderID, ClaimType: "INSURANCE", Amount: 0}
	return s.repo.CreateClaim(ctx, claim)
}

func (s *resolutionService) AutoCreateDamagedTicket(ctx context.Context, orderID, awb, reportedBy, damageDesc string, evidence []string) error {
	ticket := &model.Ticket{
		OrderID: orderID, AWB: awb, Type: events.ResolutionTypeDamaged,
		Description: "Kerusakan dilaporkan oleh " + reportedBy + ": " + damageDesc,
		Evidence: evidence,
	}
	evt := events.ResolutionCreatedEvent{
		BaseEvent: events.BaseEvent{EventID: uuid.New().String(), EventType: events.TopicResolutionCreated, Timestamp: time.Now(), Source: "resolution-service", TraceID: logger.GetTraceID(ctx)},
		TicketID: ticket.ID, OrderID: ticket.OrderID, AWB: ticket.AWB, Type: ticket.Type, Priority: ticket.Priority,
	}
	if err := s.repo.CreateTicket(ctx, ticket, events.TopicResolutionCreated, evt); err != nil {
		return err
	}
	return nil
}
func (s *resolutionService) CreateTicket(ctx context.Context, ticket *model.Ticket) error {
	evt := events.ResolutionCreatedEvent{
		BaseEvent: events.BaseEvent{EventID: uuid.New().String(), EventType: events.TopicResolutionCreated, Timestamp: time.Now(), Source: "resolution-service", TraceID: logger.GetTraceID(ctx)},
		TicketID: ticket.ID, OrderID: ticket.OrderID, AWB: ticket.AWB, Type: ticket.Type, Priority: ticket.Priority,
	}
	if err := s.repo.CreateTicket(ctx, ticket, events.TopicResolutionCreated, evt); err != nil {
		return err
	}
	return nil
}

func (s *resolutionService) GetTicketByID(ctx context.Context, id string) (*model.Ticket, error) {
	return s.repo.GetTicketByID(ctx, id)
}

func (s *resolutionService) ListTickets(ctx context.Context, status string, page, perPage int) ([]model.Ticket, int64, error) {
	return s.repo.ListTickets(ctx, status, page, perPage)
}

func (s *resolutionService) UpdateTicket(ctx context.Context, id string, req model.UpdateTicketRequest) error {
	var outboxTopic string
	var outboxPayload interface{}

	if req.Resolution != "" {
		ticket, _ := s.repo.GetTicketByID(ctx, id)
		outboxTopic = events.TopicResolutionResolved
		outboxPayload = events.ResolutionResolvedEvent{
			BaseEvent: events.BaseEvent{EventID: uuid.New().String(), EventType: events.TopicResolutionResolved, Timestamp: time.Now(), Source: "resolution-service", TraceID: logger.GetTraceID(ctx)},
			TicketID: id, OrderID: ticket.OrderID, Resolution: req.Resolution,
		}
	}

	if err := s.repo.UpdateTicket(ctx, id, req, outboxTopic, outboxPayload); err != nil {
		return err
	}
	return nil
}

func (s *resolutionService) CreateClaim(ctx context.Context, claim *model.Claim) error {
	return s.repo.CreateClaim(ctx, claim)
}

func (s *resolutionService) GetClaimByID(ctx context.Context, id string) (*model.Claim, error) {
	return s.repo.GetClaimByID(ctx, id)
}
