package resolvers

import (
	"booking"
	"booking/models"
	"booking/util"
	"context"
	"fmt"
	"github.com/rs/xid"
	"github.com/skip2/go-qrcode"
	"github.com/spf13/viper"
	"time"
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
func (r *userResolver) CreatedAt(ctx context.Context, obj *models.User) (string, error) {
	return fmt.Sprintf(obj.CreatedAt.Format("2006-01-02 15:04:05")), nil
}

func (r *userResolver) UpdatedAt(ctx context.Context, obj *models.User) (string, error) {
	return fmt.Sprintf(obj.UpdatedAt.Format("2006-01-02 15:04:05")), nil
}

func (r *userResolver) DeletedAt(ctx context.Context, obj *models.User) (string, error) {
	if obj.DeletedAt != nil {
		return fmt.Sprintf(obj.DeletedAt.Format("2006-01-02 15:04:05")), nil
	}
	return "", nil
}
func (r *userResolver) Groups(ctx context.Context, obj *models.User, pagination *booking.Pagination) (booking.QueryGroupResponse, error) {
	skip := 0
	take := 0
	if pagination != nil {
		skip = pagination.Skip
		take = pagination.Take
	}

	u, total, err := models.GetGroupsByUser(obj.ID)
	resp := booking.QueryGroupResponse{
		TotalCount: &total,
		Skip:       &skip,
		Take:       &take,
		Rows:       u.Groups,
	}
	return resp, err
}
func (r *userResolver) Roles(ctx context.Context, obj *models.User, pagination *booking.Pagination) (booking.QueryRoleResponse, error) {
	skip := 0
	take := 0
	if pagination != nil {
		skip = pagination.Skip
		take = pagination.Take
	}

	u, total, err := models.GetRolesByUser(obj.ID)
	resp := booking.QueryRoleResponse{
		TotalCount: &total,
		Skip:       &skip,
		Take:       &take,
		Rows:       u.Roles,
	}
	return resp, err
}

func (r *userResolver) Tickets(ctx context.Context, obj *models.User, pagination *booking.Pagination, filter *booking.TicketFilterInput) (booking.QueryTicketResponse, error) {
	skip := 0
	take := 0
	if pagination != nil {
		skip = pagination.Skip
		take = pagination.Take
	}

	var err error
	count := booking.Count{}
	if *filter.Count && filter != nil { // 只计算余票数量
		count.Breakfast, count.Lunch, count.Dinner, err = models.CountTicketsDetailByUser(obj.ID)
		resp := booking.QueryTicketResponse{
			Count: &count,
		}

		return resp, err
	}

	u, total, err := models.GetTicketsByUser(obj.ID)
	resp := booking.QueryTicketResponse{
		TotalCount: &total,
		Skip:       &skip,
		Take:       &take,
		Rows:       u.Tickets,
	}
	return resp, err
}

func (r *mutationResolver) CreateUser(ctx context.Context, input booking.NewUser) (user models.User, err error) {
	u := models.User{
		Email:    input.Email,
		IsSuper:  false,
		Password: input.Password,
		Username: input.Username,
	}

	// Validate the data.
	if err = u.Validate(); err != nil {
		fmt.Println("user validate error", err)
		return user, err
	}

	// Encrypt the user password.
	if err := u.Encrypt(); err != nil {
		fmt.Println("user Encrypt error", err)
		return user, err
	}

	// Insert the user to the database.
	if err := u.Create(); err != nil {
		fmt.Println("user Create error", err)
		return user, err
	}

	if err := createUserQrCode(u); err != nil {
		fmt.Println("user Create qr code error", err)
		return user, err
	}


	//如果 存在组ID，则将用户加入到该组里
	if input.GroupID != nil {
		if *input.GroupID != 0 {
			uids := []uint64{}
			uids = append(uids, u.ID)
			err = models.AddGroupUsers(uint64(*input.GroupID), uids)

			if err != nil {
				//如果不成功，还需要将该用户删除掉
				models.DeleteUser(u.ID)
			}
		}
	}

	//查看账号是否存在
	normal := viper.GetString("role.normal")
	roles, _, err := models.GetRoles("name", normal, 0, 1)

	if len(roles) > 0 {
		uids := []uint64{}
		uids = append(uids, u.ID)
		models.AddRoleUsers(roles[0].ID, uids)
	}

	return u, nil
}

func createUserQrCode(user models.User) error {
	path := "/download/qrcode/" + user.Username + "qrcode_" + time.Now().String() + ".png"

	data := make(map[string]interface{})
	data["qrcode"] = path
	data["qrcode_uuid"] = xid.New().String()

	str := "module:profile;id:" + util.Uint2Str(user.ID) + ";username:" + user.Username + ";date:" + time.Now().String() + ";qrcode_uuid:" + data["qrcode_uuid"].(string) + ";"
	err := qrcode.WriteFile(str, qrcode.Medium, 256, "."+path)

	if err != nil {
		return err
	}

	if err = user.Update(data); err != nil {
		return err
	}

	return nil
}

func (r *mutationResolver) UpdateUser(ctx context.Context, input booking.UpdateUserInput) (user models.User, err error) {
	u := models.User{}
	u.ID = uint64(input.ID)
	data := make(map[string]interface{})

	if input.Password != nil {
		//want to change password
		if err := models.ChangeUsersPassword(u.ID, *input.Password); err != nil {
			fmt.Println("err", err)
			return u, err
		}
		return u, nil
	}

	if input.ReGenQrcode != nil && *input.ReGenQrcode == true {
		//want to change password
		if err := createUserQrCode(u); err != nil {
			return u, err
		}
	}

	if input.Username != nil {
		data["username"] = *input.Username
	}
	if input.Email != nil {
		data["email"] = *input.Email
	}
	if input.Nickname != nil {
		data["nickname"] = *input.Nickname
	}
	if input.IDCard != nil {
		data["id_card"] = *input.IDCard
	}
	if input.IsSuper != nil {
		data["is_super"] = *input.IsSuper
	}
	if input.Picture != nil {
		data["picture"] = *input.Picture
	}
	if input.State != nil {
		data["state"] = *input.State
	}

	// Insert the user to the database.
	if err := u.Update(data); err != nil {
		fmt.Println("err", err)
		return u, err
	}

	return u, nil
}

func (r *mutationResolver) ResetPassword(ctx context.Context, input booking.ResetPasword) (newPwd string, err error) {
	uids := make([]uint64, len(input.Ids))

	for i, id := range input.Ids {
		uids[i] = uint64(id)
	}

	if err := models.ResetUsersPassword(uids); err != nil {
		return "", err
	}

	password := viper.GetString("default_password")

	return password, nil
}

func (r *mutationResolver) DeleteUser(ctx context.Context, input booking.DeleteIDInput) (result bool, err error) {
	if len(input.Ids) > 0 {
		id := uint64(input.Ids[0])
		if err := models.DeleteUser(id); err != nil {
			fmt.Println("err", err)
			return false, err
		}
	}

	return true, nil
}
