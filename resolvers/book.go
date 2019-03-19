package resolvers

import (
	"booking"
	"booking/models"
	"context"
	"fmt"
	"strconv"
)

type bookResolver struct{ *Resolver }

func (r *bookResolver) ID(ctx context.Context, obj *models.Book) (int, error) {
	return int(obj.ID), nil
}

func (r *bookResolver) Name(ctx context.Context, obj *models.Book) (string, error) {
	return "haha" + obj.Name, nil
}
func (r *bookResolver) CreatedAt(ctx context.Context, obj *models.Book) (string, error) {
	return fmt.Sprintf("yyyy-mm-dd HH:mm:ss : ", obj.CreatedAt.Format("2006-01-02 15:04:05")), nil
}

func (r *bookResolver) UpdatedAt(ctx context.Context, obj *models.Book) (string, error) {
	return fmt.Sprintf(obj.UpdatedAt.Format("2006-01-02 15:04:05")), nil
}

func (r *bookResolver) DeletedAt(ctx context.Context, obj *models.Book) (string, error) {
	if obj.DeletedAt != nil {
		return fmt.Sprintf(obj.DeletedAt.Format("2006-01-02 15:04:05")), nil
	}
	return "", nil
}
func (r *bookResolver) Users(ctx context.Context, obj *models.Book) (booking.QueryUserResponse, error) {
	resp := booking.QueryUserResponse{
	}

	user, err := models.GetUserByID(obj.UserID)
	if err != nil {
		return resp ,err
	}
	users := []models.User{user}

	total := 1
	resp = booking.QueryUserResponse{
		TotalCount: &total,
		Rows:users,
	}
	return resp, nil
}

func (r *bookResolver) Authors(ctx context.Context, obj *models.Book) (booking.QueryAuthorResponse, error) {
	resp := booking.QueryAuthorResponse{
	}

	authors, err := models.GetAuthorsOfBook(obj.ID)
	if err != nil {
		return resp ,err
	}

	total := len(authors)
	resp = booking.QueryAuthorResponse{
		TotalCount: &total,
		Rows:authors,
	}
	return resp, nil
}


func (r *bookResolver) Chapters(ctx context.Context, obj *models.Book, filter *booking.ChapterFilterInput, pagination *booking.Pagination) (resp booking.QueryChapterResponse, err error) {
	where := ""
	whereValue := ""
	if filter != nil {
		if filter.Index != nil {
			where = "index"
			whereValue = strconv.Itoa(*filter.Index)
		}
	}

	chapters,err := models.GetChaptersOfBook(obj.ID,where, whereValue)
	total := len(chapters)
	resp = booking.QueryChapterResponse{
		TotalCount:&total,
		Rows:chapters,
	}
	return resp, err
}

func (r *mutationResolver) CreateBook(ctx context.Context, input booking.NewBook) (models.Book, error) {

	m := models.Book{
		Name:     input.Name,
		Alias:    input.Alias,
		UserID: uint64(input.UserID),
		Picture:  input.Picture,
	}
	return m.Create()
}

func (r *mutationResolver) UpdateBook(ctx context.Context, input booking.UpdateBookInput) (models.Book, error) {
	m := models.Book{}
	data := make(map[string]interface{})
	if input.Name != nil {
		m.Name = *input.Name
		data["name"] = m.Name
	}
	if input.Alias != nil {
		m.Alias = *input.Alias
		data["alias"] = *input.Alias
	}
	if input.Picture != nil {
		m.Picture = *input.Picture
		data["picture"] = *input.Picture
	}
	if input.UserID != nil {
		m.UserID = uint64(*input.UserID)
		data["user_id"] = uint64(*input.UserID)
	}
	if input.AuthorID != nil {
		data["author_id"] = uint64(*input.AuthorID)
	}
	m.ID = uint64(input.ID)

	return m,m.Update(data)
}

func (r *mutationResolver) DeleteBook(ctx context.Context, input booking.DeleteIDInput) (bool, error) {
	if len(input.Ids) > 0 {
		id := uint64(input.Ids[0])
		if err := models.DeleteBook(id); err != nil {
			fmt.Println("err", err)
			return false, err
		}
	}
	return true, nil
}

func (r *mutationResolver) AddBookChapters(ctx context.Context, input booking.BookAndChapterRelationshipInput) (bool, error){
	bookId := uint64(input.BookID)

	if _,err := models.GetBookByID(bookId) ; err != nil {
		return false,err
	}

	cids := []uint64{}
	for _, id := range input.ChapterIds{
		cids = append(cids, uint64(id))
	}

	if err := models.AddBookChapters(bookId,cids); err != nil {
		return false, err
	}
	return true,nil

}

func (r *mutationResolver) RemoveBookChapters(ctx context.Context, input booking.BookAndChapterRelationshipInput) (bool, error){
	bookId := uint64(input.BookID)

	if _,err := models.GetBookByID(bookId) ; err != nil {
		return false,err
	}

	cids := []uint64{}
	for _, id := range input.ChapterIds{
		cids = append(cids, uint64(id))
	}

	if err := models.RemoveBookChapters(bookId,cids); err != nil {
		return false, err
	}
	return true,nil
}
