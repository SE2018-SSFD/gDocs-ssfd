package util

import "testing"

const (
	CLIENTADDR     = "127.0.0.1:1233"
	MASTER1ADDR     = "127.0.0.1:1234"
	MASTER2ADDR     = "127.0.0.1:1235"
	MASTER3ADDR     = "127.0.0.1:1236"

	PARALLELSIZE   = 3
	MAXWAITINGTIME = 5 // the time to wait for parallel tests finished
)

func AssertEqual(t *testing.T, a, b interface{}) {
	t.Helper()
	if a != b {
		t.Fail()
		t.Fatal("AssertEqual Failed :", a, b)
	}
}
func AssertGreater(t *testing.T, a, b interface{}) {
	t.Helper()
	if a != b {
		t.Fail()
		t.Fatal("AssertEqual Failed :", a, b)
	}
}
func AssertTrue(t *testing.T, a bool) {
	t.Helper()
	if a != true {
		t.Fail()
		t.Fatal("AssertTrue Failed :", a)
	}
}
func AssertNotTrue(t *testing.T, a bool) {
	t.Helper()
	if a != false {
		t.Fail()
		t.Fatal("AssertNotTrue Failed :", a)
	}
}
func AssertNotEqual(t *testing.T, a, b interface{}) {
	t.Helper()
	if a == b {
		t.Fail()
		t.Fatal("AssertNotEqual Failed :", a, b)
	}
}
func AssertNil(t *testing.T, a interface{}) {
	t.Helper()
	if a != nil {
		t.Fail()
		t.Fatal("AssertNil Failed :", a)
	}
}
func AssertNotNil(t *testing.T, a interface{}) {
	t.Helper()
	if a == nil {
		t.Fail()
		t.Fatal("AssertNotNil Failed :", a)
	}
}
