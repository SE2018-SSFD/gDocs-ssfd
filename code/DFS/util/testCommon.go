package util

import "testing"

const (
	CLIENTADDR = "127.0.0.1:1233"
	MASTERADDR = "127.0.0.1:1234"
	PARALLELSIZE = 3
)
func AssertEqual(t *testing.T, a, b interface{}) {
	t.Helper()
	if a != b {
		t.Fail()
		t.Error("AssertEqual Failed :",a, b)
	}
}
func AssertNotEqual(t *testing.T, a, b interface{}) {
	t.Helper()
	if a == b {
		t.Fail()
		t.Error("AssertNotEqual Failed :",a, b)
	}
}
func AssertNil(t *testing.T, a interface{}) {
	t.Helper()
	if a != nil {
		t.Fail()
		t.Error("AssertNil Failed :",a)
	}
}
func AssertNotNil(t *testing.T, a interface{}) {
	t.Helper()
	if a == nil {
		t.Fail()
		t.Error("AssertNotNil Failed :",a)
	}
}