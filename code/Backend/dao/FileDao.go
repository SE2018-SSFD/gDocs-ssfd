package dao

import (
	"backend/model"
	"backend/repository"
)

func CreateSheet(file model.Sheet) uint {
	return repository.CreateSheet(file)
}

func GetSheetByFid(fid uint) model.Sheet {
	return repository.GetSheetByFid(fid)
}

func GetSheetWithCheckPointsByFid(fid uint) model.Sheet {
	return repository.GetSheetWithCheckPointsByFid(fid)
}

func SetSheet(file model.Sheet) {
	repository.SaveSheet(file)
	return
}

func DeleteSheet(fid uint) {
	repository.DeleteSheet(model.Sheet{
		Fid: fid,
	})
	return
}

func AddSheetToUser(uid uint, fid uint) {
	repository.PairUidAndFid(uid, fid)
}

func GetSheetFidsByUid(uid uint) []uint {
	return repository.GetFidsByUid(uid)
}

func SaveCheckPoint(checkpoint model.CheckPoint) {

}