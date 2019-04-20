package models

import (
	"booking/util"
	"errors"
	"fmt"
	"github.com/rs/xid"
	"strconv"
	"strings"
	"time"
)

// ticket type  : 票类型，早餐、午餐、晚餐 ， 用数字表示 ，早餐=1， 午餐=2， 晚餐=3

type Ticket struct {
	ID             uint64    `gorm:"primary_key;AUTO_INCREMENT;column:id" json:"id"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"-"`
	UpdatedAt      time.Time `gorm:"column:updated_at" json:"-"`
	UUID           string    `json:"uuid"`
	UserID         uint64
	Type           int
	Price          int
	TransferRecord string `json:"transfer_record"`
}

const TicketTableName = "tickets"

// TableName :
func (g *Ticket) TableName() string {
	return TicketTableName
}

// Create : Create a new Group
func (g Ticket) Create() (ticket Ticket, err error) {
	m := &Ticket{}
	guid := xid.New()
	m.UUID = guid.String()
	err = DB.Self.Create(&m).Error
	return g, err
}

// 批量生成电子票
func BatchCreateTickets(userId uint64, number, ticketType, price int) (tickets []Ticket, err error) {
	keys := "created_at,user_id,uuid,type,price"
	sql := "insert into " + TicketTableName + "(" + keys + ") values"
	values := []string{}

	if number < 0 {
		return tickets, errors.New("must be bigger than 0")
	}
	tickets = make([]Ticket, number)
	for i := 0; i < number; i++ {
		guid := xid.New()
		s := `$1,'` + util.Uint2Str(userId) + `','` + guid.String() + `','` + strconv.Itoa(ticketType) + `','` + strconv.Itoa(price) + `'`
		values = append(values, `(`+s+`)`)
		ticket := Ticket{UserID: userId, UUID: guid.String()}
		tickets = append(tickets, ticket)
	}

	sql += strings.Join(values, ",")

	return tickets, DB.Self.Exec(sql, time.Now()).Error
}

// Update updates an Group information.
// only update name and coefficient
func (g *Ticket) Update(data map[string]interface{}) error {
	_, err := GetTicket(g.ID)
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

// GetTicket :
func GetTicket(id uint64) (result *Ticket, err error) {
	d := &Ticket{}
	if id == 0 {
		return result, errors.New("cannot find Ticket by id " + util.Uint2Str(id))
	}
	err = DB.Self.First(&d, id).Error
	return d, err
}

// GetTickets
func GetTickets(where string, value string, skip, take int) (tickets []Ticket, total int, err error) {
	u := &Ticket{}
	w := ""
	if len(where) > 0 && len(value) > 0 {
		v := ""

		if where == "user_id" {
			w = where + " = ?"
			v = value
		} else {
			w = where + " LIKE ?"
			v = "%" + value + "%"
		}

		d := DB.Self.Debug().Where(w, v).Order("id").Offset(skip).Limit(take).Find(&tickets)

		if err := DB.Self.Model(u).Where(w, v).Count(&total).Error; err != nil {
			return tickets, 0, errors.New("cannot fetch count of the row")
		}
		return tickets, total, d.Error
	}

	d := DB.Self.Debug().Order("id").Offset(skip).Limit(take).Find(&tickets)
	if err := DB.Self.Model(u).Count(&total).Error; err != nil {
		return tickets, 0, errors.New("cannot fetch count of the row")
	}
	return tickets, total, d.Error

}

// TransferTicket 将电子票转让给其它用户
func TransferTicket(fromUserId uint64, toUserId uint64, number int) (success, error int, err error) {

	fromUser, err := GetUserByID(fromUserId)

	if err != nil {
		return 0, 0, errors.New("该用户不存在")
	}

	toUser, err := GetUserByID(toUserId)

	if err != nil {
		return 0, 0, errors.New("转让用户不存在")
	}

	// 从from用户里挑选出number张电子票
	tickets, _, err := GetTickets("user_id", util.Uint2Str(fromUserId), 0, number)

	record := "由[" + util.Uint2Str(fromUserId) + "_" + fromUser.Username + "] 转让给 [" + util.Uint2Str(toUserId) + "_" + toUser.Username + "] | "

	errorCount := 0
	for _, t := range tickets {
		data := make(map[string]interface{})
		data["user_id"] = toUserId
		data["transfer_record"] = t.TransferRecord + record
		if err := t.Update(data); err != nil {
			errorCount++
		}
	}

	return number - errorCount, errorCount, nil

}

// DeleteTicket
func DeleteTicket(id uint64) error {
	ticket, err := GetTicket(id)
	if err != nil {
		return err
	}
	d := &Ticket{}
	d.ID = ticket.ID
	tx := DB.Self.Begin()
	if err := tx.Where("id = ?", id).Delete(&d).Error; err != nil {
		tx.Rollback()
		return errors.New("无法删除")
	}
	tx.Commit()

	return nil
}

// RecyclingTickets 回收用户电子票N张
func RecyclingTickets(uid uint64, ticketType, n int) error {
	//找出最近生成的票
	n = util.Abs(n)

	tickets := make([]Ticket, n)
	err := DB.Self.Debug().Where("user_id=? AND type=?", uid, ticketType).Order("created_at DESC").Limit(n).Find(&tickets).Error

	ids := make([]uint64, n)

	for _, t := range tickets {
		ids = append(ids, t.ID)
	}

	fmt.Println("ids", ids)
	tx := DB.Self.Begin()
	if err := tx.Where("id IN (?)", ids).Delete(&Ticket{}).Error; err != nil {
		tx.Rollback()
		return errors.New("无法删除")
	}
	tx.Commit()

	return err
}
