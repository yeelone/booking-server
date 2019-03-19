package models

import "errors"

type Phrase struct {
	BaseModel
	Content string `json:"content" gorm:"column:content;not null"`
}

// TableName :
func (b *Phrase) TableName() string {
	return "phrase"
}

// Create : Create a new Phrase
func (b Phrase) Create() (phrase Phrase,err error) {
	err = DB.Self.Create(&b).Error
	return b,err
}

// GetPhrases
func GetPhrases(where string,value string, skip, take int) (phrases []Phrase,total int ,err error) {
	a := &Author{}
	w := ""
	if len(where) > 0 {
		w = where + " LIKE %?%"
		d := DB.Self.Debug().Where(w,value).Order("id").Offset(skip).Limit(take).Find(&phrases)

		if err := DB.Self.Debug().Model(a).Where(w, value).Count(&total).Error; err != nil {
			return phrases, 0, errors.New("cannot fetch count of the row")
		}
		return phrases, total, d.Error
	}

	d := DB.Self.Debug().Order("id").Offset(skip).Limit(take).Find(&phrases)
	if err := DB.Self.Debug().Model(a).Count(&total).Error; err != nil {
		return phrases, 0, errors.New("cannot fetch count of the row")
	}
	return phrases,total, d.Error
}

func (b *Phrase) Update(data map[string]interface{}) error {
	tx := DB.Self.Begin()
	if err := tx.Model(&b).Update(data).Error; err != nil {
		tx.Rollback()
		return errors.New("无法更新")
	}
	tx.Commit()
	return nil
}

// DeleteBook deletes the role by the user identifier.
func DeletePhrase(id uint64) error {
	phrase := Phrase{}
	phrase.ID = id
	return DB.Self.Delete(&phrase).Error
}