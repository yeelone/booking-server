package models

import (
	"booking/pkg/auth"
	"booking/util"
	"errors"

	"github.com/spf13/viper"
	validator "gopkg.in/go-playground/validator.v9"
)

const (
	USERSTATEFREEZE = 1 //冻结状态
	USERSTATEACTIVE = 0 //激活状态
)

// User : User represents a registered user.
type User struct {
	BaseModel
	Email    string  `json:"email" gorm:"column:email;"`
	Username string  `json:"username" gorm:"column:username;not null" binding:"required" validate:"min=1,max=32"`
	Nickname string  `json:"nickname" gorm:"column:nichname;not null" `
	IDCard   string  `json:"id_card"`
	Password string  `json:"password" gorm:"column:password;not null" `
	IsSuper  bool    `json:"is_super"`
	Picture  string  `json:"picture"`
	State    int     `json:"state"`
	Groups   []Group `json:"groups" gorm:"many2many:user_groups;"`
	Roles    []Role  `json:"roles" gorm:"many2many:user_roles"`
	Books    []Book
}

// TableName :
func (u *User) TableName() string {
	return "users"
}

// Create creates a new user account.
func (u *User) Create() error {
	return DB.Self.Create(&u).Error
}

// DeleteUser deletes the user by the user identifier.
func DeleteUser(id uint64) error {
	user := User{}
	user.BaseModel.ID = id
	return DB.Self.Delete(&user).Error
}

// Update updates an user account information.
func (u *User) Update(data map[string]interface{}) (err error) {
	tx := DB.Self.Begin()
	if err := tx.Model(&u).Update(data).Error; err != nil {
		tx.Rollback()
		return errors.New("无法更新")
	}
	tx.Commit()

	return err
}

// Save :create or update account information
func (u User) Save() (User, error) {
	tx := DB.Self.Begin()

	if u.IsSuper {
		u.IsSuper = true
	} else {
		u.IsSuper = false
	}
	if u.ID > 0 {
		tx.Model(&u).Where("id = ?", u.ID).Updates(u)
	} else if len(u.Email) > 0 {
		u.Password, _ = auth.Encrypt(u.Password)

		if err := tx.Create(&u).Error; err != nil {
			tx.Rollback()
			return u, err
		}
	}

	tx.Commit()

	return u, nil
}

// GetUsers
func GetUsers(where string, value string, skip, take int) (users []User, total int, err error) {
	u := &User{}
	w := ""
	if len(where) > 0 && len(value) > 0 {

		w = where + " LIKE ?"
		v := "%" + value + "%"

		d := DB.Self.Debug().Where(w, v).Order("id").Offset(skip).Limit(take).Find(&users)

		if err := DB.Self.Model(u).Where(w, v).Count(&total).Error; err != nil {
			return users, 0, errors.New("cannot fetch count of the row")
		}
		return users, total, d.Error
	}

	d := DB.Self.Debug().Order("id").Offset(skip).Limit(take).Find(&users)
	if err := DB.Self.Model(u).Count(&total).Error; err != nil {
		return users, 0, errors.New("cannot fetch count of the row")
	}
	return users, total, d.Error

}

// GetUserGroups get groups which user belong it
func GetGroupsByUser(uid uint64) (User, int, error) {
	u := User{}
	u.ID = uid
	d := DB.Self.Preload("Groups").First(&u)
	if d.Error != nil {
		return u, 0, d.Error
	}

	total, err := CountGroupsByUser(uid)
	return u, total, err
}

// GetUserGroups get groups which user belong it
func GetRolesByUser(uid uint64) (User, int, error) {
	u := User{}
	u.ID = uid
	d := DB.Self.Preload("Roles").First(&u)
	if d.Error != nil {
		return u, 0, d.Error
	}

	total, err := CountRolesByUser(uid)
	return u, total, err
}

// CountGroupsByUser
func CountGroupsByUser(uid uint64) (total int, err error) {
	countSql := "SELECT count(user_id) from user_groups where user_id =" + util.Uint2Str(uid)
	rows, _ := DB.Self.Raw(countSql).Rows()
	for rows.Next() {
		err = rows.Scan(&total)
	}
	return total, nil
}

// CountRolesByUser
func CountRolesByUser(uid uint64) (total int, err error) {
	countSql := "SELECT count(user_id) from user_roles where user_id =" + util.Uint2Str(uid)
	rows, _ := DB.Self.Raw(countSql).Rows()
	for rows.Next() {
		err = rows.Scan(&total)
	}
	return total, nil
}

func (u *User) Compare(pwd string) (err error) {
	err = auth.Compare(u.Password, pwd)
	return
}

// Encrypt the user password.
func (u *User) Encrypt() (err error) {
	u.Password, err = auth.Encrypt(u.Password)
	return
}

// GetUser gets an user by the user identifier.
func GetUserByID(id uint64) (User, error) {
	u := User{}
	d := DB.Self.Where("id = ?", id).First(&u)

	return u, d.Error
}

// GetUser gets an user by the user identifier.
func GetUserByName(username string) (*User, error) {
	u := &User{}
	d := DB.Self.Where("username = ?", username).Preload("Roles").First(&u)

	return u, d.Error
}

func ResetUsersPassword(uids []uint64) (err error) {
	password, err := auth.Encrypt(viper.GetString("default_password"))
	tx := DB.Self.Begin()
	if err := tx.Model(&User{}).Where("id in (?)", uids).Update(map[string]interface{}{"password": password}).Error; err != nil {
		tx.Rollback()
		return errors.New("无法重置密码")
	}
	tx.Commit()

	return err
}

func ChangeUsersPassword(id uint64, password string) (err error) {
	newPassword, err := auth.Encrypt(password)
	tx := DB.Self.Begin()
	if err := tx.Model(&User{}).Where("id = ?", id).Update(map[string]interface{}{"password": newPassword}).Error; err != nil {
		tx.Rollback()
		return errors.New("无法修改密码")
	}
	tx.Commit()

	return err
}

// Validate the fields.
func (u *User) Validate() error {
	validate := validator.New()
	return validate.Struct(u)
}
