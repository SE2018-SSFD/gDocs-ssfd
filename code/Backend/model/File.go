package model

import "time"

type Sheet struct {
	Fid			uint			`gorm:"primaryKey;AUTOINCREMENT=1" json:"fid"`
	IsDeleted	bool			`gorm:"type:TINYINT;default:0" json:"isDeleted"`
	Name		string			`gorm:"type:VARCHAR(100)" json:"name"`
	CheckPoints	[]CheckPoint	`json:"checkpoints"`
	Path		string			`gorm:"type:VARCHAR(200)" json:"-"`
	Users		[]User			`gorm:"many2many:users_sheets" json:"-"`
	Columns		int				`json:"columns"`

	Content		[]string		`gorm:"-" json:"content"`
}

type CheckPoint struct {
	Cid			uint			`gorm:"primaryKey;AUTOINCREMENT=1" json:"cid"`
	SheetID		uint			`json:"-"`
	CreatedAt	time.Time
	Columns		uint			`json:"columns"`
	Path		string			`gorm:"type:VARCHAR(200)" json:"-"`

	Content		[]string		`gorm:"-" json:"content"`
}