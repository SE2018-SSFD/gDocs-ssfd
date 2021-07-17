package repository

import (
	"backend/model"
	"backend/utils"
	"backend/utils/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strconv"
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
	db.Create(&model.UserAuth{Uid: 1, Username: "test", Password: "test"})
	for i := 1; i < 10; i += 1 {
		username := "test"+strconv.Itoa(i)
		db.Create(&model.User{Uid: uint(i+1), Username: username})
		db.Create(&model.UserAuth{Uid: uint(i+1), Username: username, Password: username})
	}
}