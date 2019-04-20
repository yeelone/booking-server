package models

import (
	"booking/util"
	"errors"
	"strconv"
	"time"
)

type TicketRecord struct {
	BaseModel
	Operator    uint64 //操作员
	Owner       uint64 // 票拥有者
	Number      int    //
	Action      string // 售出 或者 加收
	Description string
}

const TicketRecordTableName = "ticket_record"

// TableName :
func (g *TicketRecord) TableName() string {
	return TicketRecordTableName
}

func (g TicketRecord) Create() (record TicketRecord, err error) {
	err = DB.Self.Create(&g).Error
	return record, err
}

// DeleteTicketRecord
func DeleteTicketRecord(id uint64) error {
	record := &TicketRecord{}
	tx := DB.Self.Begin()
	if err := tx.Where("id = ?", id).Delete(&record).Error; err != nil {
		tx.Rollback()
		return errors.New("无法删除")
	}
	tx.Commit()

	return nil
}

// GetTicketRecords
func GetTicketRecords(where string, value string, skip, take int) (records []TicketRecord, total int, err error) {
	u := &TicketRecord{}
	w := ""
	if len(where) > 0 && len(value) > 0 {
		v := ""

		if where == "owner" {
			w = where + " = ?"
			v = value
		}

		if where == "operator" {
			w = where + " = ?"
			v = value
		}

		d := DB.Self.Debug().Where(w, v).Order("id").Offset(skip).Limit(take).Find(&records)

		if err := DB.Self.Model(u).Where(w, v).Count(&total).Error; err != nil {
			return records, 0, errors.New("cannot fetch count of the row")
		}
		return records, total, d.Error
	}

	d := DB.Self.Debug().Order("id").Offset(skip).Limit(take).Find(&records)
	if err := DB.Self.Model(u).Count(&total).Error; err != nil {
		return records, 0, errors.New("cannot fetch count of the row")
	}
	return records, total, d.Error

}

func GetLatestTicketRecord(limit int) (records []string) {
	sqlstr := `select b.username as operator,a.action,c.username as owner,a.number,a.created_at  from ticket_record as a left join users as b on a.operator=b.id  left join users as c on a.owner=c.id order by  created_at desc limit  ` + strconv.Itoa(limit)
	rows, _ := DB.Self.Raw(sqlstr).Rows()
	records = []string{}
	for rows.Next() {
		operator := ""
		action := ""
		owner := ""
		number := 0
		created_at := ""
		rows.Scan(&operator, &action, &owner, &number, &created_at)

		if action == "sell" {
			action = "售出"
		}
		if action == "recycling" {
			action = "回收"
		}

		n := util.Abs(number)
		t, _ := time.Parse(time.RFC3339, created_at)

		records = append(records, t.Format("2006-01-02 15:04:05")+" | "+operator+" | "+action+" | "+owner+" | "+strconv.Itoa(n)+"张票 ")
	}
	return records
}
