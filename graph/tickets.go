package graph

import (
	"booking/graph/model"
	"booking/models"
	"context"
	"fmt"
)

type ticketResolver struct{ *Resolver }

func (r *ticketResolver) ID(ctx context.Context, obj *models.Ticket) (string, error) {
	return fmt.Sprintf("%d", obj.ID), nil
}

func (r *ticketResolver) CreatedAt(ctx context.Context, obj *models.Ticket) (string, error) {
	return fmt.Sprintf(obj.CreatedAt.Format("2006-01-02 15:04:05")), nil
}

func (r *ticketResolver) UpdatedAt(ctx context.Context, obj *models.Ticket) (string, error) {
	return fmt.Sprintf(obj.UpdatedAt.Format("2006-01-02 15:04:05")), nil
}

func (r *ticketResolver) UUID(ctx context.Context, obj *models.Ticket) (string, error) {
	return obj.UUID, nil
}

func (r *ticketResolver) UserID(ctx context.Context, obj *models.Ticket) (int, error) {
	return int(obj.UserID), nil
}

func (r *mutationResolver) GenerateTickets(ctx context.Context, input model.GenarateTicketInput) (*model.QueryTicketResponse, error) {
	// Insert the tickets to the database.
	resp := &model.QueryTicketResponse{}

	record := models.TicketRecord{}
	record.Operator = ctx.Value("user_id").(uint64)
	record.Owner = uint64(input.UserID)
	record.Number = input.Number
	record.Action = "sell"
	record.Description = ""

	rd, err := record.Create()

	if err != nil {
		return resp, err
	}

	tickets, err := models.BatchCreateTickets(uint64(input.UserID), input.Number, input.Type, input.Price)
	if err != nil {
		models.DeleteTicketRecord(rd.ID)
		return resp, err
	}

	resp = &model.QueryTicketResponse{
		TotalCount: &input.Number,
		Rows:       tickets,
		Take:       &input.Number,
	}

	return resp, nil
}

func (r *mutationResolver) TransferTickets(ctx context.Context, input model.TransferTicketInput) (*model.TransferResponse, error) {
	// Insert the tickets to the database.
	resp := &model.TransferResponse{}

	successCount, errorCount, err := models.TransferTicket(uint64(input.FromUserID), input.Type, uint64(input.ToUserID), input.Number)
	if err != nil {
		return resp, err
	}

	resp = &model.TransferResponse{
		SuccessCount: successCount,
		ErrorCount:   errorCount,
	}

	return resp, nil
}

func (r *mutationResolver) RecyclingTickets(ctx context.Context, input model.RecyclingTicketsInput) (bool, error) {
	record := models.TicketRecord{}
	record.Operator = ctx.Value("user_id").(uint64)
	record.Owner = uint64(input.UserID)
	record.Number = input.Number
	record.Action = "recycling"
	record.Description = ""

	rd, err := record.Create()

	if err != nil {
		return false, err
	}

	err = models.RecyclingTickets(uint64(input.UserID), input.Type, input.Number)
	if err != nil {
		models.DeleteTicketRecord(rd.ID)
		return false, err
	}

	return true, nil
}
