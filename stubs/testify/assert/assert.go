package assert

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func Equal(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected %v but got %v %s", expected, actual, formatMessage(msgAndArgs))
		return false
	}
	return true
}

func NotNil(t *testing.T, obj interface{}, msgAndArgs ...interface{}) bool {
	if isNil(obj) {
		t.Errorf("expected value not to be nil %s", formatMessage(msgAndArgs))
		return false
	}
	return true
}

func Nil(t *testing.T, obj interface{}, msgAndArgs ...interface{}) bool {
	if !isNil(obj) {
		t.Errorf("expected value to be nil but got %v %s", obj, formatMessage(msgAndArgs))
		return false
	}
	return true
}

func NoError(t *testing.T, err error, msgAndArgs ...interface{}) bool {
	if err != nil {
		t.Errorf("expected no error but got %v %s", err, formatMessage(msgAndArgs))
		return false
	}
	return true
}

func Error(t *testing.T, err error, msgAndArgs ...interface{}) bool {
	if err == nil {
		t.Errorf("expected error but got nil %s", formatMessage(msgAndArgs))
		return false
	}
	return true
}

func True(t *testing.T, value bool, msgAndArgs ...interface{}) bool {
	if !value {
		t.Errorf("expected true but got false %s", formatMessage(msgAndArgs))
		return false
	}
	return true
}

func False(t *testing.T, value bool, msgAndArgs ...interface{}) bool {
	if value {
		t.Errorf("expected false but got true %s", formatMessage(msgAndArgs))
		return false
	}
	return true
}

func Contains(t *testing.T, s interface{}, contains interface{}, msgAndArgs ...interface{}) bool {
	str := fmt.Sprintf("%v", s)
	substr := fmt.Sprintf("%v", contains)
	if !strings.Contains(str, substr) {
		t.Errorf("expected %q to contain %q %s", str, substr, formatMessage(msgAndArgs))
		return false
	}
	return true
}

func EqualError(t *testing.T, err error, errString string, msgAndArgs ...interface{}) bool {
	if err == nil {
		t.Errorf("expected error %q but got nil %s", errString, formatMessage(msgAndArgs))
		return false
	}
	if err.Error() != errString {
		t.Errorf("expected error %q but got %q %s", errString, err.Error(), formatMessage(msgAndArgs))
		return false
	}
	return true
}

func IsType(t *testing.T, expectedType interface{}, actual interface{}, msgAndArgs ...interface{}) bool {
	if reflect.TypeOf(expectedType) != reflect.TypeOf(actual) {
		t.Errorf("expected type %T but got %T %s", expectedType, actual, formatMessage(msgAndArgs))
		return false
	}
	return true
}

func ErrorIs(t *testing.T, err error, target error, msgAndArgs ...interface{}) bool {
	if !errors.Is(err, target) {
		t.Errorf("expected error %v to match %v %s", err, target, formatMessage(msgAndArgs))
		return false
	}
	return true
}

func NotEqual(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	if reflect.DeepEqual(expected, actual) {
		t.Errorf("expected values to differ but both were %v %s", expected, formatMessage(msgAndArgs))
		return false
	}
	return true
}

func NotSame(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	v1 := reflect.ValueOf(expected)
	v2 := reflect.ValueOf(actual)
	if v1.IsValid() && v2.IsValid() && v1.Kind() == reflect.Pointer && v2.Kind() == reflect.Pointer {
		if v1.Pointer() == v2.Pointer() {
			t.Errorf("expected pointers to refer to different objects %s", formatMessage(msgAndArgs))
			return false
		}
		return true
	}
	return NotEqual(t, expected, actual, msgAndArgs...)
}

func Greater(t *testing.T, e1, e2 interface{}, msgAndArgs ...interface{}) bool {
	f1, ok1 := toFloat64(e1)
	f2, ok2 := toFloat64(e2)
	if !ok1 || !ok2 {
		t.Errorf("greater comparison only supports numeric types %s", formatMessage(msgAndArgs))
		return false
	}
	if !(f1 > f2) {
		t.Errorf("expected %v to be greater than %v %s", e1, e2, formatMessage(msgAndArgs))
		return false
	}
	return true
}

func NotEmpty(t *testing.T, object interface{}, msgAndArgs ...interface{}) bool {
	v := reflect.ValueOf(object)
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		if v.Len() == 0 {
			t.Errorf("expected value to be non-empty %s", formatMessage(msgAndArgs))
			return false
		}
	default:
		if isNil(object) {
			t.Errorf("expected value to be non-empty %s", formatMessage(msgAndArgs))
			return false
		}
	}
	return true
}

func formatMessage(msgAndArgs []interface{}) string {
	if len(msgAndArgs) == 0 {
		return ""
	}
	return fmt.Sprintf("- %v", msgAndArgs[0])
}

func isNil(i interface{}) bool {
	if i == nil {
		return true
	}
	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	}
	return false
}

func toFloat64(v interface{}) (float64, bool) {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(rv.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return float64(rv.Uint()), true
	case reflect.Float32, reflect.Float64:
		return rv.Float(), true
	}
	return 0, false
}
