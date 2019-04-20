package resolvers

import (
	"booking/models"
	"booking/util"
	"context"
	"booking"
	"fmt"
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

type bookingResolver struct{ *Resolver }

func (r *bookingResolver) ID(ctx context.Context, obj *models.Booking) (string, error) {
	return fmt.Sprintf("%d",obj.ID), nil
}


func (r *bookingResolver) DeletedAt(ctx context.Context, obj *models.Booking) (string, error){
	if obj.DeletedAt != nil {
		return fmt.Sprintf(obj.DeletedAt.Format("2006-01-02 15:04:05")), nil
	}

	return "", nil
}

func (r *bookingResolver) CanteenID(ctx context.Context, obj *models.Booking) (int, error){
	return int(obj.CanteenID), nil
}

func (r *bookingResolver) Type(ctx context.Context, obj *models.Booking) (string, error){
	return obj.BookingType, nil
}

func (r *bookingResolver) Date(ctx context.Context, obj *models.Booking) (string, error){
	return obj.BookingDate, nil
}

func (r *bookingResolver) UserID(ctx context.Context, obj *models.Booking) (int, error){
	return int(obj.UserID), nil
}

func (r *bookingResolver) CreatedAt(ctx context.Context, obj *models.Booking) (string, error){
	return fmt.Sprintf(obj.CreatedAt.Format("2006-01-02 15:04:05")), nil
}

func (r *bookingResolver) UpdatedAt(ctx context.Context, obj *models.Booking) (string, error){
	return fmt.Sprintf(obj.UpdatedAt.Format("2006-01-02 15:04:05")), nil
}



func (r *mutationResolver) CancelBooking(ctx context.Context, input booking.CancelBookingInput) (bool, error){
	err := models.CancelBooking(uint64(input.UserID), uint64(input.BookingID))

	if err != nil {
		return false, err
	}

	return true , nil
}

func (r *mutationResolver)  Booking(ctx context.Context, input booking.BookingInput) (bool, error){
	date := strings.Split(input.Date, "-")

	errs := []string{}

	//自动预订一整个月份
	if input.AutoCurrentMonth != nil {
		year,_ := strconv.Atoi(date[0])
		month,_ := strconv.Atoi(date[1])
		days := util.CountDays(year,month)

		for i:=1;i< days;i++{
			booking := &models.Booking{}
			booking.UserID = uint64(input.UserID)
			booking.CanteenID = uint64(input.CanteenID)
			booking.BookingType = input.Type.String()
			booking.BookingDate = date[0] + "-" +  date[1] + "-" + strconv.Itoa(i)

			if _, err := booking.Create(); err != nil {
				errs = append(errs, err.Error())
			}

		}

		if len(errs) > 0 {
			return false, errors.New(strings.Join(errs, "|"))
		}
	}

	booking := &models.Booking{}
	booking.UserID = uint64(input.UserID)
	booking.CanteenID = uint64(input.CanteenID)
	booking.BookingType = input.Type.String()
	booking.BookingDate = input.Date

	if _, err := booking.Create(); err != nil {
		return false , err
	}

	return true , nil

}
