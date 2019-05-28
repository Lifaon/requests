package requests

import (
	"fmt"
	"reflect"
	"time"
)

// Store result to passed value
func storeToField(v reflect.Value, result interface{}, i int) error {
	// Check that passed element is settable
	if !v.CanSet() {
		return fmt.Errorf("field #%d isn't settable", i)
	}
	// Set element
	switch val := result.(type) {
	case []byte:
		return storeBytes(v, val, i)
	case int64:
		return storeInt(v, val, i)
	case float64:
		return storeFloat(v, val, i)
	case time.Time:
		return storeTime(v, val, i)
	case nil:
		return storeNil(v, i)
	default:
		// Unsupported type
		return fmt.Errorf("unsupported type retrieved from *sql.Row(s).Scan(): %T", val)
	}
}

// Type of elem: []byte, *[]byte, string, *string, bool, or *bool
func storeBytes(v reflect.Value, result []byte, i int) error {
	s := string(result)
	b := len(result) != 0 && result[0] != 0
	switch v.Type() {
	case reflect.TypeOf(result): // []byte
		v.SetBytes(result)
	case reflect.TypeOf(&result): // *[]byte
		ptr := new([]byte)
		*ptr = result
		v.Set(reflect.ValueOf(ptr))
	case reflect.TypeOf(s): // string
		v.SetString(s)
	case reflect.TypeOf(&s): // *string
		ptr := new(string)
		*ptr = s
		v.Set(reflect.ValueOf(ptr))
	case reflect.TypeOf(b): // bool
		v.SetBool(b)
	case reflect.TypeOf(&b): // *bool
		ptr := new(bool)
		*ptr = b
		v.Set(reflect.ValueOf(ptr))
	default:
		return fmt.Errorf("field #%d doesn't have the right type (expected: []byte, *[]byte, string, *string, bool, or *bool, got: %s)", i, v.Type().String())
	}
	return nil
}

// Type of elem: int64 or *int64
func storeInt(v reflect.Value, result int64, i int) error {
	switch v.Type() {
	case reflect.TypeOf(result): // int64
		v.SetInt(result)
	case reflect.TypeOf(&result): // *int64
		ptr := new(int64)
		*ptr = result
		v.Set(reflect.ValueOf(ptr))
	default:
		return fmt.Errorf("field #%d doesn't have the right type (expected: int64 or *int64, got: %s)", i, v.Type().String())
	}
	return nil
}

// Type of elem: float64 or *float64
func storeFloat(v reflect.Value, result float64, i int) error {
	switch v.Type() {
	case reflect.TypeOf(result): // float64
		v.SetFloat(result)
	case reflect.TypeOf(&result): // *float64
		ptr := new(float64)
		*ptr = result
		v.Set(reflect.ValueOf(ptr))
	default:
		return fmt.Errorf("field #%d doesn't have the right type (expected: float64 or *float64, got: %s)", i, v.Type().String())
	}
	return nil
}

// Type of elem: time.Time or *time.Time
func storeTime(v reflect.Value, result time.Time, i int) error {
	switch v.Type() {
	case reflect.TypeOf(result): // time.Time
		v.Set(reflect.ValueOf(result))
	case reflect.TypeOf(&result): // *time.Time
		ptr := new(time.Time)
		*ptr = result
		v.Set(reflect.ValueOf(ptr))
	default:
		return fmt.Errorf("field #%d doesn't have the right type (expected: time.Time or *time.Time, got: %s)", i, v.Type().String())
	}
	return nil
}

// Type of elem: any pointer
func storeNil(v reflect.Value, i int) error {
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("field #%d isn't a pointer when result can be <nil>, got: %s", i, v.Type().String())
	}
	v.Set(reflect.Zero(v.Type()))
	return nil
}
