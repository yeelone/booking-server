package resolvers

import (
	"booking"
	"booking/models"
	"booking/util"
	"context"
	"fmt"
)


type userResolver struct{ *Resolver }

func (r *userResolver) ID(ctx context.Context, obj *models.User) (string, error) {
	return fmt.Sprintf("%d", obj.ID), nil
}

func (r *userResolver) IsSuper(ctx context.Context, obj *models.User) (bool, error) {
	return obj.IsSuper, nil
}

func (r *userResolver) Picture(ctx context.Context, obj *models.User) (string, error) {
	return obj.Picture, nil
}

func (r *userResolver) State(ctx context.Context, obj *models.User) (int, error) {
	return obj.State, nil
}
func (r *userResolver) CreatedAt(ctx context.Context, obj *models.User) (string, error){
	return fmt.Sprintf(obj.CreatedAt.Format("2006-01-02 15:04:05")), nil
}

func (r *userResolver) UpdatedAt(ctx context.Context, obj *models.User) (string, error){
	return fmt.Sprintf(obj.UpdatedAt.Format("2006-01-02 15:04:05")), nil
}

func (r *userResolver) DeletedAt(ctx context.Context, obj *models.User) (string, error){
	if obj.DeletedAt != nil {
		return fmt.Sprintf(obj.DeletedAt.Format("2006-01-02 15:04:05")), nil
	}
	return "",nil
}
func (r *userResolver) Groups(ctx context.Context, obj *models.User,pagination *booking.Pagination) (booking.QueryGroupResponse, error){
	skip := 0
	take := 0
	if pagination != nil {
		skip = pagination.Skip
		take = pagination.Take
	}

	u,total,err := models.GetGroupsByUser(obj.ID)
	resp := booking.QueryGroupResponse{
		TotalCount:&total,
		Skip:&skip,
		Take:&take,
		Rows:u.Groups,
	}
	return resp,err
}
func (r *userResolver) Roles(ctx context.Context, obj *models.User,pagination *booking.Pagination) (booking.QueryRoleResponse, error){
	skip := 0
	take := 0
	if pagination != nil {
		skip = pagination.Skip
		take = pagination.Take
	}

	u,total,err := models.GetRolesByUser(obj.ID)
	resp := booking.QueryRoleResponse{
		TotalCount:&total,
		Skip:&skip,
		Take:&take,
		Rows:u.Roles,
	}
	return resp,err
}
func (r *mutationResolver) CreateUser(ctx context.Context, input booking.NewUser) (user models.User, err error) {
	u := models.User{
		Email:    input.Email,
		IsSuper:  false,
		Password: input.Password,
		Username: input.Username,
	}

	fmt.Println("CreateUser",util.PrettyJson(u))
	// Validate the data.
	if err = u.Validate(); err != nil {
		fmt.Println("user validate error", err )
		return user, err
	}

	// Encrypt the user password.
	if err := u.Encrypt(); err != nil {
		fmt.Println("user Encrypt error", err )
		return user, err
	}
	// Insert the user to the database.
	if err := u.Create(); err != nil {
		fmt.Println("user Create error", err )
		return user,err
	}

	return u, nil
}

func (r *mutationResolver) UpdateUser(ctx context.Context, input booking.UpdateUserInput) (user models.User, err error) {
	fmt.Println(util.PrettyJson(input))
	u := models.User{}
	u.ID = uint64(input.ID)
	data := make(map[string]interface{})

	if input.Password != nil {
		//want to change password
		if err := models.ChangeUsersPassword(u.ID,*input.Password ); err != nil {
			fmt.Println("err", err)
			return u,err
		}
		return u, nil
	}

	if input.Username != nil  {
		data["username"] = *input.Username
	}
	if input.Email != nil  {
		data["email"] = *input.Email
	}
	if input.Nickname != nil  {
		data["nickname"] = *input.Nickname
	}
	if input.IDCard != nil {
		data["id_card"] = *input.IDCard
	}
	if input.IsSuper != nil {
		data["is_super"] = *input.IsSuper
	}
	if input.Picture != nil  {
		data["picture"] = *input.Picture
	}
	if input.State != nil {
		data["state"] = *input.State
	}

	// Insert the user to the database.
	if err := u.Update(data); err != nil {
		fmt.Println("err", err)
		return u,err
	}

	return u, nil
}

func (r *mutationResolver) DeleteUser(ctx context.Context, input booking.DeleteIDInput) (result bool, err error) {
	if len(input.Ids) > 0 {
		id := uint64(input.Ids[0])
		if err := models.DeleteUser(id); err != nil {
			fmt.Println("err", err)
			return false,err
		}
	}


	return true, nil
}
