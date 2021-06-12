package system_test

import (
	"backend/dao"
	"backend/service"
	"backend/utils"
	"github.com/kataras/iris/v12"
	"testing"
)

func TestLogin(t *testing.T) {
	bodySuccess := utils.LoginParams {
		Username: "test",
		Password: "test",
	}
	bodyWrongPassword := utils.LoginParams {
		Username: "test",
		Password: "wrong_password",
	}
	bodyNoUser := utils.LoginParams {
		Username: "",
		Password: "wrong_password",
	}
	bodyWrongArgsType := map[string]int {
		"username": 123,
		"password": 123,
	}
	post(t, "/login", bodySuccess, iris.StatusOK, true, utils.LoginSuccess, nil).Value("data")
	post(t, "/login", bodyWrongPassword, iris.StatusOK, false, utils.LoginWrongPassword, nil)
	post(t, "/login", bodyNoUser, iris.StatusOK, false, utils.LoginNoSuchUser, nil)
	post(t, "/login", bodyWrongArgsType, iris.StatusBadRequest, false, utils.InvalidFormat, nil)
	if service.CheckToken(service.NewToken(1, "test")) != 1 {
		t.Fail()
	}
	oldTokenTerm := utils.TokenTerm
	utils.TokenTerm = 0
	if service.CheckToken(service.NewToken(1, "test")) != 0 {
		t.Fail()
	}
	utils.TokenTerm = oldTokenTerm
}

func TestRegister(t *testing.T) {
	bodySuccess := utils.RegisterParams {
		Username: "test_reg",
		Password: "test_reg",
		Email: "test_reg",
	}
	bodyDupUser := utils.RegisterParams {
		Username: "test_reg",
		Password: "test_reg",
	}
	bodyWrongArgsType := map[string]bool {
		"username": true,
		"password": true,
	}
	service.RemoveUserAndUserAuth("test_reg")
	post(t, "/register", bodySuccess, iris.StatusOK, true, utils.RegisterSuccess, nil)
	post(t, "/register", bodyDupUser, iris.StatusOK, false, utils.RegisterUserExists, nil)
	post(t, "/register", bodyWrongArgsType, iris.StatusBadRequest, false, utils.InvalidFormat, nil)
}

func TestGetUser(t *testing.T) {
	bodySuccess := utils.GetUserParams {
		Token: testToken1,
	}
	bodyInvalidToken := utils.GetUserParams {
		Token: "invalid!",
	}
	bodyWrongArgsType := map[string]bool {
		"token": true,
	}
	post(t, "/getuser", bodySuccess, iris.StatusOK, true, utils.UserGetSuccess, nil).
		Value("data").Object().
		ContainsKey("uid").ContainsKey("username").ContainsKey("sheets")
	post(t, "/getuser", bodyInvalidToken, iris.StatusOK, false, utils.InvalidToken, nil)
	post(t, "/getuser", bodyWrongArgsType, iris.StatusBadRequest, false, utils.InvalidFormat, nil)

	if dao.GetUserByUid(0).Uid != 0 {
		t.Fail()
	}
}

func TestModifyUser(t *testing.T) {
	bodySuccess := utils.ModifyUserParams {
		Token:    testToken2,
		Username: "test_mod",
		Email:    "test_mod",
		Turtle:   2,
		Task:     2,
	}
	bodyGetUserAfterModify := utils.GetUserParams {
		Token: testToken2,
	}
	bodyModifyDupUsername := utils.ModifyUserParams {
		Token:    testToken2,
		Username: "test",
		Email:    "test_mod",
		Turtle:   2,
		Task:     2,
	}
	bodyInvalidToken := utils.ModifyUserAuthParams {
		Token: "invalid!",
	}
	bodyWrongArgsType := map[string]bool {
		"token": true,
	}

	responseGetUser := map[string]interface{} {
		"uid": 2,
		"username": "test_mod",
	}

	post(t, "/modifyuser", bodySuccess, iris.StatusOK, true, utils.UserModifySuccess, nil)
	post(t, "/getuser", bodyGetUserAfterModify, iris.StatusOK, true, utils.UserGetSuccess, nil).
		Value("data").Object().ContainsMap(responseGetUser)
	bodySuccess.Username = "test1"
	post(t, "/modifyuser", bodySuccess, iris.StatusOK, true, utils.UserModifySuccess, nil)
	responseGetUser["username"] = "test1"
	post(t, "/getuser", bodyGetUserAfterModify, iris.StatusOK, true, utils.UserGetSuccess, nil).
		Value("data").Object().ContainsMap(responseGetUser)
	post(t, "/modifyuser", bodyModifyDupUsername, iris.StatusOK, false, utils.ModifyDupUsername, nil)
	post(t, "/modifyuser", bodyInvalidToken, iris.StatusOK, false, utils.InvalidToken, nil)
	post(t, "/modifyuser", bodyWrongArgsType, iris.StatusBadRequest, false, utils.InvalidFormat, nil)
}

func TestModifyUserAuth(t *testing.T) {
	bodySuccess := utils.ModifyUserAuthParams{
		Token:    testToken1,
		Password: "test_mod",
	}
	bodyLoginAfterModifyUserAuth := utils.LoginParams{
		Username: "test",
		Password: "test_mod",
	}
	bodyInvalidToken := utils.ModifyUserAuthParams{
		Token: "invalid!",
	}
	bodyWrongArgsType := map[string]bool{
		"token": true,
	}
	post(t, "/modifyuserauth", bodySuccess, iris.StatusOK, true, utils.UserAuthModifySuccess, nil)
	post(t, "/login", bodyLoginAfterModifyUserAuth, iris.StatusOK, true, utils.LoginSuccess, nil)
	bodySuccess.Password = "test"
	post(t, "/modifyuserauth", bodySuccess, iris.StatusOK, true, utils.UserAuthModifySuccess, nil)
	bodyLoginAfterModifyUserAuth.Password = "test"
	post(t, "/login", bodyLoginAfterModifyUserAuth, iris.StatusOK, true, utils.LoginSuccess, nil)
	post(t, "/modifyuserauth", bodyInvalidToken, iris.StatusOK, false, utils.InvalidToken, nil)
	post(t, "/modifyuserauth", bodyWrongArgsType, iris.StatusBadRequest, false, utils.InvalidFormat, nil)
}
