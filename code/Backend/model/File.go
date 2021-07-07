package model

import "time"

type Sheet struct {
	Fid				uint			`gorm:"primaryKey;AUTOINCREMENT=1" json:"fid"`
	IsDeleted		bool			`gorm:"type:TINYINT;default:0" json:"isDeleted"`
	Name			string			`gorm:"type:VARCHAR(100)" json:"name"`
	CheckPointNum	int				`gorm:"-" json:"checkpoint_num"`
	Users			[]User			`gorm:"many2many:users_sheets;constraint:OnDelete:CASCADE" json:"-"`
	Columns			int				`gorm:"-" json:"columns"`
	Owner			string			`json:"owner"`

	CreatedAt		time.Time
	UpdatedAt		time.Time

	Content			[]string		`gorm:"-" json:"content"`
}