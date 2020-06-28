package graph

import (
	"booking/graph/model"
	"booking/models"
	"booking/util"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"time"
)

type bookingResolver struct{ *Resolver }

func (r *bookingResolver) ID(ctx context.Context, obj *models.Booking) (string, error) {
	return fmt.Sprintf("%d", obj.BaseModel.ID), nil
}

func (r *bookingResolver) DeletedAt(ctx context.Context, obj *models.Booking) (string, error) {
	if obj.DeletedAt != nil {
		return fmt.Sprintf(obj.DeletedAt.Format("2006-01-02 15:04:05")), nil
	}

	return "", nil
}

func (r *bookingResolver) CanteenID(ctx context.Context, obj *models.Booking) (int, error) {
	return int(obj.CanteenID), nil
}

func (r *bookingResolver) Type(ctx context.Context, obj *models.Booking) (string, error) {
	return obj.BookingType, nil
}

func (r *bookingResolver) Date(ctx context.Context, obj *models.Booking) (string, error) {
	return obj.BookingDate, nil
}

func (r *bookingResolver) UserID(ctx context.Context, obj *models.Booking) (int, error) {
	return int(obj.UserID), nil
}

func (r *bookingResolver) CreatedAt(ctx context.Context, obj *models.Booking) (string, error) {
	return fmt.Sprintf(obj.CreatedAt.Format("2006-01-02 15:04:05")), nil
}

func (r *bookingResolver) UpdatedAt(ctx context.Context, obj *models.Booking) (string, error) {
	return fmt.Sprintf(obj.UpdatedAt.Format("2006-01-02 15:04:05")), nil
}

func (r *mutationResolver) CancelBooking(ctx context.Context, input model.CancelBookingInput) (bool, error) {
	err := models.CancelBooking(uint64(input.UserID), uint64(input.BookingID))

	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *mutationResolver) Spend(ctx context.Context, input model.SpendInput) (bool, error) {
	//defer func() {
	//	if r := recover(); r != nil {
	//		fmt.Println("Recovered in f", r)
	//	}
	//}()

	canteen := models.Canteen{}

	if input.UUID != "" {

		//tunnel := &Tunnel{Name: input.UUID, Observers: map[string]chan models.Message{}}

		//for k := range r.tunnels {
		//	fmt.Println("k",k )
		//}

		if _, ok := r.tunnels[input.UUID]; !ok {
			return false, errors.New("食堂管理员未上线，请稍候再试")
		}

		tunnel := r.tunnels[input.UUID]

		//先根据uuid查找属于的食堂
		canteens, _, err := models.GetCanteens("qrcode_uuid", input.UUID, 0, 1, "")

		if err != nil || len(canteens) < 1 {
			return false, errors.New("二维码已过期,请联系管理员")
			//fmt.Println(errors.New("二维码已过期,请联系管理员"))
		}

		canteen = canteens[0]
		user, _ := models.GetUserByID(canteen.AdminID)
		message := &models.Message{
			ID:        user.ID,
			CreatedAt: time.Now(),
			Text:      input.UUID,
			CreatedBy: user,
			Error:     false,
		}

		if uint64(input.CanteenID) != canteen.ID {
			message.Error = true
			message.ErrorText = "二维码不可用,请联系管理员"
			return false, errors.New("食堂标识ID不符.请确认二维码")
		}

		booking := models.Booking{}
		if ok, _ := util.CheckTimeRange(canteen.BreakfastTime); ok {
			// 现在是早餐时间
			booking.BookingType = "breakfast"
		}

		if ok, _ := util.CheckTimeRange(canteen.LunchTime); ok {
			// 现在是午餐时间
			booking.BookingType = "lunch"
		}

		if ok, _ := util.CheckTimeRange(canteen.DinnerTime); ok {
			//现在是晚餐时间
			booking.BookingType = "dinner"
		}

		booking.BookingType = "lunch"
		if booking.BookingType == "" {
			return false, errors.New("已过了用餐时间")
		}

		booking.BookingDate = time.Now().Format("2006-01-02")
		booking.CanteenID = uint64(input.CanteenID)
		booking.UserID = uint64(input.UserID)

		if bookings, err := booking.Check(); err == nil {

			for k, observer := range tunnel.Observers {

				for _, item := range bookings {
					data := make(map[string]interface{})
					data["available"] = false
					if err := item.Update(data); err != nil {
						return false, err
					}

					if k == strconv.Itoa(int(canteen.AdminID)) {
						message.Text = strconv.Itoa(item.Number)
						fmt.Println(util.PrettyJson(message),item.Number)
						observer <- message
						break
					}
				}

			}
		} else {
			return false, errors.New("您是否有预订？")
		}

	}

	return false, nil

}

func (r *mutationResolver) Booking(ctx context.Context, input model.BookingInput) (bool, error) {
	date := strings.Split(input.Date, "-")
	breakfast, lunch, dinner, err := models.CountTicketsDetailByUser(uint64(input.UserID))

	if err != nil {
		return false, err
	}

	errs := []string{}

	//自动预订一整个月份
	if input.AutoCurrentMonth != nil {
		year, _ := strconv.Atoi(date[0])
		month, _ := strconv.Atoi(date[1])
		days := util.CountDays(year, month)

		for i := 1; i < days; i++ {
			booking := &models.Booking{}
			booking.UserID = uint64(input.UserID)
			booking.CanteenID = uint64(input.CanteenID)
			booking.BookingType = input.Type.String()
			booking.BookingDate = date[0] + "-" + date[1] + "-" + strconv.Itoa(i)
			booking.Available = true

			switch input.Type.String() {
			case "breakfast":
				if breakfast < 1 {
					return false, errors.New("没有可用早餐卷")

				}
			case "lunch":
				if lunch < 1 {
					return false, errors.New("没有可用午餐卷")
				}
			case "dinner":
				if dinner < 1 {
					return false, errors.New("没有可用晚餐卷")
				}
			}

			if _, err := booking.Create(); err != nil {
				errs = append(errs, err.Error())
			}

			//预订即回收餐票
			if err := models.RecyclingTickets(booking.UserID, models.TicketTypeMap[booking.BookingType], 1); err != nil {
				models.DeleteBooking(booking.ID)
				return false, err
			}

			switch input.Type.String() {
			case "breakfast":
				breakfast = breakfast - 1
			case "lunch":
				lunch = lunch - 1
			case "dinner":
				dinner = dinner - 1
			}

		}

		if len(errs) > 0 {
			return false, errors.New(strings.Join(errs, "|"))
		}
	}

	switch input.Type.String() {
	case "breakfast":
		if breakfast < input.Number {
			return false, errors.New("没有可用早餐卷")
		}
	case "lunch":
		if lunch < input.Number {
			return false, errors.New("没有可用午餐卷")
		}
	case "dinner":
		if dinner < input.Number {
			return false, errors.New("没有可用晚餐卷")
		}
	}

	booking := &models.Booking{}
	booking.UserID = uint64(input.UserID)
	booking.CanteenID = uint64(input.CanteenID)
	booking.BookingType = input.Type.String()
	booking.BookingDate = input.Date
	booking.Number = input.Number
	booking.Available = true

	if bookings, _ := booking.Check(); len(bookings) > 0 {
		return false, errors.New("已有预订，您不需要再重复预订")
	}

	if _, err := booking.Create(); err != nil {
		return false, err
	}

	//预订即回收餐票
	if err := models.RecyclingTickets(booking.UserID, models.TicketTypeMap[booking.BookingType], input.Number); err != nil {
		models.DeleteBooking(booking.ID)
		return false, err
	}

	return true, nil

}
