package models

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/lexkong/log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/spf13/viper"
)

type Database struct {
	Self  *gorm.DB
	Cache *bolt.DB
}

var DB *Database

var TableNames = map[string]string{"Profile": "tb_profile", "Template": "tb_template"}

const Login_Record_BoltDB_Key = "login_record"
const CURRENTUSERID = "CURRENTUSERID"
const CLIENT_IP = "Client_IP"

func openDB(username, password, addr, name string) *gorm.DB {

	config := fmt.Sprintf("host=%s dbname=%s user=%s  password=%s sslmode=disable",
		addr,
		name,
		username,
		password,
	)
	fmt.Println("db config", config)
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

func GetCacheDB() *bolt.DB {
	db, _ := bolt.Open("db/cache.db", 0600, nil)
	return db

}

//Init :
func (db *Database) Init() {
	DB = &Database{
		Self:  GetSelfDB(),
		Cache: GetCacheDB(),
	}

	initTable()

}

//Close :
func (db *Database) Close() {
	DB.Self.Close()
	DB.Cache.Close()
}

//InitDatabaseTable :
func initTable() {
	var user User
	var group Group
	var role Role
	var ticket Ticket
	var dishes Dishes
	var canteen Canteen
	var booking Booking
	var record TicketRecord
	DB.Self.AutoMigrate(&user,&record, &ticket, &dishes,  &group, &role, &canteen, &booking)

	initAdmin()
}

//initAdmin: 初始化管理员账号
func initAdmin() {
	u := User{}
	//查看账号是否存在
	email := viper.GetString("admin.email")
	err := DB.Self.Where("email = ?", email).First(&u).Error

	if err != nil {
		u.Email = email
		u.Username = viper.GetString("admin.username")
		u.IDCard = "000000"
		u.Password = viper.GetString("admin.password")
		u.Save()
	}

	r := Role{}
	//查看账号是否存在
	organization := viper.GetString("role.organization")
	err = DB.Self.Where("name = ?", organization).First(&r).Error

	if err != nil {
		r.Name = organization
		r.Create()
	}

	r = Role{}
	//查看账号是否存在
	system := viper.GetString("role.system")
	err = DB.Self.Where("name = ?", system).First(&r).Error

	if err != nil {
		r.Name = system
		r.Create()
	}
	uids := make([]uint64, 1)
	uids[0] = u.ID

	if result, _ := CheckUsersNotInRole(r.ID, uids); len(result) > 0 {
		AddRoleUsers(r.ID, uids)
	}

	r = Role{}
	//查看账号是否存在
	canteen := viper.GetString("role.canteen")
	err = DB.Self.Where("name = ?", canteen).First(&r).Error

	if err != nil {
		r.Name = canteen
		r.Create()
	}

	r = Role{}
	//查看账号是否存在
	normal := viper.GetString("role.normal")
	err = DB.Self.Where("name = ?", normal).First(&r).Error

	if err != nil {
		r.Name = normal
		r.Create()
	}

}
