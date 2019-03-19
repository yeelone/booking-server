package resolvers

import (
	"booking"
	"booking/models"
	"context"
	"fmt"
)

type chapterResolver struct{ *Resolver }

func (r *chapterResolver) ID(ctx context.Context, obj *models.Chapter) (int, error) {
	return int(obj.ID), nil
}
func (r *chapterResolver) CreatedAt(ctx context.Context, obj *models.Chapter) (string, error) {
	return fmt.Sprintf("yyyy-mm-dd HH:mm:ss : ", obj.CreatedAt.Format("2006-01-02 15:04:05")), nil
}

func (r *chapterResolver) UpdatedAt(ctx context.Context, obj *models.Chapter) (string, error) {
	return fmt.Sprintf(obj.UpdatedAt.Format("2006-01-02 15:04:05")), nil
}

func (r *chapterResolver) DeletedAt(ctx context.Context, obj *models.Chapter) (string, error) {
	if obj.DeletedAt != nil {
		return fmt.Sprintf(obj.DeletedAt.Format("2006-01-02 15:04:05")), nil
	}
	return "", nil
}

func (r *chapterResolver) Phrases(ctx context.Context, obj *models.Chapter) (booking.QueryDictionaryResponse, error){
	resp := booking.QueryDictionaryResponse{}

	return resp, nil

}
func (r *chapterResolver) Cet4Words(ctx context.Context, obj *models.Chapter) (booking.QueryDictionaryResponse, error){

	dicts , err := models.GetChapterWord(obj.ID, "cet4")
	total := len(dicts)
	resp := booking.QueryDictionaryResponse{
		TotalCount:&total,
		Rows:dicts,
	}
	return resp, err

}
func (r *chapterResolver) Cet6Words(ctx context.Context, obj *models.Chapter) (booking.QueryDictionaryResponse, error){

	dicts , err := models.GetChapterWord(obj.ID, "cet6")
	total := len(dicts)
	resp := booking.QueryDictionaryResponse{
		TotalCount:&total,
		Rows:dicts,
	}
	return resp, err
}
func (r *chapterResolver) KyWords(ctx context.Context, obj *models.Chapter) (booking.QueryDictionaryResponse, error){

	dicts , err := models.GetChapterWord(obj.ID, "ky")
	total := len(dicts)
	resp := booking.QueryDictionaryResponse{
		TotalCount:&total,
		Rows:dicts,
	}
	return resp, err
}
func (r *chapterResolver) ToefiWords(ctx context.Context, obj *models.Chapter) (booking.QueryDictionaryResponse, error){

	dicts , err := models.GetChapterWord(obj.ID, "toefi")
	total := len(dicts)
	resp := booking.QueryDictionaryResponse{
		TotalCount:&total,
		Rows:dicts,
	}
	return resp, err
}
func (r *chapterResolver) IeltsWords(ctx context.Context, obj *models.Chapter) (booking.QueryDictionaryResponse, error){

	dicts , err := models.GetChapterWord(obj.ID, "ielts")
	total := len(dicts)
	resp := booking.QueryDictionaryResponse{
		TotalCount:&total,
		Rows:dicts,
	}
	return resp, err
}
func (r *chapterResolver) GreWords(ctx context.Context, obj *models.Chapter) (booking.QueryDictionaryResponse, error){

	dicts , err := models.GetChapterWord(obj.ID, "gre")
	total := len(dicts)
	resp := booking.QueryDictionaryResponse{
		TotalCount:&total,
		Rows:dicts,
	}
	return resp, err
}

func (r *mutationResolver) CreateChapter(ctx context.Context, input booking.NewChapter) (models.Chapter, error) {
	m := models.Chapter{
		Name:     input.Name,
		Content:  input.Content,
	}

	if input.Index != nil {
		m.Index = *input.Index
	}

	if input.BookID != 0 {
		m.BookID = uint64(input.BookID)
	}

	chapter, err := m.Create()
	if err != nil {
		return chapter, err
	}

	return chapter, nil
}

func (r *mutationResolver) UpdateChapter(ctx context.Context, input booking.UpdateChapterInput) (models.Chapter, error) {
	m := models.Chapter{}
	data := make(map[string]interface{})
	if input.Name != nil {
		m.Name = *input.Name
		data["name"] = m.Name
	}
	if input.Content != nil {
		m.Content = *input.Content
		data["content"] = m.Content
	}
	m.ID = uint64(input.ID)

	return m,m.Update(data)
}

func (r *mutationResolver) DeleteChapter(ctx context.Context, input booking.DeleteIDInput) (bool, error) {
	if len(input.Ids) > 0 {
		id := uint64(input.Ids[0])
		if err := models.DeleteChapter(id); err != nil {
			return false, err
		}
	}
	return true, nil
}

