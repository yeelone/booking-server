package resolvers

import (
	"booking"
	"booking/models"
	"booking/pkg/constvar"
	"context"
	"fmt"
	"strconv"
)

type roleResolver struct{ *Resolver }

func (r *roleResolver) ID(ctx context.Context, obj *models.Role) (string, error) {
	return fmt.Sprintf("%d", obj.ID), nil
}

func (r *roleResolver) Name(ctx context.Context, obj *models.Role) (string, error) {
	return obj.Name, nil
}
func (r *roleResolver) CreatedAt(ctx context.Context, obj *models.Role) (string, error) {
	return fmt.Sprintf("yyyy-mm-dd HH:mm:ss : ", obj.CreatedAt.Format("2006-01-02 15:04:05")), nil
}

func (r *roleResolver) UpdatedAt(ctx context.Context, obj *models.Role) (string, error) {
	return fmt.Sprintf(obj.UpdatedAt.Format("2006-01-02 15:04:05")), nil
}

func (r *roleResolver) DeletedAt(ctx context.Context, obj *models.Role) (string, error) {
	if obj.DeletedAt != nil {
		return fmt.Sprintf(obj.DeletedAt.Format("2006-01-02 15:04:05")), nil
	}
	return "", nil
}
func (r *roleResolver) Users(ctx context.Context, obj *models.Role, filter *booking.UserFilterInput, pagination *booking.Pagination) (booking.QueryUserResponse, error) {
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

		if filter.ID != nil && *filter.ID != 0 {
			where = "id"
			whereValue = strconv.Itoa(*filter.ID)
		}
	}

	users, total, err := models.GetRoleRelatedUsers(obj.ID, where, whereValue, pagination.Skip, pagination.Take)
	resp := booking.QueryUserResponse{Rows: users, TotalCount: &total}
	return resp, err
}

func (r *mutationResolver) CreateRole(ctx context.Context, input booking.NewRole) (models.Role, error) {
	m := models.Role{
		Name: input.Name,
	}
	// Insert the group to the database.
	return m.Create()
}

func (r *mutationResolver) UpdateRole(ctx context.Context, input booking.UpdateRoleInput) (models.Role, error) {
	m := models.Role{}

	if input.Name != nil {
		m.Name = *input.Name
	}
	m.ID = uint64(input.ID)
	// Insert the group to the database.
	return m.Create()
}

func (r *mutationResolver) DeleteRole(ctx context.Context, input booking.DeleteIDInput) (bool, error) {
	if len(input.Ids) > 0 {
		id := uint64(input.Ids[0])
		if err := models.DeleteRole(id); err != nil {
			fmt.Println("err", err)
			return false, err
		}
	}
	return true, nil
}

func (r *mutationResolver) CreateUserAndRoleRelationship(ctx context.Context, input booking.UserAndRoleRelationshipInput) (bool, error) {
	uids := []uint64{}
	if input.UserIds != nil {
		for _, id := range input.UserIds {
			uids = append(uids, uint64(id))
		}
	}

	// Insert the group to the database.
	if err := models.AddRoleUsers(uint64(input.RoleID), uids); err != nil {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) RemoveUserAndRoleRelationship(ctx context.Context, input booking.UserAndRoleRelationshipInput) (bool, error) {
	uids := []uint64{}
	if input.UserIds != nil {
		for _, id := range input.UserIds {
			uids = append(uids, uint64(id))
		}
	}

	// Insert the group to the database.
	if err := models.RemoveUserFromRole(uint64(input.RoleID), uids); err != nil {
		return false, err
	}
	return true, nil
}
