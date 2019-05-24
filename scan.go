package requests

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"time"
)

// Scan and store results into pointed structure
func scanToOneStruct(row *sql.Row, ptr interface{}) error {

	v := reflect.ValueOf(ptr)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("passed value should be a pointer to structure, got: %s", v.Type().String())
	}
	elem := v.Elem()
	if elem.Kind() != reflect.Struct {
		return fmt.Errorf("pointed value should be a structure, got: %s", elem.Type().String())
	}

	// Scan into slice of interface
	results := make([]interface{}, elem.NumField())
	resultsPtr := make([]interface{}, elem.NumField())
	for i := range results {
		resultsPtr[i] = &(results[i])
	}
	err := row.Scan(resultsPtr...)
	if err != nil {
		return err
	}
	return storeIntoStruct(elem, results)
}

// Scan and store results into pointed slice of structures
func scanToSliceOfStruct(rows *sql.Rows, ptr interface{}) error {

	// Check that passed value is a pointer to a slice of structures
	v := reflect.ValueOf(ptr)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("passed value should be a pointer to a slice of structures, got: %s", v.Type().String())
	}
	elem := v.Elem()
	if elem.Kind() != reflect.Slice {
		return fmt.Errorf("pointed value should be a slice of structures, got: %s", elem.Type().String())
	}
	sliceType := elem.Type().Elem()
	if sliceType.Kind() != reflect.Struct {
		return fmt.Errorf("pointed slice should store structures, got: %s", sliceType.String())
	}

	// Get zeroed structure
	zeroedSt := reflect.Zero(sliceType)

	// Make slice of interfaces to scan rows
	size := zeroedSt.NumField()
	results := make([]interface{}, size)
	resultsPtr := make([]interface{}, size)
	for i := range results {
		resultsPtr[i] = &(results[i])
	}

	// Store each row
	var i int
	for rows.Next() {
		// Scan row
		if err := rows.Scan(resultsPtr...); err != nil {
			return err
		}
		// Append slice
		if !elem.CanSet() {
			return errors.New("structure from pointed slice is not settable")
		}
		elem.Set(reflect.Append(elem, zeroedSt))
		// Store results
		if err := storeIntoStruct(elem.Index(i), results); err != nil {
			return err
		}
		i++
	}
	return nil
}

// Store scanned results to one structure
func storeIntoStruct(st reflect.Value, results []interface{}) error {
	for i, result := range results {
		f := st.Field(i)
		if err := storeToField(f, result, i+1); err != nil {
			return err
		}
	}
	return nil
}

// Scan and store results into pointed value
func scanToOnePtr(row *sql.Row, ptr interface{}) error {
	// Scan into interface
	var result interface{}
	err := row.Scan(&result)
	if err != nil {
		return err
	}
	// Store scanned result to pointed value
	return storeIntoPtr(reflect.ValueOf(ptr), result, 1)
}

// Scan and store results into slice of pointed values
func scanToSliceOfPtr(row *sql.Row, slice interface{}) error {
	// Check that passed value is a slice
	elem := reflect.ValueOf(slice)
	if elem.Kind() != reflect.Slice {
		return fmt.Errorf("passed value must be a slice of pointers, got: %s", elem.Type().String())
	}
	// Scan into slice of interface
	results := make([]interface{}, elem.Len())
	resultsPtr := make([]interface{}, elem.Len())
	for i := range results {
		resultsPtr[i] = &(results[i])
	}
	err := row.Scan(resultsPtr...)
	if err != nil {
		return err
	}
	// Store scanned results into pointed values
	for i, result := range results {
		pt := elem.Index(i)
		if err := storeIntoPtr(pt, result, i+1); err != nil {
			return err
		}
	}
	return nil
}

// Scan and store results into slice of a single value
func scanToSlice(rows *sql.Rows, ptr interface{}) error {

	// Check that passed value is a pointer to a slice of structures
	v := reflect.ValueOf(ptr)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("passed value should be a pointer to a slice, got: %s", v.Type().String())
	}
	elem := v.Elem()
	if elem.Kind() != reflect.Slice {
		return fmt.Errorf("pointed value should be a slice, got: %s", elem.Type().String())
	}

	// Get zeroed field
	zeroedField := reflect.Zero(elem.Type().Elem())

	// Store each row
	var i int
	for rows.Next() {
		// Scan row
		var result interface{}
		if err := rows.Scan(&result); err != nil {
			return err
		}
		// Append slice
		if !elem.CanSet() {
			return errors.New("value from pointed slice is not settable")
		}
		elem.Set(reflect.Append(elem, zeroedField))
		// Store results
		if err := storeToField(elem.Index(i), result, i); err != nil {
			return err
		}
		i++
	}
	return nil
}

// Store scanned result to pointed value
func storeIntoPtr(ptr reflect.Value, result interface{}, index int) error {
	// Check that passed value is a pointer
	if ptr.Kind() != reflect.Ptr {
		return fmt.Errorf("passed value should be a pointer to structure, got: %s", ptr.Type().String())
	}
	elem := ptr.Elem()
	// Store result
	return storeToField(elem, result, index)
}

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
