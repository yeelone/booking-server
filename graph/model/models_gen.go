// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"booking/models"
	"fmt"
	"io"
	"strconv"
)

type BookingExportResponses struct {
	Data []CanteenBookingExport `json:"data"`
	File string                 `json:"file"`
}

type BookingFilterInput struct {
	UserID    *int `json:"userId"`
	CanteenID *int `json:"canteenId"`
}

type BookingInput struct {
	UserID           int              `json:"userId"`
	CanteenID        int              `json:"canteenId"`
	Type             BookingTypeInput `json:"type"`
	Number           int              `json:"number"`
	Date             string           `json:"date"`
	AutoCurrentMonth *bool            `json:"autoCurrentMonth"`
}

type CanteenBookingExport struct {
	Username  string `json:"username"`
	Breakfast int    `json:"breakfast"`
	Lunch     int    `json:"lunch"`
	Dinner    int    `json:"dinner"`
}

type CanteenCount struct {
	Date      string `json:"date"`
	Breakfast int    `json:"breakfast"`
	Lunch     int    `json:"lunch"`
	Dinner    int    `json:"dinner"`
}

type CanteenFilterInput struct {
	ID      *int    `json:"id"`
	Name    *string `json:"name"`
	GroupID *int    `json:"groupID"`
	AdminID *int    `json:"adminID"`
}

type CanteenQrcodeInput struct {
	ID int `json:"id"`
}

type ClientConfig struct {
	WxAppID  *string `json:"wxAppID"`
	Prompt   *string `json:"prompt"`
	WxSecret *string `json:"wxSecret"`
}

type ConfigInput struct {
	Prompt   *string `json:"prompt"`
	WxAppID  *string `json:"wxAppID"`
	WxSecret *string `json:"wxSecret"`
}

type Count struct {
	Breakfast int `json:"breakfast"`
	Lunch     int `json:"lunch"`
	Dinner    int `json:"dinner"`
}

type CreateUsersResponse struct {
	Errors []string `json:"errors"`
}

type DashboardResponse struct {
	OrgInfo    []OrgDashboard `json:"orgInfo"`
	SystemInfo *SystemInfo    `json:"systemInfo"`
	TicketInfo []string       `json:"ticketInfo"`
}

type Data struct {
	Used    int  `json:"used"`
	Total   int  `json:"total"`
	Percent *int `json:"percent"`
}

type DeleteIDInput struct {
	Ids []int `json:"ids"`
}

type DishesFilterInput struct {
	ID   *int    `json:"id"`
	Name *string `json:"name"`
}

type GenarateTicketInput struct {
	Number int `json:"number"`
	UserID int `json:"userId"`
	Type   int `json:"type"`
	Price  int `json:"price"`
}

type GroupFilterInput struct {
	ID   *int    `json:"id"`
	Name *string `json:"name"`
}

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token       string       `json:"token"`
	Permissions []string     `json:"permissions"`
	User        *models.User `json:"user"`
}

type LogoutInput struct {
	Username string `json:"username"`
}

type NewCanteen struct {
	ID                       *int    `json:"id"`
	Name                     string  `json:"name"`
	GroupID                  int     `json:"groupID"`
	BreakfastTime            string  `json:"breakfastTime"`
	BreakfastPicture         *string `json:"breakfastPicture"`
	BookingBreakfastDeadline string  `json:"bookingBreakfastDeadline"`
	LunchTime                string  `json:"lunchTime"`
	LunchPicture             *string `json:"lunchPicture"`
	BookingLunchDeadline     string  `json:"bookingLunchDeadline"`
	DinnerTime               string  `json:"dinnerTime"`
	DinnerPicture            *string `json:"dinnerPicture"`
	BookingDinnerDeadline    string  `json:"bookingDinnerDeadline"`
	CancelTime               int     `json:"cancelTime"`
	AdminID                  int     `json:"adminId"`
}

type NewComment struct {
	UserID int    `json:"userId"`
	Body   string `json:"body"`
	Tunnel string `json:"tunnel"`
}

type NewDishes struct {
	Name    string `json:"Name"`
	Picture string `json:"Picture"`
}

type NewGroup struct {
	ID      *int   `json:"id"`
	Name    string `json:"name"`
	Admin   int    `json:"admin"`
	Parent  int    `json:"parent"`
	Picture string `json:"picture"`
	UserID  []int  `json:"userId"`
}

type NewRole struct {
	ID     *int   `json:"id"`
	Name   string `json:"name"`
	UserID []int  `json:"userId"`
}

type NewTodo struct {
	Text   string `json:"text"`
	UserID string `json:"userId"`
}

type NewUser struct {
	ID       *int    `json:"id"`
	Email    string  `json:"email"`
	Password string  `json:"password"`
	Username string  `json:"username"`
	Nickname *string `json:"nickname"`
	IDCard   *string `json:"id_card"`
	IsSuper  *bool   `json:"is_super"`
	Picture  *string `json:"picture"`
	State    *int    `json:"state"`
	GroupID  *int    `json:"groupId"`
}

type NewUsers struct {
	UploadFile string `json:"uploadFile"`
	GroupID    int    `json:"groupId"`
}

type OrgDashboard struct {
	Name         string `json:"name"`
	UserCount    int    `json:"userCount"`
	CanteenCount int    `json:"canteenCount"`
}

type Pagination struct {
	Skip int `json:"skip"`
	Take int `json:"take"`
}

type Permission struct {
	Module    string `json:"module"`
	Name      string `json:"name"`
	Resource  string `json:"resource"`
	Object    string `json:"object"`
	Checked   bool   `json:"checked"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	DeletedAt string `json:"deletedAt"`
}

type QueryBookingResponse struct {
	TotalCount *int             `json:"totalCount"`
	Skip       *int             `json:"skip"`
	Take       *int             `json:"take"`
	Rows       []models.Booking `json:"rows"`
}

type QueryCanteenResponse struct {
	TotalCount *int             `json:"totalCount"`
	Skip       *int             `json:"skip"`
	Take       *int             `json:"take"`
	Rows       []models.Canteen `json:"rows"`
}

type QueryCommentResponse struct {
	TotalCount *int             `json:"totalCount"`
	Skip       *int             `json:"skip"`
	Take       *int             `json:"take"`
	Rows       []models.Comment `json:"rows"`
}

type QueryDishesResponse struct {
	TotalCount *int            `json:"totalCount"`
	Skip       *int            `json:"skip"`
	Take       *int            `json:"take"`
	Rows       []models.Dishes `json:"rows"`
}

type QueryGroupResponse struct {
	TotalCount *int           `json:"totalCount"`
	Skip       *int           `json:"skip"`
	Take       *int           `json:"take"`
	Rows       []models.Group `json:"rows"`
}

type QueryPermissionResponse struct {
	TotalCount *int         `json:"totalCount"`
	Skip       *int         `json:"skip"`
	Take       *int         `json:"take"`
	Rows       []Permission `json:"rows"`
}

type QueryRoleResponse struct {
	TotalCount *int          `json:"totalCount"`
	Skip       *int          `json:"skip"`
	Take       *int          `json:"take"`
	Rows       []models.Role `json:"rows"`
}

type QueryTicketRecordResponse struct {
	TotalCount *int                  `json:"totalCount"`
	Skip       *int                  `json:"skip"`
	Take       *int                  `json:"take"`
	Rows       []models.TicketRecord `json:"rows"`
}

type QueryTicketResponse struct {
	TotalCount *int            `json:"totalCount"`
	Skip       *int            `json:"skip"`
	Take       *int            `json:"take"`
	Count      *Count          `json:"count"`
	Rows       []models.Ticket `json:"rows"`
}

type QueryUserResponse struct {
	TotalCount *int          `json:"totalCount"`
	Skip       *int          `json:"skip"`
	Take       *int          `json:"take"`
	Rows       []models.User `json:"rows"`
}

type RecyclingTicketsInput struct {
	Number int `json:"number"`
	UserID int `json:"userId"`
	Type   int `json:"type"`
}

type ResetPasword struct {
	Ids []int `json:"ids"`
}

type RoleAndPermissionRelationshipInput struct {
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}

type RoleAndUserFilterInput struct {
	RoleID  int   `json:"roleId"`
	UserIds []int `json:"userIds"`
}

type RoleFilterInput struct {
	ID   *int    `json:"id"`
	Name *string `json:"name"`
}

type SpendInput struct {
	CanteenID int    `json:"canteenId"`
	UserID    int    `json:"userId"`
	UUID      string `json:"uuid"`
}

type SystemInfo struct {
	CurrentLoginCount int     `json:"currentLoginCount"`
	CPU               *string `json:"cpu"`
	Disk              *string `json:"disk"`
	RAM               *string `json:"ram"`
}

type TicketFilterInput struct {
	ID     *int    `json:"id"`
	UserID *int    `json:"userId"`
	UUID   *string `json:"uuid"`
	Count  *bool   `json:"count"`
}

type TicketRecordFilterInput struct {
	Operator *int `json:"operator"`
	Owner    int  `json:"owner"`
}

type TransferResponse struct {
	SuccessCount int     `json:"successCount"`
	ErrorCount   int     `json:"errorCount"`
	ErrorMsg     *string `json:"errorMsg"`
}

type TransferTicketInput struct {
	Number     int    `json:"number"`
	Type       string `json:"type"`
	FromUserID int    `json:"fromUserId"`
	ToUserID   int    `json:"toUserId"`
}

type UpdateCanteenInput struct {
	ID                       int     `json:"id"`
	Name                     *string `json:"name"`
	GroupID                  *int    `json:"groupID"`
	BreakfastTime            *string `json:"breakfastTime"`
	BreakfastPicture         *string `json:"breakfastPicture"`
	BookingBreakfastDeadline *string `json:"bookingBreakfastDeadline"`
	LunchTime                *string `json:"lunchTime"`
	LunchPicture             *string `json:"lunchPicture"`
	BookingLunchDeadline     *string `json:"bookingLunchDeadline"`
	DinnerTime               *string `json:"dinnerTime"`
	DinnerPicture            *string `json:"dinnerPicture"`
	BookingDinnerDeadline    *string `json:"bookingDinnerDeadline"`
	CancelTime               *int    `json:"cancelTime"`
	AdminID                  *int    `json:"adminId"`
}

type UpdateDishesInput struct {
	ID      int     `json:"id"`
	Name    *string `json:"name"`
	Picture *string `json:"picture"`
}

type UpdateGroupInput struct {
	ID      int     `json:"id"`
	Name    *string `json:"name"`
	Admin   *int    `json:"admin"`
	Parent  *int    `json:"parent"`
	Picture *string `json:"picture"`
	Levels  *string `json:"levels"`
	UserID  []int   `json:"userId"`
}

type UpdateRoleInput struct {
	ID     int     `json:"id"`
	Name   *string `json:"name"`
	UserID []int   `json:"userId"`
}

type UpdateUserInput struct {
	ID          int     `json:"id"`
	Email       *string `json:"email"`
	Password    *string `json:"password"`
	Username    *string `json:"username"`
	Nickname    *string `json:"nickname"`
	IDCard      *string `json:"id_card"`
	IsSuper     *bool   `json:"is_super"`
	Picture     *string `json:"picture"`
	State       *int    `json:"state"`
	ReGenQrcode *bool   `json:"re_gen_qrcode"`
}

type UserAndGroupRelationshipInput struct {
	UserIds []int `json:"userIds"`
	GroupID int   `json:"groupId"`
}

type UserAndRoleRelationshipInput struct {
	UserIds []int `json:"userIds"`
	RoleID  int   `json:"roleId"`
}

type UserFilterInput struct {
	ID       *int    `json:"id"`
	Username *string `json:"username"`
	Email    *string `json:"email"`
	Phone    *string `json:"phone"`
	State    *int    `json:"state"`
}

type CancelBookingInput struct {
	UserID    int `json:"userId"`
	BookingID int `json:"bookingId"`
}

type BookingTypeInput string

const (
	BookingTypeInputBreakfast BookingTypeInput = "breakfast"
	BookingTypeInputLunch     BookingTypeInput = "lunch"
	BookingTypeInputDinner    BookingTypeInput = "dinner"
)

var AllBookingTypeInput = []BookingTypeInput{
	BookingTypeInputBreakfast,
	BookingTypeInputLunch,
	BookingTypeInputDinner,
}

func (e BookingTypeInput) IsValid() bool {
	switch e {
	case BookingTypeInputBreakfast, BookingTypeInputLunch, BookingTypeInputDinner:
		return true
	}
	return false
}

func (e BookingTypeInput) String() string {
	return string(e)
}

func (e *BookingTypeInput) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = BookingTypeInput(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid BookingTypeInput", str)
	}
	return nil
}

func (e BookingTypeInput) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type GroupOrderByInput string

const (
	GroupOrderByInputNameAsc       GroupOrderByInput = "name_ASC"
	GroupOrderByInputNameDesc      GroupOrderByInput = "name_DESC"
	GroupOrderByInputCreatedAtAsc  GroupOrderByInput = "created_at_ASC"
	GroupOrderByInputCreatedAtDesc GroupOrderByInput = "created_at_DESC"
)

var AllGroupOrderByInput = []GroupOrderByInput{
	GroupOrderByInputNameAsc,
	GroupOrderByInputNameDesc,
	GroupOrderByInputCreatedAtAsc,
	GroupOrderByInputCreatedAtDesc,
}

func (e GroupOrderByInput) IsValid() bool {
	switch e {
	case GroupOrderByInputNameAsc, GroupOrderByInputNameDesc, GroupOrderByInputCreatedAtAsc, GroupOrderByInputCreatedAtDesc:
		return true
	}
	return false
}

func (e GroupOrderByInput) String() string {
	return string(e)
}

func (e *GroupOrderByInput) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = GroupOrderByInput(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid GroupOrderByInput", str)
	}
	return nil
}

func (e GroupOrderByInput) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
