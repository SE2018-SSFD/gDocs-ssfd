package model

type User struct {
	Uid			uint		`gorm:"primaryKey;AUTOINCREMENT=1" json:"uid"`
	Username	string		`gorm:"uniqueIndex;type:VARCHAR(50) NOT NULL" json:"username"`
	Sheets		[]Sheet		`gorm:"many2many:users_sheets" json:"sheets"`
}

type UserAuth struct {
	Uid			uint		`gorm:"primaryKey;AUTOINCREMENT=1"`
	Username	string		`gorm:"uniqueIndex;type:VARCHAR(50) NOT NULL"`
	Password	string		`gorm:"type:VARCHAR(50) NOT NULL"`
}