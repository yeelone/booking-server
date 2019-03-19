package resolvers

import (
	"booking"
	"booking/models"
	"context"
	"fmt"
)

type dictionaryResolver struct{ *Resolver }

func (r *dictionaryResolver) ID(ctx context.Context, obj *models.Dictionary) (int, error) {
	return int(obj.ID), nil
}

func (r *dictionaryResolver) CreatedAt(ctx context.Context, obj *models.Dictionary) (string, error) {
	return fmt.Sprintf("yyyy-mm-dd HH:mm:ss : ", obj.CreatedAt.Format("2006-01-02 15:04:05")), nil
}

func (r *dictionaryResolver) UpdatedAt(ctx context.Context, obj *models.Dictionary) (string, error) {
	return fmt.Sprintf(obj.UpdatedAt.Format("2006-01-02 15:04:05")), nil
}

func (r *dictionaryResolver) DeletedAt(ctx context.Context, obj *models.Dictionary) (string, error) {
	if obj.DeletedAt != nil {
		return fmt.Sprintf(obj.DeletedAt.Format("2006-01-02 15:04:05")), nil
	}
	return "", nil
}


func (r *mutationResolver) CreateChapterAndDictionaryRelationship(ctx context.Context, input booking.ChapterAndDictionaryRelationshipInput) (bool, error){
	chapterId := uint64(input.ChapterID)

	if _,err := models.GetChapterByID(chapterId) ; err != nil {
		return false,err
	}

	ids := []uint64{}
	for _, id := range input.DictID{
		ids = append(ids, uint64(id))
	}

	if err := models.AddChapterDictionary(chapterId,ids,input.Level.String()); err != nil {
		return false, err
	}
	return true,nil
}

func (r *mutationResolver) RemoveChapterAndDictionaryRelationship(ctx context.Context, input booking.ChapterAndDictionaryRelationshipInput) (bool, error){
	chapterId := uint64(input.ChapterID)

	if _,err := models.GetChapterByID(chapterId) ; err != nil {
		return false,err
	}

	ids := []uint64{}
	for _, id := range input.DictID{
		ids = append(ids, uint64(id))
	}

	if err := models.RemoveChapterDictionary(chapterId,ids,input.Level.String()); err != nil {
		return false, err
	}
	return true,nil
}


func (r *mutationResolver) CreateDictionary(ctx context.Context, input booking.NewDictionary) (models.Dictionary, error){
	m := models.Dictionary{
		Word :input.Word,
		Phonetic:*input.Phonetic,
		Definition:*input.Definition,
		Translation:input.Translation,
		Pos :*input.Pos,
		Collins :*input.Collins,
		Oxford   :*input.Oxford,
		Tag   :*input.Tag,
		Bnc    :*input.Bnc,
		Frq        :*input.Frq,
		Exchange   :*input.Exchange,
		Detail     :*input.Detail,
		Audio:*input.Audio,
	}
	return m.Create()
}
func (r *mutationResolver) UpdateDictionary(ctx context.Context, input booking.UpdateDictionaryInput) (models.Dictionary, error){
	m := models.Dictionary{}
	data := make(map[string]interface{})
	if input.Word != nil {
		m.Word = *input.Word
		data["word"] = m.Word
	}
	if input.Phonetic != nil {
		m.Phonetic = *input.Phonetic
		data["phonetic"] = *input.Phonetic
	}
	if input.Definition != nil {
		m.Definition = *input.Definition
		data["definition"] = *input.Definition
	}
	if input.Translation != nil {
		m.Translation = *input.Translation
		data["translation"] = *input.Translation
	}
	if input.Pos != nil {
		data["pos"] = *input.Pos
	}
	if input.Collins != nil {
		data["collins"] = *input.Collins
	}
	if input.Oxford != nil {
		data["oxford"] = *input.Oxford
	}
	if input.Tag != nil {
		data["tag"] = *input.Tag
	}
	if input.Bnc != nil {
		data["bnc"] = *input.Bnc
	}
	if input.Frq != nil {
		data["frq"] = *input.Frq
	}
	if input.Exchange != nil {
		data["exchange"] = *input.Exchange
	}
	if input.Detail != nil {
		data["detail"] = *input.Detail
	}
	if input.Audio != nil {
		data["audio"] = *input.Audio
	}
	m.ID = uint64(input.ID)

	return m,m.Update(data)
}
func (r *mutationResolver) DeleteDictionary(ctx context.Context, input booking.DeleteIDInput) (bool, error){
	if len(input.Ids) > 0 {
		id := uint64(input.Ids[0])
		if err := models.DeleteDictionary(id); err != nil {
			fmt.Println("err", err)
			return false, err
		}
	}
	return true, nil
}