package models

import (
	"booking/pkg/constvar"
	"booking/util"
	"errors"
	"strconv"
	"strings"
)

type Role struct {
	BaseModel
	Name  string `json:"name" gorm:"name"`
	Users []User `json:"users" gorm:"many2many:user_roles;"`
}

const RoleTableName = "role"

// TableName :
func (r *Role) TableName() string {
	return RoleTableName
}

// Create creates a new Role.
func (r *Role) Create() (Role, error) {
	err := DB.Self.Create(&r).Error
	return *r, err
}

// GetRoles
func GetRoles(where string, value string, skip, take int) (roles []Role, total int, err error) {
	u := &Role{}
	w := ""
	if len(where) > 0 {
		w = where + " = ?"
		d := DB.Self.Debug().Where(w, value).Order("id").Offset(skip).Limit(take).Find(&roles)

		if err := DB.Self.Debug().Model(u).Where(where+" = ?", value).Count(&total).Error; err != nil {
			return roles, 0, errors.New("cannot fetch count of the row")
		}
		return roles, total, d.Error
	}

	d := DB.Self.Debug().Order("id").Offset(skip).Limit(take).Find(&roles)
	if err := DB.Self.Debug().Model(u).Count(&total).Error; err != nil {
		return roles, 0, errors.New("cannot fetch count of the row")
	}
	return roles, total, d.Error

}

// DeleteRole deletes the role by the user identifier.
func DeleteRole(id uint64) error {
	role := Role{}
	role.ID = id
	return DB.Self.Delete(&role).Error
}

// Update updates an user Role information.
func (r *Role) Update() (err error) {
	tx := DB.Self.Begin()
	if err := tx.Model(&r).Update(map[string]interface{}{"name": r.Name}).Error; err != nil {
		tx.Rollback()
		return errors.New("无法更新")
	}
	tx.Commit()

	return err
}

//AddRoleUsers :
func AddRoleUsers(rid uint64, uids []uint64) (err error) {
	r := &Role{}

	if r, err = GetRole(rid, false); err != nil {
		return errors.New("User Role is not existed!")
	}

	tx := DB.Self.Begin()

	var users []User
	//需要先清掉用户原先关联的其它角色，一个用户不能同时属于多个角色
	for _, id := range uids {
		u := User{}
		u.ID = id
		users = append(users, u)
		tx.Model(&u).Association("Roles").Clear()
	}

	err = tx.Model(&r).Association("Users").Append(users).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

//RemoveGroupUsers :
func RemoveUserFromRole(rid uint64, idList []uint64) (err error) {
	r := &Role{}
	if r, err = GetRole(rid, false); err != nil {
		return errors.New("Role is not existed!")
	}

	tx := DB.Self.Begin()

	uids := make([]string, len(idList))

	for i, id := range idList {
		uids[i] = util.Uint2Str(id)
	}
	err = tx.Model(&r).Exec(" delete from user_roles where user_id in (" + strings.Join(uids, ",") + ") and role_id = " + util.Uint2Str(rid) + " ;").Error
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

// GetRoles :
func GetRole(id uint64, withUsers bool) (result *Role, err error) {
	r := &Role{}
	if id == 0 {
		return result, errors.New("cannot find Role by id " + util.Uint2Str(id))
	}
	err = DB.Self.Select("id,name").First(&r, id).Error
	if withUsers {
		DB.Self.Model(&result).Select("id").Association("Users").Find(&r.Users)
	}
	return r, err
}

//ListRoles :
func ListRoles(offset, limit int, where string, whereKeyword string) (rs []*Role, total int, err error) {
	r := &Role{}
	if limit == 0 {
		limit = constvar.DefaultLimit
	}
	fieldsStr := "id,name"
	if len(where) > 0 {
		if err := DB.Self.Select(fieldsStr).Where(where+" = ?", whereKeyword).Offset(offset).Limit(limit).Find(&rs).Error; err != nil {
			return rs, 0, errors.New("cannot get Role list by where " + where + " and keyword " + whereKeyword)
		}

		if err := DB.Self.Model(r).Where(where+" = ?", whereKeyword).Count(&total).Error; err != nil {
			return rs, 0, errors.New("cannot fetch count of the row")
		}
	} else {
		if err := DB.Self.Select(fieldsStr).Offset(offset).Limit(limit).Find(&rs).Error; err != nil {
			return rs, 0, errors.New("cannot get Role list ")
		}
		if err := DB.Self.Model(r).Count(&total).Error; err != nil {
			return rs, 0, errors.New("cannot fetch count of the row")
		}
	}

	return rs, total, nil

}

//GetRoleRelatedUsers :
//func GetRoleRelatedUsers(rid uint64, offset, limit int) (users []User, total int, err error) {
//	if limit == 0 {
//		limit = constvar.DefaultLimit
//	}
//	r := &Role{}
//	r.ID = rid
//
//	uids := []uint64{}
//
//	selectSql := ""
//	countSql := ""
//	if rid == 0 {
//		selectSql = "SELECT user_id from user_roles offset " + strconv.Itoa(offset) + " limit " + strconv.Itoa(limit)
//	} else {
//		selectSql = "SELECT user_id from user_roles where role_id = " + util.Uint2Str(rid) + " offset " + strconv.Itoa(offset) + " limit " + strconv.Itoa(limit)
//		countSql = "SELECT  count(user_id) from user_roles where role_id = " + util.Uint2Str(rid)
//	}
//	rows, _ := DB.Self.Raw(selectSql).Rows() // Note: Ignoring errors for brevity
//
//	for rows.Next() {
//		var id uint64
//		if err := rows.Scan(&id); err != nil {
//			return nil, 0, err
//		}
//		uids = append(uids, id)
//	}
//
//	if err := DB.Self.Where(" id in (?)", uids).Find(&users).Error; err != nil {
//		return users, 0, err
//	}
//
//	if rid == 0 {
//		DB.Self.Model(User{}).Count(&total)
//	} else {
//		rows, _ := DB.Self.Raw(countSql).Rows()
//		for rows.Next() {
//			rows.Scan(&total)
//		}
//	}
//
//	return users, total, nil
//}

// CheckUsersNotInRole 根据给出的id列表 ，判断哪些不存在于role中
func CheckUsersNotInRole(roleId uint64, uids []uint64) (exceptList []uint64, err error) {
	//SELECT user_id
	//FROM (VALUES(4),(5),(6),(87)) V(user_id)
	//except
	//SELECT user_id
	//FROM user_roles where role_id=1;

	values := []string{}

	for _, id := range uids {
		values = append(values, "("+util.Uint2Str(id)+")")
	}

	//selectSql := `SELECT user_id FROM (VALUES`+strings.Join(values,",")+`) V(user_id) EXCEPT SELECT user_id FROM user_roles where role_id=` + util.Uint2Str(roleId)
	selectSql := `SELECT user_id FROM (VALUES` + strings.Join(values, ",") + `) V(user_id) EXCEPT SELECT user_id FROM user_roles where role_id=0`
	rows, _ := DB.Self.Debug().Raw(selectSql).Rows()
	for rows.Next() {
		var user_id uint64
		if err := rows.Scan(&user_id); err == nil {
			exceptList = append(exceptList, user_id)
		}
	}

	return exceptList, err
}

//GetRoleRelatedUsers :
func GetRoleRelatedUsers(id uint64, where string, value string, offset, limit int) (users []User, total int, err error) {
	if limit == 0 {
		limit = constvar.DefaultLimit
	}
	rs := []Role{}
	if err := DB.Self.Where("id = ?", id).Order("id").Find(&rs).Error; err != nil {
		return nil, 0, err
	}

	rids := make([]string, len(rs))
	for i, r := range rs {
		rids[i] = util.Uint2Str(r.ID)
	}
	uids := []uint64{}

	selectSql := ""
	countSql := ""
	if id == 0 {
		//如果id =0 ,就查询所有的用户
		selectSql = "SELECT user_id from user_roles ORDER BY id offset " + strconv.Itoa(offset) + " limit " + strconv.Itoa(limit)

		rows, _ := DB.Self.Debug().Raw(selectSql).Rows() // Note: Ignoring errors for brevity

		for rows.Next() {
			var id uint64
			if err := rows.Scan(&id); err != nil {
				return nil, 0, err
			}
			uids = append(uids, id)
		}

		if err := DB.Self.Where(" id in (?)", uids).Find(&users).Error; err != nil {
			return users, 0, err
		}

	} else {

		if len(where) > 0 {
			selectSql = "SELECT * FROM users left join user_roles on users.id=user_roles.user_id where users." + where + " like '%" + value + "%' And user_roles.role_id in (" + strings.Join(rids, ",") + ")" + " offset " + strconv.Itoa(offset) + " limit " + strconv.Itoa(limit)
			countSql = "SELECT count(users.id) FROM users left join user_roles on users.id=user_roles.user_id where users." + where + " like '%" + value + "%' And user_roles.role_id in (" + strings.Join(rids, ",") + ")"
		} else {
			selectSql = "SELECT * FROM users left join user_roles on users.id=user_roles.user_id where user_roles.role_id in (" + strings.Join(rids, ",") + ")" + " offset " + strconv.Itoa(offset) + " limit " + strconv.Itoa(limit)
			countSql = "SELECT count(users.id) FROM users left join user_roles on users.id=user_roles.user_id where user_roles.role_id in (" + strings.Join(rids, ",") + ")"
		}

		DB.Self.Debug().Raw(selectSql).Scan(&users) // Note: Ignoring errors for brevity

	}

	if id == 0 {
		DB.Self.Model(User{}).Count(&total)
	} else {
		rows, _ := DB.Self.Raw(countSql).Rows()
		for rows.Next() {
			rows.Scan(&total)
		}
	}

	return users, total, nil
}
