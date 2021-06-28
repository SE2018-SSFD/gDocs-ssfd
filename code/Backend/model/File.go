package model

import "time"

type Sheet struct {
	Fid				uint			`gorm:"primaryKey;AUTOINCREMENT=1" json:"fid"`
	IsDeleted		bool			`gorm:"type:TINYINT;default:0" json:"isDeleted"`
	Name			string			`gorm:"type:VARCHAR(100)" json:"name"`
	CheckPointNum	int				`json:"checkpoint_num"`
	Path			string			`gorm:"type:VARCHAR(200)" json:"-"`
	Users			[]User			`gorm:"many2many:users_sheets" json:"-"`
	Columns			int				`gorm:"-" json:"columns"`
	Owner			string			`json:"owner"`

	CreatedAt		time.Time
	UpdatedAt		time.Time

	Content			[]string		`gorm:"-" json:"content"`
}