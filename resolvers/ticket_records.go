package resolvers

import (
	"booking/models"
	"context"
	"fmt"
)

type ticketRecordResolver struct{ *Resolver }

func (r *ticketRecordResolver) ID(ctx context.Context, obj *models.TicketRecord) (string, error) {
	return fmt.Sprintf("%d",obj.ID), nil
}

func (r *ticketRecordResolver) CreatedAt(ctx context.Context, obj *models.TicketRecord) (string, error){
	return fmt.Sprintf(obj.CreatedAt.Format("2006-01-02 15:04:05")), nil
}

func (r *ticketRecordResolver) UpdatedAt(ctx context.Context, obj *models.TicketRecord) (string, error){
	return fmt.Sprintf(obj.UpdatedAt.Format("2006-01-02 15:04:05")), nil
}

func (r *ticketRecordResolver) DeletedAt(ctx context.Context, obj *models.TicketRecord) (string, error){
	if obj.DeletedAt != nil {
		return fmt.Sprintf(obj.DeletedAt.Format("2006-01-02 15:04:05")), nil
	}

	return "", nil
}

func (r *ticketRecordResolver) Operator(ctx context.Context, obj *models.TicketRecord) (int, error){
	return int(obj.Operator),nil
}

func (r *ticketRecordResolver) Owner(ctx context.Context, obj *models.TicketRecord) (int, error){
	return int(obj.Owner),nil
}