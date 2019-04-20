package models

import (
	"errors"
	"fmt"
	"hrgdrc/util"
	"strings"
	"time"
)

type Booking struct{
	BaseModel
	UserID  uint64 `gorm:"column:user_id;not null"`
	CanteenID uint64 `gorm:"not null"`
	BookingDate    string `gorm:"not null"`
	BookingType    string //预订的类型，早餐 ，午餐 ，晚餐 Breakfast,Lunch,Dinner
}

// TableName :
func (b *Booking) TableName() string {
	return "booking"
}

// Create :
func (b Booking) Create() (booking Booking, err error) {
	if b.BookingType == "breakfast" ||  b.BookingType == "lunch" || b.BookingType == "dinner" {
		err = DB.Self.Create(&b).Error
	}else{
		err = errors.New("booking type is unavailable:" + b.BookingType)
	}

	return b, err
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

	canteen, err  := GetCanteen(booking.CanteenID)

	if err != nil {
		return err
	}

	duringTime := ""

	if booking.BookingType == "Breakfast" {
		duringTime =  canteen.BreakfastTime
	}
	if booking.BookingType == "Lunch" {
		duringTime =  canteen.LunchTime
	}
	if booking.BookingType == "Dinner" {
		duringTime =  canteen.DinnerTime
	}

	// 时间格式 ：  "07:00-9:00"
	during := strings.Split(duringTime, "-")

	startTimeString := fillTimeFormat(during[0])

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

	return nil
}

//如果是7:00 ,要将之转化为07:00
func fillTimeFormat(timeStr string) string {
	hourMinute := strings.Split(timeStr, ":")
	hour := hourMinute[0]
	minute := hourMinute[1]

	if len(hour) == 1 {
		hour = "0" + hour
	}

	if len(minute) == 1 {
		minute = "0" + minute
	}

	t := time.Now().Format("2006-01-02")
	return t + " " + hour + ":" + minute + ":00"

}


// GetBooking :
func GetBooking(id uint64) (result *Booking, err error) {
	d := &Booking{}
	if id == 0 {
		return result, errors.New("cannot find Booking by id " + util.Uint2Str(id))
	}
	err = DB.Self.First(&d, id).Error
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

// CountBookingByCanteen
// 只显示未来7天的预订情况

func CountBookingByCanteen(cid uint64) (data map[string]map[string]int, err error) {

	dates := []string{}

	today := time.Now().Format("2006-01-02")
	startTime, _ := time.ParseInLocation("2006-01-02",today, time.Local)

	h, _ := time.ParseDuration("1h")
	dates = append(dates,`'`+today +`'`)
	dates = append(dates,`'`+startTime.Add(time.Duration(24) * h).Format("2006-01-02")+`'`)
	dates = append(dates,`'`+startTime.Add(time.Duration(48) * h).Format("2006-01-02")+`'`)
	dates = append(dates,`'`+startTime.Add(time.Duration(72) * h).Format("2006-01-02")+`'`)
	dates = append(dates,`'`+startTime.Add(time.Duration(96) * h).Format("2006-01-02")+`'`)
	dates = append(dates,`'`+startTime.Add(time.Duration(120) * h).Format("2006-01-02")+`'`)
	dates = append(dates,`'`+startTime.Add(time.Duration(144) * h).Format("2006-01-02")+`'`)


	countSql := `SELECT booking_date,sum(case when booking_type='breakfast' then 1 else 0 end) as breakfast , sum(case when booking_type='lunch' then 1 else 0 end) as lunch  ,`+
				`sum(case when booking_type='dinner' then 1 else 0 end) as dinner  from booking where canteen_id =` + util.Uint2Str(cid) +  ` AND booking_date IN (`+ strings.Join(dates,",") +`) group by booking_date order by booking_date`
	rows, _ := DB.Self.Debug().Raw(countSql).Rows()

	data = make(map[string]map[string]int)

	for rows.Next() {
		date := ""
		breakfast := 0
		lunch  := 0
		dinner := 0

		err = rows.Scan(&date, &breakfast,&lunch,&dinner)
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
