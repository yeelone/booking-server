package resolvers

import (
	"booking"
	"booking/models"
	"booking/pkg/constvar"
	"context"
	"fmt"
)


type groupResolver struct{ *Resolver }

func (r *groupResolver) ID(ctx context.Context, obj *models.Group) (string, error) {
	return fmt.Sprintf("%d",obj.ID), nil
}
func (r *groupResolver) Parent(ctx context.Context, obj *models.Group) (int, error){
	return int(obj.Parent), nil
}
func (r *groupResolver) CreatedAt(ctx context.Context, obj *models.Group) (string, error){
	return fmt.Sprintf(obj.CreatedAt.Format("2006-01-02 15:04:05")), nil
}

func (r *groupResolver) UpdatedAt(ctx context.Context, obj *models.Group) (string, error){
	return fmt.Sprintf(obj.UpdatedAt.Format("2006-01-02 15:04:05")), nil
}

func (r *groupResolver) DeletedAt(ctx context.Context, obj *models.Group) (string, error){
	if obj.DeletedAt != nil {
		return fmt.Sprintf(obj.DeletedAt.Format("2006-01-02 15:04:05")), nil
	}
	return "",nil
}

func (r *groupResolver)  Users(ctx context.Context, group *models.Group,filter *booking.UserFilterInput,pagination *booking.Pagination) (booking.QueryUserResponse, error){
	if pagination == nil {
		pagination = &booking.Pagination{
			Skip: 0,
			Take: constvar.DefaultLimit,
		}
	}
	users, total, err := models.GetGroupRelatedUsers(group.ID, pagination.Skip, pagination.Take)
	resp := booking.QueryUserResponse{Rows:users,TotalCount:&total}
	return resp, err
}

func (r *mutationResolver) CreateGroup(ctx context.Context, input booking.NewGroup) (models.Group, error) {
	m := models.Group{
		Name:        input.Name,
		Parent:      uint64(input.Parent),
		Levels:      input.Levels,
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
	if input.Name != nil {
		g.Name = *input.Name
	}
	if input.Parent != nil {
		g.Parent = uint64(*input.Parent)
	}
	if input.Levels != nil {
		g.Levels = *input.Levels
	}

	g.ID = uint64(input.ID)

	// Insert the group to the database.
	err := g.Update()
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
	return true,nil
}


func (r *mutationResolver) CreateUserAndGroupRelationship(ctx context.Context, input booking.UserAndGroupRelationshipInput) (bool, error) {
	uids := []uint64{}
	if input.UserIds != nil {
		for _, id := range input.UserIds{
			uids = append(uids, uint64(id))
		}

	}

	// Insert the group to the database.
	if err := models.AddGroupUsers(uint64(input.GroupID),uids);err != nil {
		fmt.Println(err)
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) RemoveUserAndGroupRelationship(ctx context.Context, input booking.UserAndGroupRelationshipInput) (bool, error) {
	uids := []uint64{}
	if input.UserIds != nil {
		for _, id := range input.UserIds{
			uids = append(uids, uint64(id))
		}
	}

	// Insert the group to the database.
	if err := models.RemoveUsersFromGroup(uint64(input.GroupID),uids);err != nil {
		return false, err
	}
	return true, nil
}

