package models

import (
	"booking/pkg/constvar"
	"booking/util"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

//项目比较急，一开始设想的时候将Group用于管理职员档案了，现在需要一个新的群组Model来专门管理User，只能将model命名为Group了
// Group : 用于User管理的群组
type Group struct {
	BaseModel
	Name   string `json:"name" gorm:"column:name;not null"`
	Users  []User `json:"users" gorm:"many2many:user_groups;"`
	Parent uint64 `json:"parent" gorm:"column:parent;"`
	Levels string `json:"levels" gorm:"column:levels"` //保存父子层级关系图,例如 pppid.ppid.pid.id
}

// TableName :
func (g *Group) TableName() string {
	return "groups"
}

// Create : Create a new Group
func (g Group) Create() (group Group, err error) {
	pm := &Group{}
	g.Levels = "0."
	if g.Parent != 0 {
		pm.BaseModel.ID = g.Parent
		if err := DB.Self.First(&pm).Error; err != nil {
			return group, errors.New("找不到父目录")
		}
		g.Levels = pm.Levels + util.Uint2Str(g.Parent) + "."
	}

	err = DB.Self.Create(&g).Error
	return g, err
}

// Update updates an Group information.
// only update name and coefficient
func (g *Group) Update() error {
	_, err := GetGroup(g.ID, false)
	if err != nil {
		return err
	}

	tx := DB.Self.Begin()
	if err := tx.Model(&g).Update(map[string]interface{}{"name": g.Name, "parent": g.Parent}).Error; err != nil {
		tx.Rollback()
		return errors.New("无法更新")
	}
	tx.Commit()
	return nil
}

//GetAllGroup :
// params:
// @orderBy : 格式如下:created_at_DESC or created_at_ASC
func GetGroups(where string, value string, skip, take int, orderBy string) (gs []Group, total int, err error) {

	g := &Group{}

	fieldsStr := "id,name,parent,levels,created_at,updated_at,deleted_at"
	orderKey := ""
	orderType := "ASC"

	if len(orderBy) > 0 {
		if strings.Contains(orderBy, "_DESC") {
			orderKey = string([]rune(orderBy)[:len(orderBy)-5])
			orderType = "DESC"
		}

		if strings.Contains(orderBy, "_ASC") {
			orderKey = string([]rune(orderBy)[:len(orderBy)-4])
			orderType = "ASC"
		}

	}

	if len(where) > 0 {
		if len(orderKey) > 0 {
			if err := DB.Self.Select(fieldsStr).Where(where+" = ?", value).Order(orderKey + " " + orderType).Offset(skip).Limit(take).Find(&gs).Error; err != nil {
				return gs, 0, errors.New("cannot get Group list by where " + where + " and keyword " + value)
			}
		} else {
			if err := DB.Self.Select(fieldsStr).Where(where+" = ?", value).Offset(skip).Limit(take).Find(&gs).Error; err != nil {
				return gs, 0, errors.New("cannot get Group list by where " + where + " and keyword " + value)
			}
		}

		if err := DB.Self.Model(g).Where(where+" = ?", value).Count(&total).Error; err != nil {
			return gs, 0, errors.New("cannot fetch count of the row")
		}
	} else {
		if len(orderKey) > 0 {
			if err := DB.Self.Select(fieldsStr).Order(orderKey + " " + orderType).Offset(skip).Limit(take).Find(&gs).Error; err != nil {
				return gs, 0, errors.New("cannot get Group list ")
			}
		} else {
			if err := DB.Self.Select(fieldsStr).Offset(skip).Limit(take).Find(&gs).Error; err != nil {
				return gs, 0, errors.New("cannot get Group list ")
			}
		}
		if err := DB.Self.Model(g).Count(&total).Error; err != nil {
			return gs, 0, errors.New("cannot fetch count of the row")
		}
	}

	return gs, total, nil

}

//GetGroupRelatedUsers :
func GetGroupRelatedUsers(id uint64, offset, limit int) (users []User, total int, err error) {
	if limit == 0 {
		limit = constvar.DefaultLimit
	}
	gs := []Group{}
	if err := DB.Self.Where("levels LIKE ? OR id = ?", "%."+util.Uint2Str(id)+".%", id).Order("id").Find(&gs).Error; err != nil {
		fmt.Println(err)
		return nil, 0, err
	}

	gids := make([]string, len(gs))
	for i, g := range gs {
		gids[i] = util.Uint2Str(g.ID)
	}
	uids := []uint64{}

	selectSql := ""
	countSql := ""
	if id == 0 {
		selectSql = "SELECT user_id from user_groups ORDER BY id offset " + strconv.Itoa(offset) + " limit " + strconv.Itoa(limit)
	} else {
		selectSql = "SELECT user_id from user_groups where group_id in (" + strings.Join(gids, ",") + ")" + " ORDER BY id offset " + strconv.Itoa(offset) + " limit " + strconv.Itoa(limit)
		countSql = "SELECT  count(user_id) from user_groups where group_id in (" + strings.Join(gids, ",") + ")"
	}
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

func AddUserToDefaultGroup(uid uint64) (err error) {
	gname := viper.GetString("company.name")
	g := &Group{}
	if err := DB.Self.Where("name = ?", gname).First(g).Error; err != nil {
		return err
	}

	idlist := []uint64{uid}
	err = AddGroupUsers(g.ID, idlist)
	return err

}

//AddGroupUsers :
func AddGroupUsers(gid uint64, uids []uint64) (err error) {

	g := &Group{}

	if g, err = GetGroup(gid, false); err != nil {
		return errors.New("Group not existed!")
	}

	tx := DB.Self.Begin()

	var users []User
	for _, id := range uids {
		u := User{}
		u.ID = id
		users = append(users, u)
		tx.Model(&u).Association("Groups").Clear()
	}

	err = tx.Model(&g).Association("Users").Append(users).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

//RemoveGroupUsers :
func RemoveUsersFromGroup(gid uint64, uids []uint64) (err error) {
	g := &Group{}
	if g, err = GetGroup(gid, false); err != nil {
		return errors.New("User Group is not existed!")
	}

	tx := DB.Self.Begin()

	newUids := make([]string, len(uids))

	for i, id := range uids {
		newUids[i] = util.Uint2Str(id)
	}
	err = tx.Model(&g).Exec(" delete from user_groups where user_id in (" + strings.Join(newUids, ",") + ") and group_id = " + util.Uint2Str(gid) + " ;").Error
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

// GetGroup :
func GetGroup(id uint64, withUsers bool) (result *Group, err error) {
	g := &Group{}
	if id == 0 {
		return result, errors.New("cannot find Group by id " + util.Uint2Str(id))
	}
	err = DB.Self.Select("id,name,parent,levels").First(&g, id).Error
	if withUsers {
		DB.Self.Model(&result).Select("id").Association("Users").Find(&g.Users)
	}
	return g, err
}

// GetGroupByName :
func GetGroupByName(name string) (result *Group, err error) {
	g := &Group{}
	err = DB.Self.Select("id,name,parent,levels").Where("name = ?", name).First(&g).Error
	return g, err
}

// DeleteGroup : delete children Group when parent had deleted
func DeleteGroup(id uint64) error {

	group, err := GetGroup(id, false)
	if err != nil {
		return err
	}
	cat := &Group{}
	cat.ID = group.ID
	tx := DB.Self.Begin()
	if err := tx.Where("levels LIKE ? OR id = ?", "%."+util.Uint2Str(id)+".%", id).Delete(&cat).Error; err != nil {
		tx.Rollback()
		return errors.New("无法删除")
	}
	tx.Commit()

	return nil
}
