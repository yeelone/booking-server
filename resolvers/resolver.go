package resolvers

import (
	"booking"
	g "booking"
	"booking/models"
	"booking/pkg/auth"
	"booking/pkg/constvar"
	"booking/pkg/permission"
	"booking/pkg/token"
	"booking/service"
	"context"
	"fmt"
	"github.com/spf13/viper"
	"strings"
)

type Resolver struct {
	groups       []models.Group
	users        []models.User
	roles        []models.Role
	dictionaries []models.Dictionary
	books        []models.Book
	chapters     []models.Chapter
	authors      []models.Author
	phrases      []models.Phrase
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

func (r *Resolver) Role() g.RoleResolver {
	return &roleResolver{r}
}

func (r *Resolver) Book() g.BookResolver {
	return &bookResolver{r}
}

func (r *Resolver) Author() g.AuthorResolver {
	return &authorResolver{r}
}

func (r *Resolver) Chapter() g.ChapterResolver {
	return &chapterResolver{r}
}
func (r *Resolver) Phrase() g.PhraseResolver {
	return &phraseResolver{r}
}
func (r *Resolver) Dictionary() g.DictionaryResolver {
	return &dictionaryResolver{r}
}

// Subscription returns a subscription resolver
func (rr *Resolver) Subscription() g.SubscriptionResolver {
	return &subscriptionResolver{}
}


type mutationResolver struct{ *Resolver }

func (r *mutationResolver) Login(ctx context.Context, input booking.LoginInput) (string, error) {
	// Get the user information by the login username.
	d, err := models.GetUserByName(input.Username)
	if err != nil {
		return "", err
	}

	// Compare the login password with the user password.
	if err := auth.Compare(d.Password, input.Password); err != nil {
		return "", err
	}

	role := ""
	if len(d.Roles) > 0 {
		role = d.Roles[0].Name
	}
	fmt.Println(role)
	// Sign the json web token.
	t, err := token.Sign(token.Context{ID: d.ID, Username: d.Username, Role: role}, "")
	if err != nil {
		return "", err
	}
	return t, nil
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
		if filter.Name != nil {
			where = "name"
			whereValue = *filter.Name
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

func (r *queryResolver) Dictionaries(ctx context.Context, filter *booking.DictionaryFilterInput, pagination *booking.Pagination,) (booking.QueryDictionaryResponse, error) {
	if pagination == nil {
		pagination = &booking.Pagination{
			Skip: 0,
			Take: constvar.DefaultLimit,
		}
	}

	where := ""
	whereValue := ""
	if filter != nil {
		if filter.Word != nil {
			where = "word"
			whereValue = *filter.Word
		}
		if filter.Translation != nil {
			where = "translation"
			whereValue = *filter.Translation
		}
		if filter.Tag != nil {
			where = "tag"
			whereValue = *filter.Tag
		}
	}

	dictionaries, total, err := models.GetDictionaries(where, whereValue, pagination.Skip, pagination.Take)

	resp := booking.QueryDictionaryResponse{
		TotalCount: &total,
		Skip:       &pagination.Skip,
		Take:       &pagination.Take,
		Rows:       dictionaries,
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

func (r *queryResolver) Permissions(ctx context.Context, filter booking.RoleFilterInput) (resp booking.QueryPermissionResponse, err error) {
	runtimeViper := viper.New()
	runtimeViper.AddConfigPath("conf/permissions") // 如果没有指定配置文件，则解析默认的配置文件
	runtimeViper.SetConfigName("permission")

	runtimeViper.SetConfigType("yaml")                  // 设置配置文件格式为YAML
	if err := runtimeViper.ReadInConfig(); err != nil { // viper解析配置文件
		fmt.Println(err)
		return resp, err
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

func (r *queryResolver) Authors(ctx context.Context, filter *booking.AuthorFilterInput, pagination *booking.Pagination) (booking.QueryAuthorResponse, error) {
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

	authors, total, err := models.GetAuthors(where, whereValue, pagination.Skip, pagination.Take)
	resp := booking.QueryAuthorResponse{
		Rows:       authors,
		Skip:       &pagination.Skip,
		Take:       &pagination.Take,
		TotalCount: &total,
	}

	return resp, err
}

func (r *queryResolver) Books(ctx context.Context, filter *booking.BookFilterInput, pagination *booking.Pagination) (booking.QueryBookResponse, error) {
	if pagination == nil {
		pagination = &booking.Pagination{
			Skip: 0,
			Take: constvar.DefaultLimit,
		}
	}
	resp := booking.QueryBookResponse{}
	where := ""
	whereValue := ""
	if filter != nil {
		if filter.ID != nil {
			where = "id"
			whereValue = string(*filter.ID)
		}

		if filter.Name != nil {
			where = "name"
			whereValue = *filter.Name
		}

		//如果是按照作者来查询的话，需要先根据作者的名字来查出相关ID
		if filter.Author != nil {
			author,err := models.GetAuthor("name", *filter.Author)
			if err != nil {
				return resp, err
			}
			where = "author_id"
			whereValue = string(author.ID)
		}
	}

	books, total, err := models.GetBooks(where, whereValue, pagination.Skip, pagination.Take)
	resp = booking.QueryBookResponse{
		Rows:       books,
		Skip:       &pagination.Skip,
		Take:       &pagination.Take,
		TotalCount: &total,
	}

	return resp, err
}

func (r *queryResolver) Chapters(ctx context.Context, filter *booking.ChapterFilterInput) ([]models.Chapter, error) {
	ids := []uint64{}
	if filter != nil {
		for _, id := range filter.Ids{
			ids = append(ids, uint64(id))
		}
	}


	chapters, err := models.GetChapters(ids)
	return chapters, err
}

func (r *queryResolver) Phrases(ctx context.Context, filter *booking.PhraseFilterInput, pagination *booking.Pagination) (booking.QueryPhraseResponse, error) {
	if pagination == nil {
		pagination = &booking.Pagination{
			Skip: 0,
			Take: constvar.DefaultLimit,
		}
	}

	where := ""
	whereValue := ""

	if filter != nil {
		if filter.Content != nil {
			where = "name"
			whereValue = *filter.Content
		}
	}

	phrases, total, err := models.GetPhrases(where, whereValue, pagination.Skip, pagination.Take)
	resp := booking.QueryPhraseResponse{
		Rows:phrases,
		Skip:       &pagination.Skip,
		Take:       &pagination.Take,
		TotalCount: &total,
	}
	return resp, err
}

func (r *queryResolver) Analysis(ctx context.Context, input booking.ChapterAnalysisInput) (booking.QueryDictionaryResponse, error){
	resp := booking.QueryDictionaryResponse{}
	chapters,err  := models.GetChapters([]uint64{uint64(input.ChapterID)})
	if err != nil {
		return resp ,err
	}
	for _, c := range chapters {
		dictionaries , err := service.Ngrams(c.Content)
		if err != nil {
			return resp, err
		}
		total := len(dictionaries)
		resp.TotalCount = &total
		resp.Rows = dictionaries
	}

	return resp,nil
}

func (r *queryResolver) Messages(ctx context.Context) (string, error){
	return "hello world",nil
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
