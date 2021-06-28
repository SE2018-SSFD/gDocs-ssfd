package model

import "time"

type Sheet struct {
	Fid			uint			`gorm:"primaryKey;AUTOINCREMENT=1" json:"fid"`
	IsDeleted	bool			`gorm:"type:TINYINT;default:0" json:"isDeleted"`
	Name		string			`gorm:"type:VARCHAR(100)" json:"name"`
	CheckPoints	[]CheckPoint	`json:"checkpoints"`
	Path		string			`gorm:"type:VARCHAR(200)" json:"-"`
	Users		[]User			`gorm:"many2many:users_sheets" json:"-"`
	Columns		int				`gorm:"-" json:"columns"`

	CreatedAt	time.Time
	UpdatedAt	time.Time

	Content		[]string		`gorm:"-" json:"content"`
}

type CheckPoint struct {
	Cid			uint			`gorm:"primaryKey;AUTOINCREMENT=1" json:"cid"`
	SheetID		uint			`json:"-"`
	Columns		uint			`json:"columns"`
	Path		string			`gorm:"type:VARCHAR(200)" json:"-"`

	CreatedAt	time.Time

	Content		[]string		`gorm:"-" json:"content"`
}