package models

import (
	"errors"
	"hrgdrc/util"
)

type Dishes struct {
	BaseModel
	Name    string `json:"name" gorm:"column:name;not null"`
	Picture string `json:"picture"`
}

// TableName :
func (g *Dishes) TableName() string {
	return "dishes"
}

// Create : Create a new Group
func (g Dishes) Create() (dish Dishes, err error) {
	m := &Dishes{}
	err = DB.Self.Create(&m).Error
	return g, err
}

// Update updates an Group information.
// only update name and coefficient
func (g *Dishes) Update(data map[string]interface{}) error {
	_, err := GetDish(g.ID)
	if err != nil {
		return err
	}

	tx := DB.Self.Begin()
	if err := tx.Model(&g).Update(data).Error; err != nil {
		tx.Rollback()
		return errors.New("无法更新")
	}
	tx.Commit()
	return nil
}

// GetDish :
func GetDish(id uint64) (result *Dishes, err error) {
	d := &Dishes{}
	if id == 0 {
		return result, errors.New("cannot find Dish by id " + util.Uint2Str(id))
	}
	err = DB.Self.First(&d, id).Error
	return d, err
}

// GetUsers
func GetDishes(where string, value string, skip, take int) (dishes []Dishes, total int, err error) {
	u := &Dishes{}
	w := ""
	if len(where) > 0 && len(value) > 0 {

		w = where + " LIKE ?"
		v := "%" + value + "%"

		d := DB.Self.Debug().Where(w, v).Order("id").Offset(skip).Limit(take).Find(&dishes)

		if err := DB.Self.Model(u).Where(w, v).Count(&total).Error; err != nil {
			return dishes, 0, errors.New("cannot fetch count of the row")
		}
		return dishes, total, d.Error
	}

	d := DB.Self.Debug().Order("id").Offset(skip).Limit(take).Find(&dishes)
	if err := DB.Self.Model(u).Count(&total).Error; err != nil {
		return dishes, 0, errors.New("cannot fetch count of the row")
	}
	return dishes, total, d.Error

}

// DeleteDish
func DeleteDish(id uint64) error {

	dish, err := GetDish(id)
	if err != nil {
		return err
	}
	d := &Dishes{}
	d.ID = dish.ID
	tx := DB.Self.Begin()
	if err := tx.Where("id = ?", id).Delete(&d).Error; err != nil {
		tx.Rollback()
		return errors.New("无法删除")
	}
	tx.Commit()

	return nil
}
