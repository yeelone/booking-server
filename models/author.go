package models

import (
	"booking/pkg/constvar"
	"booking/util"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Author struct {
	BaseModel
	Name  string `json:"name" gorm:"column:name;not null"`
	Books []Book `json:"books" gorm:"many2many:book_authors;"`
}

// TableName :
func (b *Author) TableName() string {
	return "authors"
}

// Create : Create a new Phrase
func (b Author) Create() (author Author, err error) {
	err = DB.Self.Create(&b).Error
	return b, err
}

// GetAuthors
func GetAuthors(where string, value string, skip, take int) (authors []Author, total int, err error) {
	a := &Author{}
	w := ""
	if len(where) > 0 {
		w = where + " LIKE ?"
		v := "%" + value + "%"
		d := DB.Self.Debug().Where(w, v).Order("id").Offset(skip).Limit(take).Find(&authors)

		if err := DB.Self.Debug().Model(a).Where(w, v).Count(&total).Error; err != nil {
			return authors, 0, errors.New("cannot fetch count of the row")
		}
		return authors, total, d.Error
	}

	d := DB.Self.Debug().Order("id").Offset(skip).Limit(take).Find(&authors)
	if err := DB.Self.Debug().Model(a).Count(&total).Error; err != nil {
		fmt.Println(err)
		return authors, 0, errors.New("cannot fetch count of the row")
	}
	return authors, total, d.Error

}

// GetAuthor : 根据名字查找作者，必须提供精确的条件
func GetAuthor(where string, value string) (author Author, err error) {
	w := ""
	if len(where) > 0 {
		w = where + " = ?"
		d := DB.Self.Debug().Where(w, value).First(&author)

		return author, d.Error
	}

	return author, errors.New("condition can not be empty")
}

func GetAuthorByID(id uint64) (Author, error) {
	a := Author{}
	d := DB.Self.Where("id = ?", id).First(&a)

	return a, d.Error
}

func (b *Author) Update(data map[string]interface{}) error {
	tx := DB.Self.Begin()
	if err := tx.Model(&b).Update(data).Error; err != nil {
		tx.Rollback()
		return errors.New("无法更新")
	}
	tx.Commit()
	return nil
}

// DeleteBook deletes the role by the user identifier.
func DeleteAuthor(id uint64) error {
	author := Author{}
	author.ID = id
	return DB.Self.Delete(&author).Error
}

//GetAuthorRelatedBooks :
func GetAuthorRelatedBooks(aid uint64, offset, limit int) (books []Book, total int, err error) {
	if limit == 0 {
		limit = constvar.DefaultLimit
	}
	a := &Author{}
	a.ID = aid

	uids := []uint64{}

	selectSql := ""
	countSql := ""
	if aid == 0 {
		selectSql = "SELECT book_id from book_authors offset " + strconv.Itoa(offset) + " limit " + strconv.Itoa(limit)
	} else {
		selectSql = "SELECT book_id from book_authors where author_id = " + util.Uint2Str(aid) + " offset " + strconv.Itoa(offset) + " limit " + strconv.Itoa(limit)
		countSql = "SELECT  count(book_id) from book_authors where author_id = " + util.Uint2Str(aid)
	}
	rows, _ := DB.Self.Raw(selectSql).Rows() // Note: Ignoring errors for brevity

	for rows.Next() {
		var id uint64
		if err := rows.Scan(&id); err != nil {
			return nil, 0, err
		}
		uids = append(uids, id)
	}

	if err := DB.Self.Where(" id in (?)", uids).Find(&books).Error; err != nil {
		return books, 0, err
	}

	if aid == 0 {
		DB.Self.Model(Book{}).Count(&total)
	} else {
		rows, _ := DB.Self.Raw(countSql).Rows()
		for rows.Next() {
			rows.Scan(&total)
		}
	}

	return books, total, nil
}

//AddAuthorBooks :
func AddAuthorBooks(rid uint64, ids []uint64) (err error) {
	r := Author{}

	if r, err = GetAuthorByID(rid); err != nil {
		return errors.New("Author is not existed!")
	}

	tx := DB.Self.Begin()

	var books []Book
	for _, id := range ids {
		b := Book{}
		b.ID = id
		books = append(books, b)
		//tx.Model(&b).Association("Authors").Clear()
	}

	err = tx.Model(&r).Association("Books").Append(books).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

//RemoveBookFromAuthor :
func RemoveBookFromAuthor(rid uint64, idList []uint64) (err error) {
	a := Author{}
	if a, err = GetAuthorByID(rid); err != nil {
		return errors.New("author not existed!")
	}

	tx := DB.Self.Begin()

	ids := make([]string, len(idList))

	for i, id := range idList {
		ids[i] = util.Uint2Str(id)
	}
	err = tx.Model(&a).Exec(" delete from book_authors where book_id in (" + strings.Join(ids, ",") + ") and author_id = " + util.Uint2Str(rid) + " ;").Error
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}
