package graph

import (
	"booking/models"
	"context"
	"fmt"
)

type messageResolver struct{ *Resolver }

func (r *messageResolver) ID(ctx context.Context, obj *models.Message) (string, error) {
	return fmt.Sprintf("%d", obj.ID), nil
}

func (r *messageResolver) CreatedBy(ctx context.Context, obj *models.Message) (models.User, error) {
	return obj.CreatedBy, nil
}

func (r *messageResolver) CreatedAt(ctx context.Context, obj *models.Message) (string, error) {
	return fmt.Sprintf(obj.CreatedAt.Format("2006-01-02 15:04:05")), nil
}
