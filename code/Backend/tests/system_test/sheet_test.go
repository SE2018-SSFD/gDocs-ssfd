package system_test

import (
	"backend/utils"
	"github.com/kataras/iris/v12"
	"testing"
)

func TestSheet(t *testing.T) {
	fid := uint(post(t, "/newsheet", utils.NewSheetParams{
		Token: testToken1,
		Name: "test1",
		InitRows: 5,
		InitColumns: 3,
	}, iris.StatusOK, true, utils.SheetNewSuccess, nil).
		Value("data").Raw().(float64))

	post(t, "/getsheet", utils.GetSheetParams{
		Token: testToken1,
		Fid: fid,
	}, iris.StatusOK, true, utils.SheetGetSuccess, nil).
		Value("data").Object().Raw()
}
