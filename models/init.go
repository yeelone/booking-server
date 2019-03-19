package models
import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/spf13/viper"
)

type Database struct {
	Self   *gorm.DB
}

var DB *Database

var TableNames = map[string]string{"Profile": "tb_profile", "Template": "tb_template"}

func openDB(username, password, addr, name string) *gorm.DB {

	config := fmt.Sprintf("host=%s dbname=%s user=%s  password=%s sslmode=disable",
		addr,
		name,
		username,
		password,
	)

	db, err := gorm.Open("postgres", config)
	if err != nil {
		//log.Errorf(err, "Database connection failed. Database name: %s", name)
	}

	// set for db connection
	setupDB(db)

	return db
}

func setupDB(db *gorm.DB) {
	db.LogMode(viper.GetBool("gormlog"))
	//db.DB().SetMaxOpenConns(20000) // 用于设置最大打开的连接数，默认值为0表示不限制.设置最大的连接数，可以避免并发太高导致连接mysql出现too many connections的错误。
	db.DB().SetMaxIdleConns(0) // 用于设置闲置的连接数.设置闲置的连接数则当开启的一个连接使用完成后可以放在池里等候下一次使用。
}

// InitSelfDB ; used for cli
func InitSelfDB() *gorm.DB {
	return openDB(viper.GetString("db.username"),
		viper.GetString("db.password"),
		viper.GetString("db.addr"),
		viper.GetString("db.name"))
}

//GetSelfDB :
func GetSelfDB() *gorm.DB {
	return InitSelfDB()
}

//Init :
func (db *Database) Init() {
	DB = &Database{
		Self: GetSelfDB(),
	}

	initTable()

}

//Close :
func (db *Database) Close() {
	DB.Self.Close()
}

//InitDatabaseTable :
func initTable() {
	var chapter Chapter
	var user User
	var group Group
	var role Role
	var book Book
	var phrase Phrase
	var author Author
	var dict   Dictionary
	DB.Self.AutoMigrate(&dict,&author, &book, &chapter, &phrase, &user,&group,&role)
}
