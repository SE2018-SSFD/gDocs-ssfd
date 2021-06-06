package repository

import (
	"backend/model"
	"backend/utils"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB
var err error
var uri = "test:test@tcp(123.57.65.161:30087)/test?charset=utf8mb4&parseTime=True&loc=Local"


func InitDBConn() {
	db, err = gorm.Open(mysql.Open(uri), &gorm.Config{})
	if err != nil {
		panic("Failed when connecting to DB " + uri)
	}

	_ = db.AutoMigrate(&model.UserAuth{})
	_ = db.AutoMigrate(&model.User{})
	_ = db.AutoMigrate(&model.Sheet{})
	_ = db.AutoMigrate(&model.CheckPoint{})

	if utils.IsTest {
		testPrepare()
	}
}

func testPrepare() {
}