package resolvers

import (
	"booking"
	"booking/models"
	"booking/util"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/xid"
	"github.com/skip2/go-qrcode"
	"time"
)

type canteenResolver struct{ *Resolver }

func (r *canteenResolver) ID(ctx context.Context, obj *models.Canteen) (string, error) {
	return fmt.Sprintf("%d", obj.ID), nil
}

func (r *canteenResolver) GroupID(ctx context.Context, obj *models.Canteen) (int, error) {
	return int(obj.GroupID), nil
}

func (r *canteenResolver) CreatedAt(ctx context.Context, obj *models.Canteen) (string, error) {
	return fmt.Sprintf(obj.CreatedAt.Format("2006-01-02 15:04:05")), nil
}

func (r *canteenResolver) UpdatedAt(ctx context.Context, obj *models.Canteen) (string, error) {
	return fmt.Sprintf(obj.UpdatedAt.Format("2006-01-02 15:04:05")), nil
}

func (r *canteenResolver) DeletedAt(ctx context.Context, obj *models.Canteen) (string, error) {
	if obj.DeletedAt != nil {
		return fmt.Sprintf(obj.DeletedAt.Format("2006-01-02 15:04:05")), nil
	}

	return "", nil
}

func (r *canteenResolver) Count(ctx context.Context, obj *models.Canteen) ([]booking.CanteenCount, error) {
	counts := make([]booking.CanteenCount, 0)
	result, err := models.CountBookingByCanteen(obj.ID)

	for date, data := range result {
		count := booking.CanteenCount{}
		count.Date = date
		count.Breakfast = data["breakfast"]
		count.Lunch = data["lunch"]
		count.Dinner = data["dinner"]
		counts = append(counts, count)
	}

	return counts, err
}

func (r *canteenResolver) Admin(ctx context.Context, obj *models.Canteen) (models.User, error) {
	user, err := models.GetUserByID(obj.AdminID)

	if err != nil {
		return user, errors.New("admin cannot found")
	}

	return user, nil
}

func (r *mutationResolver) CreateCanteens(ctx context.Context, input booking.NewCanteen) (canteen models.Canteen, err error) {
	c := models.Canteen{
		Name:                     input.Name,
		GroupID:                  uint64(input.GroupID),
		BreakfastTime:            input.BreakfastTime,
		LunchTime:                input.LunchTime,
		DinnerTime:               input.DinnerTime,
		BookingBreakfastDeadline: input.BookingBreakfastDeadline,
		BookingLunchDeadline:     input.BookingLunchDeadline,
		BookingDinnerDeadline:    input.BookingDinnerDeadline,
		CancelTime:               input.CancelTime,
		AdminID:                  uint64(input.AdminID),
	}

	if input.BreakfastPicture != nil {
		if *input.BreakfastPicture == "" {
			c.BreakfastPicture = "assets/ticket_default.png"
		} else {
			c.BreakfastPicture = *input.BreakfastPicture
		}
	}

	if input.LunchPicture != nil {
		if *input.LunchPicture == "" {
			c.LunchPicture = "assets/ticket_default.png"
		} else {
			c.LunchPicture = *input.LunchPicture
		}
	}

	if input.DinnerPicture != nil {
		if *input.DinnerPicture == "" {
			c.DinnerPicture = "assets/ticket_default.png"
		} else {
			c.DinnerPicture = *input.DinnerPicture
		}
	}

	// Insert the canteen to the database.
	if _, err := c.Create(); err != nil {
		return c, err
	}

	return c, nil
}

func (r *mutationResolver) UpdateCanteens(ctx context.Context, input booking.UpdateCanteenInput) (models.Canteen, error) {
	c := models.Canteen{}

	data := make(map[string]interface{})

	if input.Name != nil {
		c.Name = *input.Name
		data["name"] = *input.Name
	}

	if input.GroupID != nil {
		c.GroupID = uint64(*input.GroupID)
		data["group_id"] = *input.GroupID
	}

	if input.AdminID != nil {
		c.AdminID = uint64(*input.AdminID)
		data["admin_id"] = *input.AdminID
	}

	if input.BreakfastTime != nil {
		c.BreakfastTime = *input.BreakfastTime
		data["breakfast_time"] = *input.BreakfastTime
	}

	if input.LunchTime != nil {
		c.LunchTime = *input.LunchTime
		data["lunch_time"] = *input.LunchTime
	}

	if input.DinnerTime != nil {
		c.DinnerTime = *input.DinnerTime
		data["dinner_time"] = *input.DinnerTime
	}

	if input.BreakfastPicture != nil {
		c.BreakfastTime = *input.BreakfastPicture
		data["breakfast_picture"] = *input.BreakfastPicture
	}

	if input.LunchTime != nil {
		c.LunchTime = *input.LunchTime
		data["lunch_picture"] = *input.LunchTime
	}

	if input.DinnerTime != nil {
		c.DinnerTime = *input.DinnerTime
		data["dinner_picture"] = *input.DinnerTime
	}

	if input.BookingBreakfastDeadline != nil {
		c.BookingBreakfastDeadline = *input.BookingBreakfastDeadline
		data["booking_breakfast_deadline"] = *input.BookingBreakfastDeadline
	}

	if input.BookingLunchDeadline != nil {
		c.BookingLunchDeadline = *input.BookingLunchDeadline
		data["booking_lunch_deadline"] = *input.BookingLunchDeadline
	}

	if input.BookingDinnerDeadline != nil {
		c.BookingDinnerDeadline = *input.BookingDinnerDeadline
		data["booking_dinner_deadline"] = *input.BookingDinnerDeadline
	}

	if input.CancelTime != nil {
		c.CancelTime = *input.CancelTime
		data["cancel_time"] = *input.CancelTime
	}
	c.ID = uint64(input.ID)

	fmt.Println("canteen update ", util.PrettyJson(input))
	// Insert the group to the database.
	err := c.Update(data)
	if err != nil {
		return c, err
	}
	return c, nil
}

func (r *mutationResolver) DeleteCanteens(ctx context.Context, input booking.DeleteIDInput) (bool, error) {
	if len(input.Ids) > 0 {
		id := uint64(input.Ids[0])
		if err := models.DeleteCanteen(id); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (r *mutationResolver) CreateQrcode(ctx context.Context, input booking.CanteenQrcodeInput) (string, error) {
	canteen, err := models.GetCanteen(uint64(input.ID))
	if err != nil {
		return "", err
	}

	path := "/download/qrcode/" + canteen.Name + "qrcode_" + time.Now().String() + ".png"

	canteen.Qrcode = path
	data := make(map[string]interface{})

	data["qrcode"] = path
	data["qrcode_uuid"] = xid.New().String()

	str := "BookingConfirm:true;id:" + util.Uint2Str(canteen.ID) + ";name:" + canteen.Name + ";date:" + time.Now().String() + ";qrcode_uuid:" + data["qrcode_uuid"].(string) + ";"
	err = qrcode.WriteFile(str, qrcode.Medium, 256, "."+path)

	if err != nil {
		return "", err
	}

	if err = canteen.Update(data); err != nil {
		return "", err
	}

	return path, nil
}
