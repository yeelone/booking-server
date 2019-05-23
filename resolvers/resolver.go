package resolvers

import (
	"booking"
	g "booking"
	"booking/models"
	"booking/pkg/auth"
	"booking/pkg/constvar"
	"booking/pkg/permission"
	"booking/pkg/sd"
	"booking/pkg/token"
	"booking/util"
	"context"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"os"
	"strconv"
	"strings"
	"time"
)

type Tunnel struct {
	Name      string
	Observers map[string]chan models.Message
}

type CommentTunnel struct {
	Name      string
	Observers map[string]chan models.Comment
}

type Resolver struct {
	tunnels  map[string]*Tunnel
	commentTunnels map[string]*CommentTunnel
	groups   []models.Group
	users    []models.User
	roles    []models.Role
	tickets  []models.Ticket
	dishes   []models.Dishes
	canteens []models.Canteen
	bookings []models.Booking
}

func (r *Resolver) Mutation() g.MutationResolver {
	return &mutationResolver{r}
}

func (r *Resolver) Query() g.QueryResolver {
	return &queryResolver{r}
}

func (r *Resolver) Group() g.GroupResolver {
	return &groupResolver{r}
}

func (r *Resolver) User() g.UserResolver {
	return &userResolver{r}
}

func (r *Resolver) Comment() g.CommentResolver {
	return &commentResolver{r}
}
func (r *Resolver) Role() g.RoleResolver {
	return &roleResolver{r}
}

func (r *Resolver) Ticket() g.TicketResolver {
	return &ticketsResolver{r}
}

func (r *Resolver) Dishes() g.DishesResolver {
	return &dishesResolver{r}
}

func (r *Resolver) Canteen() g.CanteenResolver {
	return &canteenResolver{r}
}

func (r *Resolver) TicketRecord() g.TicketRecordResolver {
	return &ticketRecordResolver{r}
}

func (r *Resolver) Booking() g.BookingResolver {
	return &bookingResolver{r}
}

// Subscription returns a subscription resolver
func (r *Resolver) Subscription() g.SubscriptionResolver {
	return &subscriptionResolver{r}
}

func (r *Resolver) Message() g.MessageResolver {
	return &messageResolver{r}
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) Login(ctx context.Context, input booking.LoginInput) (booking.LoginResponse, error) {
	resp := booking.LoginResponse{}

	// Get the user information by the login username.
	d, err := models.GetUserByName(input.Username)
	if err != nil {
		return resp, err
	}

	// Compare the login password with the user password.
	if err := auth.Compare(d.Password, input.Password); err != nil {
		return resp, err
	}

	role := ""
	if len(d.Roles) > 0 {
		role = d.Roles[0].Name
	}

	// Sign the json web token.
	t, err := token.Sign(token.Context{ID: d.ID, Username: d.Username, Role: role}, "")
	if err != nil {
		return resp, err
	}

	err = models.DB.Cache.Update(func(tx *bolt.Tx) error {
		// 这里还有另外一层：k-v存储在bucket中，
		// 可以将bucket当做一个key的集合或者是数据库中的表。
		//（顺便提一句，buckets中可以包含其他的buckets，这将会相当有用）
		// Buckets 是键值对在数据库中的集合.所有在bucket中的key必须唯一，
		// 使用DB.CreateBucket() 函数建立buket
		//Tx.DeleteBucket() 删除bucket
		//b := tx.Bucket([]byte("MyBucket"))
		b, err := tx.CreateBucketIfNotExists([]byte(models.Login_Record_BoltDB_Key))
		//要将 key/value 对保存到 bucket，请使用 Bucket.Put() 函数：
		//这将在 MyBucket 的 bucket 中将 "answer" key的值设置为"42"。
		id := util.Uint2Str(d.ID)
		client_ip := ctx.Value(models.CLIENT_IP).(string)
		client_date := time.Now().UTC().String()
		key := []byte(id)
		value := []byte("username:" + d.Username + " ip:" + client_ip + " date:" + client_date)
		err = b.Put(key, value)
		return err
	})

	perms := getPermissionAccessResourceByRole(role)
	resp.Token = t
	resp.Permissions = perms
	resp.User = *d
	return resp, nil
}

func (r *mutationResolver) Logout(ctx context.Context, input booking.LogoutInput) (bool, error) {

	// Get the user information by the login username.
	d, err := models.GetUserByName(input.Username)
	if err != nil {
		return false, err
	}

	err = models.DB.Cache.Update(func(tx *bolt.Tx) error {
		// 这里还有另外一层：k-v存储在bucket中，
		// 可以将bucket当做一个key的集合或者是数据库中的表。
		//（顺便提一句，buckets中可以包含其他的buckets，这将会相当有用）
		// Buckets 是键值对在数据库中的集合.所有在bucket中的key必须唯一，
		// 使用DB.CreateBucket() 函数建立buket
		//Tx.DeleteBucket() 删除bucket
		//b := tx.Bucket([]byte("MyBucket"))
		b, err := tx.CreateBucketIfNotExists([]byte(models.Login_Record_BoltDB_Key))
		//要将 key/value 对保存到 bucket，请使用 Bucket.Put() 函数：
		//这将在 MyBucket 的 bucket 中将 "answer" key的值设置为"42"。
		id := util.Uint2Str(d.ID)
		key := []byte(id)
		err = b.Delete(key)
		return err
	})

	return true, err
}

func getPermissionAccessResourceByRole(role string) (perms []string) {
	runtimeViper := viper.New()
	runtimeViper.AddConfigPath("conf/permissions") // 如果没有指定配置文件，则解析默认的配置文件
	runtimeViper.SetConfigName("permission")

	runtimeViper.SetConfigType("yaml")                  // 设置配置文件格式为YAML
	if err := runtimeViper.ReadInConfig(); err != nil { // viper解析配置文件
		return perms
	}

	if role == "" {
		return perms
	}

	checkedMap := permission.GetRolePermissionFromCSVFile()
	existed := make(map[string]bool)

	for _, key := range runtimeViper.AllKeys() {
		s := strings.Split(key, ".")
		if len(s) >= 2 {
			k := s[0] + "." + s[1]

			if _, ok := existed[k]; ok {
				continue
			}
			existed[k] = true
			object := runtimeViper.GetString(s[0] + "." + s[1] + ".object")
			if _, ok := checkedMap[role][object]; ok {
				fmt.Println(runtimeViper.GetString(s[0] + "." + s[1] + ".resource"))
				perms = append(perms, runtimeViper.GetString(s[0]+"."+s[1]+".resource"))
			}
		}
	}

	return perms
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Groups(ctx context.Context, filter *booking.GroupFilterInput, pagination *booking.Pagination, orderBy *booking.GroupOrderByInput) (booking.QueryGroupResponse, error) {
	if pagination == nil {
		pagination = &booking.Pagination{
			Skip: 0,
			Take: constvar.DefaultLimit,
		}
	}

	where := ""
	whereValue := ""
	if filter != nil {
		if filter.Name != nil && *filter.Name != "" {
			where = "name"
			whereValue = *filter.Name
		}

		if filter.ID != nil && *filter.ID != 0 {
			where = "id"
			whereValue = strconv.Itoa(*filter.ID)
		}
	}

	order := ""

	if orderBy != nil {
		order = orderBy.String()
	}

	groups, total, err := models.GetGroups(where, whereValue, pagination.Skip, pagination.Take, order)

	resp := booking.QueryGroupResponse{
		TotalCount: &total,
		Skip:       &pagination.Skip,
		Take:       &pagination.Take,
		Rows:       groups,
	}
	return resp, err
}

func (r *queryResolver) Users(ctx context.Context, filter *booking.UserFilterInput, pagination *booking.Pagination) (booking.QueryUserResponse, error) {

	if pagination == nil {
		pagination = &booking.Pagination{
			Skip: 0,
			Take: constvar.DefaultLimit,
		}
	}

	where := ""
	whereValue := ""
	if filter != nil {
		if filter.Username != nil {
			where = "username"
			whereValue = *filter.Username
		}
		if filter.Email != nil {
			where = "email"
			whereValue = *filter.Email
		}
	}
	users, total, err := models.GetUsers(where, whereValue, pagination.Skip, pagination.Take)
	resp := booking.QueryUserResponse{
		Rows:       users,
		Skip:       &pagination.Skip,
		Take:       &pagination.Take,
		TotalCount: &total,
	}
	return resp, err
}

func (r *queryResolver) Comments(ctx context.Context, pagination *booking.Pagination) (booking.QueryCommentResponse, error) {

	if pagination == nil {
		pagination = &booking.Pagination{
			Skip: 0,
			Take: constvar.DefaultLimit,
		}
	}

	comments, total, err := models.GetComments("", "", pagination.Skip, pagination.Take)
	resp := booking.QueryCommentResponse{
		Rows:       comments,
		Skip:       &pagination.Skip,
		Take:       &pagination.Take,
		TotalCount: &total,
	}
	return resp, err
}

func (r *queryResolver) Roles(ctx context.Context, filter *booking.RoleFilterInput, pagination *booking.Pagination) (booking.QueryRoleResponse, error) {
	if pagination == nil {
		pagination = &booking.Pagination{
			Skip: 0,
			Take: constvar.DefaultLimit,
		}
	}

	where := ""
	whereValue := ""
	if filter != nil {
		if filter.Name != nil {
			where = "name"
			whereValue = *filter.Name
		}

		if filter.ID != nil {
			where = "id"
			whereValue = strconv.Itoa(*filter.ID)
		}

	}

	roles, total, err := models.GetRoles(where, whereValue, pagination.Skip, pagination.Take)
	resp := booking.QueryRoleResponse{
		Rows:       roles,
		Skip:       &pagination.Skip,
		Take:       &pagination.Take,
		TotalCount: &total,
	}

	return resp, err
}

func (r *queryResolver) Dishes(ctx context.Context, filter *booking.DishesFilterInput, pagination *booking.Pagination) (booking.QueryDishesResponse, error) {
	if pagination == nil {
		pagination = &booking.Pagination{
			Skip: 0,
			Take: constvar.DefaultLimit,
		}
	}

	where := ""
	whereValue := ""
	if filter != nil {
		if filter.Name != nil {
			where = "name"
			whereValue = *filter.Name
		}
	}

	dishes, total, err := models.GetDishes(where, whereValue, pagination.Skip, pagination.Take)
	resp := booking.QueryDishesResponse{
		Rows:       dishes,
		Skip:       &pagination.Skip,
		Take:       &pagination.Take,
		TotalCount: &total,
	}

	return resp, err
}

func (r *queryResolver) Tickets(ctx context.Context, filter *booking.TicketFilterInput, pagination *booking.Pagination) (booking.QueryTicketResponse, error) {
	if pagination == nil {
		pagination = &booking.Pagination{
			Skip: 0,
			Take: constvar.DefaultLimit,
		}
	}
	var err error
	count := booking.Count{}
	if filter != nil {
		if filter.Count != nil { // 只计算余票数量
			count.Breakfast, count.Lunch, count.Dinner, err = models.CountTicketsDetailByUser(uint64(*filter.UserID))
			resp := booking.QueryTicketResponse{
				Count: &count,
			}

			return resp, err
		}
	}

	where := ""
	whereValue := ""
	if filter != nil {
		if filter.UserID != nil {
			where = "user_id"
			whereValue = strconv.Itoa(*filter.UserID)
		}
		if filter.UUID != nil {
			where = "uuid"
			whereValue = *filter.UUID
		}
	}

	tickets, total, err := models.GetTickets(where, whereValue, pagination.Skip, pagination.Take)
	resp := booking.QueryTicketResponse{
		Rows:       tickets,
		Skip:       &pagination.Skip,
		Take:       &pagination.Take,
		TotalCount: &total,
	}

	return resp, err
}

func (r *queryResolver) TicketRecords(ctx context.Context, filter *booking.TicketRecordFilterInput, pagination *booking.Pagination) (booking.QueryTicketRecordResponse, error) {
	if pagination == nil {
		pagination = &booking.Pagination{
			Skip: 0,
			Take: constvar.DefaultLimit,
		}
	}

	where := ""
	whereValue := ""

	if filter != nil {
		where = "owner"
		whereValue = strconv.Itoa(filter.Owner)
	}

	rs, total, err := models.GetTicketRecords(where, whereValue, pagination.Skip, pagination.Take)
	resp := booking.QueryTicketRecordResponse{
		Rows:       rs,
		Skip:       &pagination.Skip,
		Take:       &pagination.Take,
		TotalCount: &total,
	}

	return resp, err
}

func (r *queryResolver) Canteens(ctx context.Context, filter *booking.CanteenFilterInput, pagination *booking.Pagination) (booking.QueryCanteenResponse, error) {
	if pagination == nil {
		pagination = &booking.Pagination{
			Skip: 0,
			Take: constvar.DefaultLimit,
		}
	}

	where := ""
	whereValue := ""
	if filter != nil {
		if filter.Name != nil {
			where = "name"
			whereValue = *filter.Name
		}

		if filter.GroupID != nil {
			where = "group_id"
			whereValue = strconv.Itoa(*filter.GroupID)
		}

		if filter.AdminID != nil {
			where = "admin_id"
			whereValue = strconv.Itoa(*filter.AdminID)
		}

		if filter.ID != nil {
			where = "id"
			whereValue = strconv.Itoa(*filter.ID)
		}

	}

	cs, total, err := models.GetCanteens(where, whereValue, pagination.Skip, pagination.Take, "")
	resp := booking.QueryCanteenResponse{
		TotalCount: &total,
		Skip:       &pagination.Skip,
		Take:       &pagination.Take,
		Rows:       cs,
	}
	return resp, err
}

func (r *queryResolver) Dashboard(ctx context.Context) (response booking.DashboardResponse, err error) {
	// 查看当前登录人数
	cpu := sd.CPUCheck()
	ram := sd.RAMCheck()
	disk := sd.DiskCheck()

	systemInfo := booking.SystemInfo{
		CPU: &cpu,
		Disk: &disk,
		RAM: &ram,
	}
	systemInfo.CurrentLoginCount = 0


	//只读事务在db.View函数之中：在函数中可以读取，但是不能做修改。
	err = models.DB.Cache.View(func(tx *bolt.Tx) error {
		//要检索这个value，我们可以使用 Bucket.Get() 函数：
		//由于Get是有安全保障的，所有不会返回错误，不存在的key返回nil
		b := tx.Bucket([]byte(models.Login_Record_BoltDB_Key))
		//tx.Bucket([]byte("MyBucket")).Cursor() 可这样写
		//游标遍历key

		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			//fmt.Printf("key=%s, value=%s\n", k, v)
			systemInfo.CurrentLoginCount += 1
		}
		return nil
	})

	//列出各机构人数
	countData := models.CountGroupUsers()

	orgInfo := make([]booking.OrgDashboard, 0)

	for name, count := range countData {
		d := booking.OrgDashboard{}
		d.Name = name
		d.UserCount = count
		orgInfo = append(orgInfo, d)
	}

	resp := booking.DashboardResponse{
		SystemInfo: systemInfo,
		OrgInfo:    orgInfo,
		TicketInfo: models.GetLatestTicketRecord(10),
	}
	return resp, err
}

func (r *queryResolver) Booking(ctx context.Context, filter *booking.BookingFilterInput) (response booking.QueryBookingResponse, err error) {

	where := ""
	whereValue := ""
	if filter != nil {
		if filter.UserID != nil {
			where = "user_id"
			whereValue = strconv.Itoa(*filter.UserID)
		}
		if filter.CanteenID != nil {
			where = "canteen_id"
			whereValue = strconv.Itoa(*filter.CanteenID)
		}
	}

	bookings, total, err := models.GetAllBooking(where, whereValue, 0, 10000)

	resp := booking.QueryBookingResponse{
		TotalCount: &total,
		Rows:       bookings,
	}
	return resp, err
}

func (r *queryResolver) Permissions(ctx context.Context, filter booking.RoleFilterInput) (resp booking.QueryPermissionResponse, err error) {
	runtimeViper := viper.New()
	runtimeViper.AddConfigPath("conf/permissions") // 如果没有指定配置文件，则解析默认的配置文件
	runtimeViper.SetConfigName("permission")

	runtimeViper.SetConfigType("yaml")                  // 设置配置文件格式为YAML
	if err := runtimeViper.ReadInConfig(); err != nil { // viper解析配置文件
		return resp, err
	}

	if filter.Name == nil {
		return resp, errors.New("必须指定角色名称")
	}

	role := *filter.Name

	checkedMap := permission.GetRolePermissionFromCSVFile()
	existed := make(map[string]bool)

	for _, key := range runtimeViper.AllKeys() {
		s := strings.Split(key, ".")
		if len(s) >= 2 {
			k := s[0] + "." + s[1]

			if _, ok := existed[k]; ok {
				continue
			}
			existed[k] = true
			object := runtimeViper.GetString(s[0] + "." + s[1] + ".object")

			r := booking.Permission{
				Module:   s[0],
				Name:     s[1],
				Resource: runtimeViper.GetString(s[0] + "." + s[1] + ".resource"),
				Object:   object,
			}
			if _, ok := checkedMap[role][object]; ok {
				r.Checked = true
			}
			resp.Rows = append(resp.Rows, r)
		}
	}
	total := len(resp.Rows)
	resp.TotalCount = &total
	return resp, nil
}

func (r *queryResolver) Messages(ctx context.Context) (string, error) {
	return "hello world", nil
}

func (r *mutationResolver) CreateRoleAndPermissionRelationship(ctx context.Context, input booking.RoleAndPermissionRelationshipInput) (bool, error) {
	// 已配置的权限策略
	permdMap := permission.GetRolePermissionFromCSVFile()

	// 先清除原来的配置
	permdMap[input.Role] = make(map[string]bool)

	for _, p := range input.Permissions {
		permdMap[input.Role][p] = true
	}

	err := permission.SavePermissionsToCSV(permdMap)
	if err != nil {
		return false, nil
	}

	auth.RefreshEnforcer()

	return true, nil
}

func (r *queryResolver) CheckUserNotInRole(ctx context.Context, filter *booking.RoleAndUserFilterInput) ([]int, error) {

	uids := make([]uint64, len(filter.UserIds))

	for i, id := range filter.UserIds {
		uids[i] = uint64(id)
	}

	ids, err := models.CheckUsersNotInRole(uint64(filter.RoleID), uids)

	newids := make([]int, 0)

	for _, id := range ids {
		newids = append(newids, int(id))
	}

	return newids, err
}

func (r *queryResolver) ExportBooking(ctx context.Context, year string, month string) (resps *booking.BookingExportResponses,err error){
	y, _ := strconv.Atoi(year)
	m, _ := strconv.Atoi(month)

	//获取到用户ID，查找用户所管理的食堂，用户只能导出所管理食堂的数据
	uid := ctx.Value("user_id").(uint64)

	canteens,total,err := models.GetCanteens("admin_id",util.Uint2Str(uid),0,1000,"")

	fmt.Println("total, err", uid, total, err )
	if err != nil {
		return resps, err
	}

	if total == 0 {
		return resps, nil
	}

	canteenIds := make([]uint64,total)

	for i,c := range canteens {
		canteenIds[i] = c.ID
	}

	data, err := models.CountBookingByMonth(y,m,canteenIds)
	if err != nil {
		return resps,err
	}

	resps = &booking.BookingExportResponses{}
	resps.Data = make([]booking.CanteenBookingExport,0)

	templateFile := "conf/export/booking.xlsx"

	if !util.Exists(templateFile){
		return resps, errors.New("导出模板文件不存在，请联系管理员 ")
	}

	var xlsx *excelize.File
	sheetName := "Sheet1"

	xlsx, err = excelize.OpenFile(templateFile)


	for i, d := range data {
		resp := booking.CanteenBookingExport{
			Username:d["username"].(string),
			Breakfast:d["breakfast"].(int),
			Lunch:d["lunch"].(int),
			Dinner:d["dinner"].(int),
		}

		resps.Data = append(resps.Data,resp)
		row := strconv.Itoa(i+2)
		xlsx.SetCellValue(sheetName, "A"+row,d["username"].(string) )
		xlsx.SetCellValue(sheetName, "B"+row,d["breakfast"].(int) )
		xlsx.SetCellValue(sheetName, "C"+row,d["lunch"].(int) )
		xlsx.SetCellValue(sheetName, "D"+row,d["dinner"].(int) )
	}
	xlsx.SetActiveSheet(xlsx.GetSheetIndex(sheetName))

	dir := "download/table/"
	filename := dir + "预订情况汇总表_" + year + "-" + month + ".xlsx"

	if !util.Exists(dir) {
		err := os.Mkdir(dir, os.ModePerm)
		if err != nil {
			return resps, err
		}
	}


	err = xlsx.SaveAs(filename)

	if err != nil {
		return resps, err
	}
	resps.File = filename
	return resps, nil

}