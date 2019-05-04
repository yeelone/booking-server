package models

import (
	"booking/util"
	"errors"
	"strings"
)

// Canteen : 食堂，一个组可对应多个食堂，一个食堂只属于一个组
type Canteen struct {
	BaseModel
	Name                     string `json:"name" gorm:"column:name;not null"`
	GroupID                  uint64 `json:"group_id" gorm:"column:group_id;not null"`
	BreakfastTime            string //早餐时间，格式为"7:00, 9:00"
	BreakfastPicture         string
	BookingBreakfastDeadline string //预订截止早餐时间，格式为"7:00"
	LunchPicture             string
	LunchTime                string //午餐时间，格式为"7:00, 9:00"
	BookingLunchDeadline     string
	DinnerPicture            string
	DinnerTime               string //晚餐时间，格式为"7:00, 9:00"
	BookingDinnerDeadline    string
	CancelTime               int    //取消时间，提前一小时，两小时
	Qrcode                   string //
	QrcodeUUID               string // 用于辨别二维码的有效性
	AdminID                  uint64
}

// TableName :
func (c *Canteen) TableName() string {
	return "canteens"
}

// Create : Create a new Canteen
func (c Canteen) Create() (canteen Canteen, err error) {
	_, err = GetGroup(c.GroupID, false)
	if err != nil {
		return canteen, err
	}

	err = DB.Self.Create(&c).Error
	return c, err
}

// Update updates an Canteen information.
// only update name and coefficient
func (c *Canteen) Update(data map[string]interface{}) error {
	_, err := GetCanteen(c.ID)
	if err != nil {
		return err
	}

	tx := DB.Self.Begin()
	if err := tx.Model(&c).Update(data).Error; err != nil {
		tx.Rollback()
		return errors.New("无法更新")
	}
	tx.Commit()
	return nil
}

// GetAllCanteen :
// params:
// @orderBy : 格式如下:created_at_DESC or created_at_ASC
func GetCanteens(where string, value string, skip, take int, orderBy string) (cs []Canteen, total int, err error) {

	g := &Canteen{}

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
			if err := DB.Self.Where(where+" = ?", value).Order(orderKey + " " + orderType).Offset(skip).Limit(take).Find(&cs).Error; err != nil {
				return cs, 0, errors.New("cannot get Canteen list by where " + where + " and keyword " + value)
			}
		} else {
			if err := DB.Self.Where(where+" = ?", value).Offset(skip).Limit(take).Find(&cs).Error; err != nil {
				return cs, 0, errors.New("cannot get Canteen list by where " + where + " and keyword " + value)
			}
		}

		if err := DB.Self.Model(g).Where(where+" = ?", value).Count(&total).Error; err != nil {
			return cs, 0, errors.New("cannot fetch count of the row")
		}
	} else {
		if len(orderKey) > 0 {
			if err := DB.Self.Order(orderKey + " " + orderType).Offset(skip).Limit(take).Find(&cs).Error; err != nil {
				return cs, 0, errors.New("cannot get Canteen list ")
			}
		} else {
			if err := DB.Self.Offset(skip).Limit(take).Find(&cs).Error; err != nil {
				return cs, 0, errors.New("cannot get Canteen list ")
			}
		}
		if err := DB.Self.Model(g).Count(&total).Error; err != nil {
			return cs, 0, errors.New("cannot fetch count of the row")
		}
	}

	return cs, total, nil

}

// GetCanteen :
func GetCanteen(id uint64) (result *Canteen, err error) {
	g := &Canteen{}
	if id == 0 {
		return result, errors.New("cannot find Canteen by id " + util.Uint2Str(id))
	}
	err = DB.Self.First(&g, id).Error
	return g, err
}

// DeleteCanteen : delete children Group when parent had deleted
func DeleteCanteen(id uint64) error {

	canteen, err := GetCanteen(id)
	if err != nil {
		return err
	}
	cant := &Canteen{}
	cant.ID = canteen.ID
	tx := DB.Self.Begin()
	if err := tx.Where("id = ?", id).Delete(&cant).Error; err != nil {
		tx.Rollback()
		return errors.New("无法删除")
	}
	tx.Commit()

	return nil
}
