package models

import "errors"

type Comment struct {
	BaseModel
	UserID      uint64
	Body       string
}

// TableName :
func (b *Comment) TableName() string {
	return "comments"
}

// Create :
func (b *Comment) Create() (comment Comment, err error) {
	err = DB.Self.Create(&b).Error
	return *b, err
}


// GetComments
func GetComments(where string, value string, skip, take int) (comments []Comment, total int, err error) {
	u := &Comment{}
	w := ""
	if len(where) > 0 && len(value) > 0 {

		w = where + " LIKE ?"
		v := "%" + value + "%"

		d := DB.Self.Where(w, v).Order("id").Offset(skip).Limit(take).Find(&comments)

		if err := DB.Self.Model(u).Where(w, v).Count(&total).Error; err != nil {
			return comments, 0, errors.New("cannot fetch count of the row")
		}
		return comments, total, d.Error
	}

	d := DB.Self.Debug().Order("id").Offset(skip).Limit(take).Find(&comments)
	if err := DB.Self.Model(u).Count(&total).Error; err != nil {
		return comments, 0, errors.New("cannot fetch count of the row")
	}
	return comments, total, d.Error

}