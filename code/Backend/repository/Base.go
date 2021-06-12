package repository

import (
	"backend/model"
	"backend/utils"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB
var err error
var uri = "test:test@tcp(127.0.0.1:31087)/test?charset=utf8mb4&parseTime=True&loc=Local"


func InitDBConn() {
	db, err = gorm.Open(mysql.Open(uri), &gorm.Config{})
	if err != nil {
		panic("Failed when connecting to DB " + uri)
	}

	_ = db.AutoMigrate(&model.UserAuth{})
	_ = db.AutoMigrate(&model.CheckPoint{})
	_ = db.AutoMigrate(&model.User{})
	_ = db.AutoMigrate(&model.Sheet{})

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