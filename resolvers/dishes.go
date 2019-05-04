package resolvers

import (
	"booking"
	"booking/models"
	"context"
	"fmt"
)

type dishesResolver struct{ *Resolver }

func (r *dishesResolver) ID(ctx context.Context, obj *models.Dishes) (string, error) {
	return fmt.Sprintf("%d", obj.ID), nil
}
func (r *dishesResolver) CreatedAt(ctx context.Context, obj *models.Dishes) (string, error) {
	return fmt.Sprintf(obj.CreatedAt.Format("2006-01-02 15:04:05")), nil
}

func (r *dishesResolver) UpdatedAt(ctx context.Context, obj *models.Dishes) (string, error) {
	return fmt.Sprintf(obj.UpdatedAt.Format("2006-01-02 15:04:05")), nil
}

func (r *dishesResolver) DeletedAt(ctx context.Context, obj *models.Dishes) (string, error) {
	if obj.DeletedAt != nil {
		return fmt.Sprintf(obj.DeletedAt.Format("2006-01-02 15:04:05")), nil
	}
	return "", nil
}

func (r *mutationResolver) CreateDishes(ctx context.Context, input booking.NewDishes) (models.Dishes, error) {
	m := models.Dishes{
		Name:    input.Name,
		Picture: input.Picture,
	}

	// Insert the Dishes to the database.
	g, err := m.Create()
	if err != nil {
		return g, err
	}
	return g, nil
}

func (r *mutationResolver) UpdateDishes(ctx context.Context, input booking.UpdateDishesInput) (models.Dishes, error) {
	g := models.Dishes{}

	data := make(map[string]interface{})

	if input.Name != nil {
		g.Name = *input.Name
		data["name"] = *input.Name
	}
	if input.Picture != nil {
		g.Picture = *input.Picture
		data["picture"] = *input.Picture
	}

	g.ID = uint64(input.ID)

	// Insert the Dishes to the database.
	err := g.Update(data)
	if err != nil {
		return g, err
	}
	return g, nil
}

func (r *mutationResolver) DeleteDishes(ctx context.Context, input booking.DeleteIDInput) (bool, error) {
	if len(input.Ids) > 0 {
		id := uint64(input.Ids[0])
		if err := models.DeleteDish(id); err != nil {
			fmt.Println("err", err)
			return false, err
		}
	}
	return true, nil
}
