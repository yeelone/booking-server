package resolvers

import (
	"booking"
	"booking/models"
	"booking/pkg/constvar"
	"context"
	"fmt"
)

type authorResolver struct{ *Resolver }

func (r *authorResolver) ID(ctx context.Context, obj *models.Author) (int, error) {
	return int(obj.ID), nil
}

func (r *authorResolver) Name(ctx context.Context, obj *models.Author) (string, error) {
	return obj.Name, nil
}
func (r *authorResolver) CreatedAt(ctx context.Context, obj *models.Author) (string, error) {
	return fmt.Sprintf("yyyy-mm-dd HH:mm:ss : ", obj.CreatedAt.Format("2006-01-02 15:04:05")), nil
}

func (r *authorResolver) UpdatedAt(ctx context.Context, obj *models.Author) (string, error) {
	return fmt.Sprintf(obj.UpdatedAt.Format("2006-01-02 15:04:05")), nil
}

func (r *authorResolver) DeletedAt(ctx context.Context, obj *models.Author) (string, error) {
	if obj.DeletedAt != nil {
		return fmt.Sprintf(obj.DeletedAt.Format("2006-01-02 15:04:05")), nil
	}
	return "", nil
}

func (r *authorResolver) Books(ctx context.Context, obj *models.Author, filter *booking.BookFilterInput, pagination *booking.Pagination) (booking.QueryBookResponse, error) {
	if pagination == nil {
		pagination = &booking.Pagination{
			Skip: 0,
			Take: constvar.DefaultLimit,
		}
	}
	books, total, err := models.GetAuthorRelatedBooks(obj.ID, pagination.Skip, pagination.Take)
	resp := booking.QueryBookResponse{Rows:books,TotalCount:&total}
	return resp, err
}

func (r *mutationResolver) CreateAuthor(ctx context.Context, input booking.NewAuthor) (models.Author, error) {
	m := models.Author{
		Name:     input.Name,
	}
	return m.Create()
}

func (r *mutationResolver) UpdateAuthor(ctx context.Context, input booking.UpdateAuthorInput) (models.Author, error) {
	m := models.Author{}
	data := make(map[string]interface{})
	if input.Name != nil {
		m.Name = *input.Name
		data["name"] = m.Name
	}
	m.ID = uint64(input.ID)

	return m,m.Update(data)
}

func (r *mutationResolver) DeleteAuthor(ctx context.Context, input booking.DeleteIDInput) (bool, error) {
	if len(input.Ids) > 0 {
		id := uint64(input.Ids[0])
		if err := models.DeleteAuthor(id); err != nil {
			return false, err
		}
	}
	return true, nil
}


func (r *mutationResolver) CreateAuthorAndBookRelationship(ctx context.Context, input booking.AuthorAndBookRelationshipInput) (bool, error) {
	ids := []uint64{}
	if input.BookIds != nil {
		for _, id := range input.BookIds{
			ids = append(ids, uint64(id))
		}
	}

	// Insert the group to the database.
	if err := models.AddAuthorBooks(uint64(input.AuthorID),ids);err != nil {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) RemoveAuthorAndBookRelationship(ctx context.Context, input booking.AuthorAndBookRelationshipInput) (bool, error) {
	ids := []uint64{}
	if input.BookIds != nil {
		for _, id := range input.BookIds{
			ids = append(ids, uint64(id))
		}
	}

	// Insert the group to the database.
	if err := models.RemoveBookFromAuthor(uint64(input.AuthorID),ids);err != nil {
		return false, err
	}
	return true, nil
}
