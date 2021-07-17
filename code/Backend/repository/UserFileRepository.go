package repository

func PairUidAndFid(uid uint, fid uint) {
	db.Table("users_sheets").Create(&map[string]interface{}{
		"user_uid": uid,
		"sheet_fid": fid,
	})
	return
}

func GetFidsByUid(uid uint) (ret []uint) {
	type queryType struct {
		user_uid uint
		sheet_fid uint
	}

	rows, _ := db.Table("users_sheets").Select("user_uid", "sheet_fid").Where("user_uid = ?", uid).Rows()

	for rows.Next() {
		query := queryType{}
		_ = rows.Scan(&query.user_uid, &query.sheet_fid)
		ret = append(ret, query.sheet_fid)
	}
	return
}