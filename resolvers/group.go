package resolvers

import (
	"booking"
	"booking/models"
	"booking/pkg/constvar"
	"booking/util"
	"context"
	"fmt"
)

type groupResolver struct{ *Resolver }

func (r *groupResolver) ID(ctx context.Context, obj *models.Group) (string, error) {
	return fmt.Sprintf("%d", obj.ID), nil
}
func (r *groupResolver) Parent(ctx context.Context, obj *models.Group) (int, error) {
	return int(obj.Parent), nil
}
func (r *groupResolver) AdminID(ctx context.Context, obj *models.Group) (int, error) {
	fmt.Println(util.PrettyJson(obj))
	return int(obj.AdminID), nil
}
func (r *groupResolver) AdminInfo(ctx context.Context, obj *models.Group) (user models.User, err error) {
	user, err = models.GetUserByID(obj.AdminID)
	fmt.Println(obj.AdminID)
	if err != nil {
		return models.User{}, nil
	}
	return user, nil
}

func (r *groupResolver) CreatedAt(ctx context.Context, obj *models.Group) (string, error) {
	return fmt.Sprintf(obj.CreatedAt.Format("2006-01-02 15:04:05")), nil
}

func (r *groupResolver) UpdatedAt(ctx context.Context, obj *models.Group) (string, error) {
	return fmt.Sprintf(obj.UpdatedAt.Format("2006-01-02 15:04:05")), nil
}

func (r *groupResolver) DeletedAt(ctx context.Context, obj *models.Group) (string, error) {
	if obj.DeletedAt != nil {
		return fmt.Sprintf(obj.DeletedAt.Format("2006-01-02 15:04:05")), nil
	}
	return "", nil
}

func (r *groupResolver) Users(ctx context.Context, group *models.Group, filter *booking.UserFilterInput, pagination *booking.Pagination) (booking.QueryUserResponse, error) {
	if pagination == nil {
		pagination = &booking.Pagination{
			Skip: 0,
			Take: constvar.DefaultLimit,
		}
	}

	where := ""
	whereValue := ""
	if filter != nil {
		if filter.Username != nil && *filter.Username != "" {
			where = "username"
			whereValue = *filter.Username
		}

		if filter.Email != nil && *filter.Email != "" {
			where = "email"
			whereValue = *filter.Email
		}
	}

	users, total, err := models.GetGroupRelatedUsers(group.ID, where, whereValue, pagination.Skip, pagination.Take)
	fmt.Println("error", err)
	resp := booking.QueryUserResponse{Rows: users, TotalCount: &total}
	return resp, err
}

func (r *groupResolver) Canteens(ctx context.Context, obj *models.Group, filter *booking.CanteenFilterInput, pagination *booking.Pagination) (booking.QueryCanteenResponse, error) {
	canteens, err := models.GetGroupRelatedCanteens(obj.ID)
	total := len(canteens)
	resp := booking.QueryCanteenResponse{Rows: canteens, TotalCount: &total}
	return resp, err

}

func (r *mutationResolver) CreateGroup(ctx context.Context, input booking.NewGroup) (models.Group, error) {
	fmt.Println("input.Picture", input.Picture )


	m := models.Group{
		Name:    input.Name,
		AdminID: uint64(input.Admin),
		Picture: input.Picture,
		Parent:  uint64(input.Parent),
	}

	// Insert the group to the database.
	g, err := m.Create()
	if err != nil {
		return g, err
	}
	return g, nil
}

func (r *mutationResolver) UpdateGroup(ctx context.Context, input booking.UpdateGroupInput) (models.Group, error) {
	g := models.Group{}
	data := make(map[string]interface{})

	if input.Name != nil {
		g.Name = *input.Name
		data["name"] = *input.Name
	}

	if input.Picture != nil {
		g.Picture = *input.Picture
		if len(*input.Picture) == 0 {
			data["picture"] = `/assets/canteen-min.jpg`
		}else{
			data["picture"] = *input.Picture
		}
	}

	if input.Parent != nil {
		g.Parent = uint64(*input.Parent)
		data["parent"] = uint64(*input.Parent)
	}
	if input.Levels != nil {
		g.Levels = *input.Levels
		data["levels"] = *input.Levels
	}

	if input.Admin != nil {
		data["admin_id"] = *input.Admin
	}

	g.ID = uint64(input.ID)

	// Insert the group to the database.
	err := g.Update(data)
	if err != nil {
		return g, err
	}
	return g, nil
}

func (r *mutationResolver) DeleteGroup(ctx context.Context, input booking.DeleteIDInput) (bool, error) {
	if len(input.Ids) > 0 {
		id := uint64(input.Ids[0])
		if err := models.DeleteGroup(id); err != nil {
			fmt.Println("err", err)
			return false, err
		}
	}
	return true, nil
}

func (r *mutationResolver) CreateUserAndGroupRelationship(ctx context.Context, input booking.UserAndGroupRelationshipInput) (bool, error) {
	uids := []uint64{}
	if input.UserIds != nil {
		for _, id := range input.UserIds {
			uids = append(uids, uint64(id))
		}

	}

	// Insert the group to the database.
	if err := models.AddGroupUsers(uint64(input.GroupID), uids); err != nil {
		fmt.Println(err)
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) RemoveUserAndGroupRelationship(ctx context.Context, input booking.UserAndGroupRelationshipInput) (bool, error) {
	uids := []uint64{}
	if input.UserIds != nil {
		for _, id := range input.UserIds {
			uids = append(uids, uint64(id))
		}
	}

	// Insert the group to the database.
	if err := models.RemoveUsersFromGroup(uint64(input.GroupID), uids); err != nil {
		return false, err
	}
	return true, nil
}
