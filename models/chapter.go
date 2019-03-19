package models

import (
	"booking/util"
	"errors"
)

type Chapter struct {
	BaseModel
	Index      int          `gorm:"column:index;not null"`
	Name       string       `gorm:"column:name;not null"`
	Content    string       `gorm:"column:content;not null"`
	Phrases    []Dictionary `json:"phrases" gorm:"many2many:chapter_phrase;"` //记录短语词汇
	Cet4Words  []Dictionary `json:"phrases" gorm:"many2many:chapter_cet4;;"`  //记录4级词汇
	Cet6Words  []Dictionary `json:"phrases" gorm:"many2many:chapter_cet6;"`   //记录6级词汇
	KyWords    []Dictionary `json:"phrases" gorm:"many2many:chapter_ky;"`     //记录考研词汇
	ToefiWords []Dictionary `json:"phrases" gorm:"many2many:chapter_toefi;"`  //记录托福词汇
	IeltsWords []Dictionary `json:"phrases" gorm:"many2many:chapter_ielts;"`  //记录雅思词汇
	GreWords   []Dictionary `json:"phrases" gorm:"many2many:chapter_gre;"`    //gre词汇
	BookID     uint64
}

// TableName :
func (b *Chapter) TableName() string {
	return "chapters"
}

// Create : Create a new Group
func (b *Chapter) Create() (chapter Chapter, err error) {
	err = DB.Self.Debug().Create(&b).Error
	return *b, err
}

func GetChapters(ids []uint64) (chapters []Chapter, err error) {
	err = DB.Self.Where(" id in (?) ", ids).Find(&chapters).Error
	return chapters, err
}

// GetChapterByID :
func GetChapterByID(id uint64) (chapter Chapter, err error) {
	if id == 0 {
		return chapter, errors.New("cannot find chapter by id " + util.Uint2Str(id))
	}
	err = DB.Self.First(&chapter, id).Error
	return chapter, err
}

func (b *Chapter) Update(data map[string]interface{}) error {
	tx := DB.Self.Begin()
	if err := tx.Model(&b).Update(data).Error; err != nil {
		tx.Rollback()
		return errors.New("无法更新")
	}
	tx.Commit()
	return nil
}

// DeleteBook deletes the role by the user identifier.
func DeleteChapter(id uint64) error {
	chapter := Chapter{}
	chapter.ID = id
	return DB.Self.Delete(&chapter).Error
}

//AddChapterDictionary 将选定的词汇与章节关联起来，level所指为四级、六级、考研、托福、雅思、GRE， phrase 则为短语
// level 参数为  cet4 ,cet6, ky, toefi, ielts, gre ,phrase
func AddChapterDictionary(chapterId uint64, dictIds []uint64, level string) error {
	book := Chapter{}
	book.ID = chapterId
	dicts := []Dictionary{}
	for _, id := range dictIds {
		c := Dictionary{}
		c.ID = id
		dicts = append(dicts, c)
	}

	if level == "cet4" {
		return DB.Self.Debug().Model(&book).Association("Cet4Words").Append(&dicts).Error
	}

	if level == "cet6" {
		return DB.Self.Debug().Model(&book).Association("Cet6Words").Append(&dicts).Error
	}
	if level == "ky" {
		return DB.Self.Debug().Model(&book).Association("KyWords").Append(&dicts).Error
	}
	if level == "toefi" {
		return DB.Self.Debug().Model(&book).Association("ToefiWords").Append(&dicts).Error
	}
	if level == "ielts" {
		return DB.Self.Debug().Model(&book).Association("IeltsWords").Append(&dicts).Error
	}
	if level == "gre" {
		return DB.Self.Debug().Model(&book).Association("GreWords").Append(&dicts).Error
	}
	if level == "phrase" {
		return DB.Self.Debug().Model(&book).Association("Phrases").Append(&dicts).Error
	}
	return nil
}

////RemoveChapterDictionary
func RemoveChapterDictionary(chapterId uint64, dictIds []uint64, level string) error {
	chapter := Chapter{}
	chapter.ID = chapterId
	dictionaries := []Dictionary{}
	for _, id := range dictIds {
		c := Dictionary{}
		c.ID = id
		dictionaries = append(dictionaries, c)
	}
	if level == "cet4" {
		return DB.Self.Debug().Model(&chapter).Association("Cet4Words").Append(&dictionaries).Error
	}
	if level == "cet6" {
		return DB.Self.Debug().Model(&chapter).Association("Cet6Words").Append(&dictionaries).Error
	}
	if level == "ky" {
		return DB.Self.Debug().Model(&chapter).Association("KyWords").Append(&dictionaries).Error
	}
	if level == "toefi" {
		return DB.Self.Debug().Model(&chapter).Association("ToefiWords").Append(&dictionaries).Error
	}
	if level == "ielts" {
		return DB.Self.Debug().Model(&chapter).Association("IeltsWords").Append(&dictionaries).Error
	}
	if level == "gre" {
		return DB.Self.Debug().Model(&chapter).Association("GreWords").Append(&dictionaries).Error
	}
	if level == "phrase" {
		return DB.Self.Debug().Model(&chapter).Association("Phrases").Append(&dictionaries).Error
	}
	return nil

}

func GetChapterWord(chapterId uint64, level string) (dicts []Dictionary, err error) {
	chapter := Chapter{}
	chapter.ID = chapterId

	if level == "cet4" {
		err = DB.Self.Debug().Model(&chapter).Association("Cet4Words").Find(&dicts).Error
	}

	if level == "cet6" {
		err = DB.Self.Debug().Model(&chapter).Association("Cet6Words").Find(&dicts).Error
	}
	if level == "ky" {
		err = DB.Self.Debug().Model(&chapter).Association("KyWords").Find(&dicts).Error
	}
	if level == "toefi" {
		err = DB.Self.Debug().Model(&chapter).Association("ToefiWords").Find(&dicts).Error
	}
	if level == "ielts" {
		err = DB.Self.Debug().Model(&chapter).Association("IeltsWords").Find(&dicts).Error
	}
	if level == "gre" {
		err = DB.Self.Debug().Model(&chapter).Association("GreWords").Find(&dicts).Error
	}
	if level == "phrase" {
		err = DB.Self.Debug().Model(&chapter).Association("Phrases").Find(&dicts).Error
	}
	return dicts, err
}
