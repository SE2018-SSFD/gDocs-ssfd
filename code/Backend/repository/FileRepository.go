package repository

import (
	"backend/model"
)

func GetSheetByFid(fid uint) (ret model.Sheet) {
	db.First(&ret, "fid = ?", fid)
	return
}

func GetSheetWithCheckPointsByFid(fid uint) (ret model.Sheet) {
	db.Preload("CheckPoints").First(&ret, "fid = ?", fid)
	return
}

func CreateSheet(sheet model.Sheet) uint {
	db.Create(&sheet)
	return sheet.Fid
}

func SaveSheet(sheet model.Sheet) {
	db.Save(&sheet)
	return
}

func DeleteSheet(sheet model.Sheet) {
	db.Delete(&sheet)
	return
}