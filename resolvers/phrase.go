package resolvers

import (
	"booking"
	"booking/models"
	"context"
	"fmt"
)

type phraseResolver struct{ *Resolver }

func (r *phraseResolver) ID(ctx context.Context, obj *models.Phrase) (int, error) {
	return int(obj.ID), nil
}

func (r *phraseResolver) Name(ctx context.Context, obj *models.Phrase) (string, error) {
	return "haha" + obj.Content, nil
}
func (r *phraseResolver) CreatedAt(ctx context.Context, obj *models.Phrase) (string, error) {
	return fmt.Sprintf("yyyy-mm-dd HH:mm:ss : ", obj.CreatedAt.Format("2006-01-02 15:04:05")), nil
}

func (r *phraseResolver) UpdatedAt(ctx context.Context, obj *models.Phrase) (string, error) {
	return fmt.Sprintf(obj.UpdatedAt.Format("2006-01-02 15:04:05")), nil
}

func (r *phraseResolver) DeletedAt(ctx context.Context, obj *models.Phrase) (string, error) {
	if obj.DeletedAt != nil {
		return fmt.Sprintf(obj.DeletedAt.Format("2006-01-02 15:04:05")), nil
	}
	return "", nil
}

func (r *phraseResolver) Translation(ctx context.Context, obj *models.Phrase) (string, error){
	return "",nil
}




func (r *mutationResolver) CreatePhrase(ctx context.Context, input booking.NewPhrase) (models.Phrase, error) {
	m := models.Phrase{
		Content:     input.Content,
	}
	return m.Create()
}

func (r *mutationResolver) UpdatePhrase(ctx context.Context, input booking.UpdatePhraseInput) (models.Phrase, error) {
	m := models.Phrase{}
	data := make(map[string]interface{})
	if input.Content != nil {
		m.Content = *input.Content
		data["content"] = m.Content
	}
	m.ID = uint64(input.ID)

	return m,m.Update(data)
}

func (r *mutationResolver) DeletePhrase(ctx context.Context, input booking.DeleteIDInput) (bool, error) {
	if len(input.Ids) > 0 {
		id := uint64(input.Ids[0])
		if err := models.DeletePhrase(id); err != nil {
			return false, err
		}
	}
	return true, nil
}

