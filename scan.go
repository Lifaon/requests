package requests

import (
	"database/sql"
	"fmt"
	"reflect"
	"time"

	"github.com/mlantonn/WSK_Watcher/utils"
)

// scan and store values into pointed structure
func scanOneRow(row *sql.Row, ptr interface{}) error {

	v := reflect.ValueOf(ptr)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("passed value should be a pointer to structure, got: %s", utils.GetReflectType(v))
	}
	elem := v.Elem()
	if elem.Kind() != reflect.Struct {
		return fmt.Errorf("pointed value should be a structure, got: %s", utils.GetReflectType(elem))
	}

	// Scan into slice of interface
	values := make([]interface{}, elem.NumField())
	valuesptr := make([]interface{}, elem.NumField())
	for i := range values {
		valuesptr[i] = &(values[i])
	}
	err := row.Scan(valuesptr...)
	if err != nil {
		return err
	}
	return storeToStruct(elem, values)
}

func scanRows(rows *sql.Rows, ptr interface{}) error {

	v := reflect.ValueOf(ptr)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("passed value should be a pointer to slice of structures, got: %s", utils.GetReflectType(v))
	}
	elem := v.Elem()
	if elem.Kind() != reflect.Slice {
		return fmt.Errorf("pointed value should be a slice of structures, got: %s", utils.GetReflectType(elem))
	}
	if !elem.CanSet() {
		return fmt.Errorf("pointed value (of type %s) is not settable", utils.GetReflectType(elem))
	}

	if !rows.Next() {
		return nil
	}

	elem.Set(reflect.MakeSlice(elem.Type(), 1, 1))

	size := elem.Index(0).NumField()
	values := make([]interface{}, size)
	valuesptr := make([]interface{}, size)
	for i := range values {
		valuesptr[i] = &(values[i])
	}

	for i := 0; ; i++ {
		f := elem.Index(i)
		if !f.CanSet() {
			return fmt.Errorf("index #%d of created slice (of type %s) is not settable", i, utils.GetReflectType(elem))
		}
		err := rows.Scan(valuesptr...)
		if err != nil {
			return err
		}
		err = storeToStruct(f, values)
		if err != nil {
			return err
		}
		if rows.Next() {
			elem.Set(reflect.Append(elem, f))
		} else {
			break
		}
	}
	return nil
}

// store scanned values to structure
func storeToStruct(st reflect.Value, values []interface{}) error {
	for i, value := range values {
		// Retrieve structure field
		f := st.Field(i)
		if !f.CanSet() {
			return fmt.Errorf("structure field #%d of type %s is not settable", i, utils.GetReflectType(f))
		}
		// Set structure field
		switch val := value.(type) {
		case []byte:
			// struct field: []byte, *[]byte, string, *string, bool, or *bool
			s := string(val)
			b := false
			if len(val) != 0 && val[0] != 0 {
				b = true
			}
			switch f.Type() {
			case reflect.TypeOf(val): // []byte
				f.SetBytes(val)
			case reflect.TypeOf(&val): // *[]byte
				f.Set(reflect.ValueOf(&val))
			case reflect.TypeOf(s): // string
				f.SetString(s)
			case reflect.TypeOf(&s): // *string
				f.Set(reflect.ValueOf(&s))
			case reflect.TypeOf(b): // bool
				f.SetBool(b)
			case reflect.TypeOf(&b): // *bool
				f.Set(reflect.ValueOf(&b))
			default:
				return fmt.Errorf("structure field #%d doesn't have the right type (expected: []byte, *[]byte, string, *string, bool, or *bool, got: %s)", i, utils.GetReflectType(f))
			}
		case int64:
			// struct field: int64 or *int64
			switch f.Type() {
			case reflect.TypeOf(val): // int64
				f.SetInt(val)
			case reflect.TypeOf(&val): // *int64
				f.Set(reflect.ValueOf(&val))
			default:
				return fmt.Errorf("structure field #%d doesn't have the right type (expected: int64 or *int64, got: %s)", i, utils.GetReflectType(f))
			}
		case float64:
			// struct field: float64 or *float64
			switch f.Type() {
			case reflect.TypeOf(val): // float64
				f.SetFloat(val)
			case reflect.TypeOf(&val): // *float64
				f.Set(reflect.ValueOf(&val))
			default:
				return fmt.Errorf("structure field #%d doesn't have the right type (expected: float64 or *float64, got: %s)", i, utils.GetReflectType(f))
			}
		case time.Time:
			// struct field: time.Time or *time.Time
			switch f.Type() {
			case reflect.TypeOf(val): // time.Time
				f.Set(reflect.ValueOf(val))
			case reflect.TypeOf(&val): // *time.Time
				f.Set(reflect.ValueOf(&val))
			default:
				return fmt.Errorf("structure field #%d doesn't have the right type (expected: time.Time or *time.Time, got: %s)", i, utils.GetReflectType(f))
			}
		case nil:
			// struct field: any pointer
			if f.Kind() == reflect.Ptr {
				f.Set(reflect.Zero(f.Type()))
			} else {
				return fmt.Errorf("structure field #%d isn't a pointer when value can be <nil>, got: %s", i, utils.GetReflectType(f))
			}
		default:
			// unsupported type retrieved from *sql.Row(s).Scan()
			return fmt.Errorf("unsupported type retrieved from *sql.Row(s).Scan(): %T", val)
		}
	}
	return nil
}
