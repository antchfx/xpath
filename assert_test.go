package xpath

import (
	"reflect"
	"testing"
)

func assertEqual(tb testing.TB, v1, v2 interface{}) {
	if !reflect.DeepEqual(v1, v2) {
		tb.Fatalf("'%+v' and '%+v' are not equal", v1, v2)
	}
}

func assertNoErr(tb testing.TB, err error) {
	if err != nil {
		tb.Fatalf("expected no err, but got: %s", err.Error())
	}
}

func assertErr(tb testing.TB, err error) {
	if err == nil {
		tb.Fatal("expected err, but got nil")
	}
}

func assertTrue(tb testing.TB, v bool) {
	if !v {
		tb.Fatal("expected true, but got false")
	}
}

func assertFalse(tb testing.TB, v bool) {
	if v {
		tb.Fatal("expected false, but got true")
	}
}

func assertNil(tb testing.TB, v interface{}) {
	if v != nil && !reflect.ValueOf(v).IsNil() {
		tb.Fatalf("expected nil, but got: %+v", v)
	}
}

func assertPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	f()
}
