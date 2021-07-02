package repository

import (
	"backend/model"
	"backend/utils"
	"backend/utils/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB
var err error

func InitDBConn() {
	uri := config.Get().MySqlAddr
	db, err = gorm.Open(mysql.Open(uri), &gorm.Config{})
	if err != nil {
		panic("Failed when connecting to DB " + uri)
	}

	utf8DB := db.Set("gorm:table_options","DEFAULT CHARSET=utf8mb4")
	_ = utf8DB.AutoMigrate(&model.UserAuth{})
	_ = utf8DB.AutoMigrate(&model.User{})
	_ = utf8DB.AutoMigrate(&model.Sheet{})

	if utils.IsTest {
		testPrepare()
	}
}

func testPrepare() {
	db.Create(&model.User{Uid: 1, Username: "test"})
	db.Create(&model.User{Uid: 2, Username: "test1"})
	db.Create(&model.User{Uid: 3, Username: "test2"})

	db.Create(&model.UserAuth{Uid: 1, Username: "test", Password: "test"})
	db.Create(&model.UserAuth{Uid: 2, Username: "test1", Password: "test1"})
	db.Create(&model.UserAuth{Uid: 3, Username: "test2", Password: "test2"})
}