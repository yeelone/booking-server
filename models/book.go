package models

import (
	"booking/util"
	"errors"
)

type Book struct {
	BaseModel
	Name     string   `gorm:"column:name;not null"`
	Authors  []Author `gorm:"many2many:book_authors;"`
	Picture  string
	Alias    string
	User     User
	UserID   uint64
	Chapters []Chapter
}

// TableName :
func (b *Book) TableName() string {
	return "books"
}

// Create : Create a new Book
func (b Book) Create() (book Book, err error) {
	if b.UserID != 0 {
		_, err := GetUserByID(b.UserID)
		if err != nil {
			return b, err
		}
	}
	err = DB.Self.Create(&b).Error
	return b, err
}

// GetBook :
func GetBookByID(id uint64) (book Book, err error) {
	if id == 0 {
		return book, errors.New("cannot find Book by id " + util.Uint2Str(id))
	}
	err = DB.Self.Select("id,name").First(&book, id).Error
	return book, err
}

// GetBooks
func GetBooks(where string, value string, skip, take int) (books []Book, total int, err error) {
	a := &Book{}
	w := ""
	if len(where) > 0 {
		w = where + " = ?"
		d := DB.Self.Debug().Where(w, value).Order("id").Offset(skip).Limit(take).Find(&books)

		if err := DB.Self.Debug().Model(a).Where(w, value).Count(&total).Error; err != nil {
			return books, 0, errors.New("cannot fetch count of the row")
		}
		return books, total, d.Error
	}

	d := DB.Self.Debug().Order("id").Offset(skip).Limit(take).Find(&books)
	if err := DB.Self.Debug().Model(a).Count(&total).Error; err != nil {
		return books, 0, errors.New("cannot fetch count of the row")
	}
	return books, total, d.Error

}

// GetChaptersOfBook :
func GetChaptersOfBook(bookId uint64, where string, value string) (chapters []Chapter, err error) {
	if bookId == 0 {
		return chapters, errors.New("cannot find chapters by book id " + util.Uint2Str(bookId))
	}

	book := Book{}
	book.ID = bookId

	w := ""
	if len(where) > 0 {
		w = where + " = ?"
		err = DB.Self.Debug().Model(&book).Where(w, value).Related(&chapters).Error
	} else {
		err = DB.Self.Debug().Model(&book).Related(&chapters).Error
	}

	return chapters, err
}

// GetAuthorsOfBook :
func GetAuthorsOfBook(bookId uint64) (authors []Author, err error) {
	if bookId == 0 {
		return authors, errors.New("cannot find authors by book id " + util.Uint2Str(bookId))
	}

	book := Book{}
	book.ID = bookId

	err = DB.Self.Debug().Model(&book).Association("Authors").Find(&authors).Error

	return authors, err
}

func (b *Book) Update(data map[string]interface{}) error {
	tx := DB.Self.Begin()
	if err := tx.Model(&b).Update(data).Error; err != nil {
		tx.Rollback()
		return errors.New("无法更新")
	}
	tx.Commit()
	return nil
}

// DeleteBook deletes the role by the user identifier.
func DeleteBook(id uint64) error {
	book := Book{}
	book.ID = id
	return DB.Self.Delete(&book).Error
}

//AddBookChapters 为书本添加章节
func AddBookChapters(bookId uint64, chapterIds []uint64) error {
	book := Book{}
	book.ID = bookId
	chapters := []Chapter{}
	for _, id := range chapterIds {
		c := Chapter{}
		c.ID = id
		chapters = append(chapters, c)
	}

	return DB.Self.Debug().Model(&book).Association("Chapters").Append(&chapters).Error
}

//removeBookChapter 移除章节
func RemoveBookChapters(bookId uint64, chapterIds []uint64) error {
	book := Book{}
	book.ID = bookId
	chapters := []Chapter{}
	for _, id := range chapterIds {
		c := Chapter{}
		c.ID = id
		chapters = append(chapters, c)
	}

	return DB.Self.Model(&book).Association("Chapters").Delete(chapters).Error
}
