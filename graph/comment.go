package graph

import (
	"booking/graph/model"
	"booking/models"
	"context"
	"errors"
	"fmt"
)

type commentResolver struct{ *Resolver }

func (r *commentResolver) ID(ctx context.Context, obj *models.Comment) (string, error) {
	return fmt.Sprintf("%d", obj.ID), nil
}

func (r *commentResolver) CreatedAt(ctx context.Context, obj *models.Comment) (string, error) {
	return fmt.Sprintf(obj.CreatedAt.Format("2006-01-02 15:04:05")), nil
}

func (r *commentResolver) UpdatedAt(ctx context.Context, obj *models.Comment) (string, error) {
	return fmt.Sprintf(obj.UpdatedAt.Format("2006-01-02 15:04:05")), nil
}

func (r *commentResolver) DeletedAt(ctx context.Context, obj *models.Comment) (string, error) {
	if obj.DeletedAt != nil {
		return fmt.Sprintf(obj.DeletedAt.Format("2006-01-02 15:04:05")), nil
	}

	return "", nil
}
func (r *commentResolver) User(ctx context.Context, obj *models.Comment) (user *models.User, err error) {
	if obj.UserID != 0 {
		u , err := models.GetUserByID(obj.UserID)
		return &u ,err
	}

	return user, nil
}
func (r *mutationResolver) CreateComment(ctx context.Context, input model.NewComment) (comment *models.Comment, err error) {
	c := &models.Comment{
		UserID:                     uint64(input.UserID),
		Body:    input.Body,
	}

	if _, err := c.Create(); err != nil {
		return c, err
	}


	if _, ok := r.commentTunnels[input.Tunnel]; !ok {
		return comment, errors.New("频道未上线，请稍候再试")
	}

	tunnel := r.commentTunnels[input.Tunnel]

	for _, observer := range tunnel.Observers {
		observer <- c
	}

	return c, nil
}