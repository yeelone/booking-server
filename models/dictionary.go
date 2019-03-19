package models

import (
	"errors"
	"strings"
)


var WordTag = map[string]string{
	"zk":"中考",
	"gk":"高考",
	"cet4":"四级",
	"cet6":"六级",
	"ky":"考研",
	"toefl":"托福",
	"ielts": "雅思",
	"gre":"GRE",
}

// Dictionary 字典
type Dictionary struct {
	BaseModel
	Word        string `json:"word"`
	Phonetic    string `json:"phonetic"`
	Definition  string `json:"definition"`
	Translation string `json:"translation"`
	Pos         string `json:"pos"`
	Collins     string `json:"collins"`
	Oxford      string `json:"oxford"`
	Tag         string `json:"tag"`
	Bnc         string `json:"bnc"`
	Frq         string `json:"frq"`
	Exchange    string `json:"exchange"`
	Detail      string `json:"detail"`
	Audio       string `json:"audio"`
}

//DictionaryTableName 词典表名
const DictionaryTableName = "dictionary"

// TableName :
func (d *Dictionary) TableName() string {
	return DictionaryTableName
}

// Create : Create a new Book
func (d Dictionary) Create() (dict Dictionary, err error) {
	err = DB.Self.Create(&d).Error
	return d, err
}

func (d *Dictionary) Update(data map[string]interface{}) error {
	tx := DB.Self.Begin()
	if err := tx.Model(&d).Update(data).Error; err != nil {
		tx.Rollback()
		return errors.New("无法更新")
	}
	tx.Commit()
	return nil
}

// DeleteBook deletes the role by the user identifier.
func DeleteDictionary(id uint64) error {
	dict := Dictionary{}
	dict.ID = id
	return DB.Self.Delete(&dict).Error
}

// QueryPhrase  查询短语
// @params phrase
func QueryPhrase(phrase string) (result []Dictionary, err error) {
	// err = DB.Self.Debug().Where("word=?", phrase).Find(&result).Error
	return result, err
}

// QueryPhrases  查询短语
// @params phrases 支持多个短语
// 查询数据太多，必须给数据库加索引
func QueryPhrases(phrases []string) (result []Dictionary, err error) {
	err = DB.Self.Where("word IN (?)", phrases).Find(&result).Error
	return result, err
}

//QueryPhrasesV2 查询短语
func QueryPhrasesV2(phrases []string) (result []Dictionary, err error) {

	sql := `SELECT * FROM ` + DictionaryTableName + ` WHERE `
	where := []string{}

	for _, p := range phrases {
		where = append(where, " word='"+p+"'")
	}

	w := strings.Join(where, " OR ")

	sql += w

	// err = DB.Self.Raw(sql).Scan(&result).Error

	// err = DB.Self.Debug().Where("word IN (?)", phrases).Find(&result).Error
	return result, err
}


// GetDictionaries
func GetDictionaries(where string,value string, skip, take int) (dictionaries []Dictionary,total int ,err error) {
	a := &Dictionary{}
	w := ""
	if len(where) > 0 {

		w = where + " LIKE ?"
		v := "%" +  value + "%"

		d := DB.Self.Debug().Where(w,v).Order("id").Offset(skip).Limit(take).Find(&dictionaries)

		if err := DB.Self.Debug().Model(a).Where(w, v).Count(&total).Error; err != nil {
			return dictionaries, 0, errors.New("cannot fetch count of the row")
		}
		return dictionaries, total, d.Error
	}

	d := DB.Self.Debug().Order("id").Offset(skip).Limit(take).Find(&dictionaries)
	if err := DB.Self.Debug().Model(a).Count(&total).Error; err != nil {
		return dictionaries, 0, errors.New("cannot fetch count of the row")
	}
	return dictionaries,total, d.Error

}
