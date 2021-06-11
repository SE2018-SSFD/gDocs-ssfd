package util

import "testing"

const (
	CLIENTADDR = "127.0.0.1:1233"
	MASTERADDR = "127.0.0.1:1234"
)
func AssertEqual(t *testing.T, a, b interface{}) {
	if a != b {
		t.Error("AssertEqual Failed :",a, b)
	}
}
func AssertNotEqual(t *testing.T, a, b interface{}) {
	if a == b {
		t.Error("AssertNotEqual Failed :",a, b)
	}
}
func AssertNil(t *testing.T, a interface{}) {
	if a != nil {
		t.Error("AssertNil Failed :",a)
	}
}
func AssertNotNil(t *testing.T, a interface{}) {
	if a == nil {
		t.Error("AssertNotNil Failed :",a)
	}
}