package models

import (
	"booking/util"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Booking struct {
	BaseModel
	UserID      uint64 `gorm:"column:user_id;not null"`
	CanteenID   uint64 `gorm:"not null"`
	BookingDate string `gorm:"not null"`
	Number      int
	BookingType string //预订的类型，早餐 ，午餐 ，晚餐 breakfast,lunch,dinner
	Available   bool
}

// TableName :
func (b *Booking) TableName() string {
	return "booking"
}

// Create :
func (b Booking) Create() (booking Booking, err error) {
	if b.BookingType == "breakfast" || b.BookingType == "lunch" || b.BookingType == "dinner" {
		err = DB.Self.Create(&b).Error
	} else {
		err = errors.New("booking type is unavailable:" + b.BookingType)
	}

	return b, err
}

//检查是否预订过
func (b Booking) Check() (results []Booking, err error) {
	w := "user_id=" + util.Uint2Str(b.UserID) + " AND canteen_id=" + util.Uint2Str(b.CanteenID) + " AND booking_date='" +
		b.BookingDate + "' AND booking_type='" + b.BookingType + "' AND available=true "

	bookings := make([]Booking, 0)

	if err = DB.Self.Debug().Where(w).Find(&bookings).Error; err != nil {
		return bookings, err
	}

	if len(bookings) > 0 {
		return bookings, nil
	}

	return bookings, errors.New("nothing to find")

}

// CancelBooking  取消预订
func CancelBooking(userID, bookingID uint64) error {
	booking, err := GetBooking(bookingID)
	if err != nil {
		return err
	}

	if booking.UserID != userID {
		return errors.New("无权操作")
	}

	canteen, err := GetCanteen(booking.CanteenID)

	if err != nil {
		return err
	}

	duringTime := ""

	if booking.BookingType == "breakfast" {
		duringTime = canteen.BreakfastTime
	}
	if booking.BookingType == "lunch" {
		duringTime = canteen.LunchTime
	}
	if booking.BookingType == "dinner" {
		duringTime = canteen.DinnerTime
	}

	// 时间格式 ：  "07:00-9:00"
	during := strings.Split(duringTime, "-")

	startTimeString := booking.BookingDate + " " + util.FillTimeFormat(during[0])

	startTime, _ := time.ParseInLocation("2006-01-02 15:04:05", startTimeString, time.Local)

	// 2个小时前
	h, _ := time.ParseDuration("-1h")
	cancelTime := startTime.Add(time.Duration(canteen.CancelTime) * h)
	now := time.Now()

	//
	if now.Before(cancelTime) {
		//处理逻辑
		return DeleteBooking(bookingID)
	} else {
		return errors.New("必须提前" + fmt.Sprint(canteen.CancelTime) + "小时才能取消预订")
	}

	//取消预订之后，要重新给用户发放一个票
	_, err = BatchCreateTickets(booking.UserID, booking.Number, TicketTypeMap[booking.BookingType], 0)

	if err != nil {
		return err
	}

	return nil
}

func (b *Booking) Update(data map[string]interface{}) error {
	_, err := GetBooking(b.ID)
	if err != nil {
		return err
	}

	tx := DB.Self.Begin()
	if err := tx.Model(&b).Update(data).Error; err != nil {
		tx.Rollback()
		return errors.New("无法更新")
	}
	tx.Commit()
	return nil
}

// GetBooking :
func GetBooking(id uint64) (result *Booking, err error) {
	d := &Booking{}
	if id == 0 {
		return result, errors.New("cannot find Booking by id " + util.Uint2Str(id))
	}
	err = DB.Self.Debug().First(&d, id).Error
	return d, err
}

// GetAllBooking
func GetAllBooking(where string, value string, skip, take int) (bookings []Booking, total int, err error) {
	u := &Booking{}
	w := ""
	if len(where) > 0 && len(value) > 0 {

		w = where + " = ?"
		v := value

		d := DB.Self.Debug().Where(w, v).Order("id").Offset(skip).Limit(take).Find(&bookings)

		if err := DB.Self.Model(u).Where(w, v).Count(&total).Error; err != nil {
			return bookings, 0, errors.New("cannot fetch count of the row")
		}
		return bookings, total, d.Error
	}

	d := DB.Self.Debug().Order("id").Offset(skip).Limit(take).Find(&bookings)
	if err := DB.Self.Model(u).Count(&total).Error; err != nil {
		return bookings, 0, errors.New("cannot fetch count of the row")
	}
	return bookings, total, d.Error

}


func CountBookingByMonth(year,month int,canteenIds []uint64) (data []map[string]interface{}, err error) {

	dates := []string{}

	yearStr := strconv.Itoa(year)
	monthStr := strconv.Itoa(month)

	if len(yearStr) == 1 {
		yearStr = "0" + yearStr
	}

	if len(monthStr) == 1 {
		monthStr = "0" + monthStr
	}

	startDay :=  yearStr+"-"+monthStr+"-01"

	startTime, _ := time.ParseInLocation("2006-01-02",startDay, time.Local)

	h, _ := time.ParseDuration("1h")
	dates = append(dates, `'`+startDay+`'`)
	days := util.CountDays(year,month)

	for i := 1;i < days; i++ {
		dates = append(dates, `'`+startTime.Add(time.Duration(i*24)*h).Format("2006-01-02")+`'`)
	}

	countSql := `SELECT a.user_id,b.username,sum(case when booking_type='breakfast' then a.number else 0 end) as breakfast , sum(case when booking_type='lunch' then a.number else 0 end) as lunch,
		sum(case when booking_type='dinner' then a.number else 0 end) as dinner  from booking as a right join tb_users as b on a.user_id=b.id  AND a.booking_date IN (` + strings.Join(dates, ",") + `) AND a.canteen_id IN (`+util.ArrayToString(canteenIds,",") +`) group by a.user_id,b.username;`
	rows, _ := DB.Self.Debug().Raw(countSql).Rows()

	data = make([]map[string]interface{},0)

	for rows.Next() {
		user_id := ""
		username := ""
		breakfast := 0
		lunch := 0
		dinner := 0

		err = rows.Scan(&user_id,&username, &breakfast, &lunch, &dinner)
		d := make(map[string]interface{})
		d["username"] = username
		d["breakfast"] = breakfast
		d["lunch"] = lunch
		d["dinner"] = dinner
		data = append(data,d)
	}
	return data, nil
}



// CountBookingByCanteen
// 只显示未来7天的预订情况

func CountBookingByCanteen(cid uint64) (data map[string]map[string]int, err error) {

	dates := []string{}

	today := time.Now().Format("2006-01-02")
	startTime, _ := time.ParseInLocation("2006-01-02", today, time.Local)

	h, _ := time.ParseDuration("1h")
	dates = append(dates, `'`+today+`'`)
	dates = append(dates, `'`+startTime.Add(time.Duration(24)*h).Format("2006-01-02")+`'`)
	dates = append(dates, `'`+startTime.Add(time.Duration(48)*h).Format("2006-01-02")+`'`)
	dates = append(dates, `'`+startTime.Add(time.Duration(72)*h).Format("2006-01-02")+`'`)
	dates = append(dates, `'`+startTime.Add(time.Duration(96)*h).Format("2006-01-02")+`'`)
	dates = append(dates, 	`'`+startTime.Add(time.Duration(120)*h).Format("2006-01-02")+`'`)
	dates = append(dates, `'`+startTime.Add(time.Duration(144)*h).Format("2006-01-02")+`'`)

	countSql := `SELECT booking_date,sum(case when booking_type='breakfast' then 1 else 0 end) as breakfast , sum(case when booking_type='lunch' then 1 else 0 end) as lunch  ,` +
		`sum(case when booking_type='dinner' then 1 else 0 end) as dinner  from booking where canteen_id =` + util.Uint2Str(cid) + ` AND booking_date IN (` + strings.Join(dates, ",") + `) group by booking_date order by booking_date`
	rows, _ := DB.Self.Debug().Raw(countSql).Rows()

	data = make(map[string]map[string]int)

	for rows.Next() {
		date := ""
		breakfast := 0
		lunch := 0
		dinner := 0

		err = rows.Scan(&date, &breakfast, &lunch, &dinner)
		if _, ok := data[date]; !ok {
			data[date] = make(map[string]int)
		}
		data[date]["breakfast"] = breakfast
		data[date]["lunch"] = lunch
		data[date]["dinner"] = dinner
	}
	return data, nil
}

// DeleteBooking
func DeleteBooking(id uint64) error {

	booking, err := GetBooking(id)
	if err != nil {
		return err
	}
	b := &Booking{}
	b.ID = booking.ID
	tx := DB.Self.Begin()
	if err := tx.Where("id = ?", id).Delete(&b).Error; err != nil {
		tx.Rollback()
		return errors.New("无法删除")
	}
	tx.Commit()

	return nil
}
