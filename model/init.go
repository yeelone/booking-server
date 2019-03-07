package model

import (
	"fmt"

	"github.com/lexkong/log"
	"github.com/spf13/viper"
	// MySQL driver.
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type Database struct {
	Self   *gorm.DB
	Docker *gorm.DB
}

var DB *Database

func openDB(username, password, addr, name string) *gorm.DB {

	config := fmt.Sprintf("host=%s dbname=%s user=%s  password=%s sslmode=disable",
		addr,
		name,
		username,
		password,
	)

	db, err := gorm.Open("postgres", config)
	if err != nil {
		log.Errorf(err, "Database connection failed. Database name: %s", name)
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

// used for cli
func InitSelfDB() *gorm.DB {
	return openDB(viper.GetString("db.username"),
		viper.GetString("db.password"),
		viper.GetString("db.addr"),
		viper.GetString("db.name"))
}

func GetSelfDB() *gorm.DB {
	return InitSelfDB()
}

func (db *Database) Init() {
	DB = &Database{
		Self: GetSelfDB(),
	}

	initTable()
}

func (db *Database) Close() {
	DB.Self.Close()
}

//InitDatabaseTable :
func initTable() {
	var userModel UserModel
	DB.Self.AutoMigrate(&userModel)
	initAdmin()
}

//initAdmin: 初始化管理员账号
func initAdmin() {
	u := UserModel{}
	//查看账号是否存在
	email := viper.GetString("admin.email")
	err := DB.Self.Where("email = ?", email).First(&u).Error

	if err != nil {
		u.Email = email
		u.Username = viper.GetString("admin.username")
		u.Password = viper.GetString("admin.password")
		u.Save()
	}
}
