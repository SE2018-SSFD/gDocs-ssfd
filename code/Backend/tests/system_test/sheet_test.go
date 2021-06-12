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

	post(t, "/modifysheet", utils.ModifySheetParams{
		Token: testToken1,
		Fid: fid,
		Row: 0,
		Col: 0,
		Content: "test0,0",
	}, iris.StatusOK, true, utils.SheetModifySuccess, nil)

	res0 := post(t, "/getsheet", utils.GetSheetParams{
		Token: testToken1,
		Fid: fid,
	}, iris.StatusOK, true, utils.SheetGetSuccess, nil).
		Value("data").Object().Value("content").Array()
	res0.Length().Equal(15)
	res0.Iter()[0].Equal("test0,0")

	post(t, "/modifysheet", utils.ModifySheetParams{
		Token: testToken1,
		Fid: fid,
		Row: 5,
		Col: 3,
		Content: "test5,3",
	}, iris.StatusOK, true, utils.SheetModifySuccess, nil)

	res1 := post(t, "/getsheet", utils.GetSheetParams{
		Token: testToken1,
		Fid: fid,
	}, iris.StatusOK, true, utils.SheetGetSuccess, nil).
		Value("data").Object().Value("content").Array()

	res1.Length().Equal(6*4)
	res1.Iter()[4*5+3].Equal("test5,3")
}
