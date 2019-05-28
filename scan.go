package requests

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
)

type scanner interface {
	Scan(dest ...interface{}) error
}

// Scan row(s) into slice of interface
func scanRows(row scanner, size int) ([]interface{}, error) {
	results := make([]interface{}, size)
	resultsPtr := make([]interface{}, size)
	for i := range results {
		resultsPtr[i] = &(results[i])
	}
	err := row.Scan(resultsPtr...)
	return results, err
}

// Scan and store results into pointed slice of structures
func scanIntoStructs(rows *sql.Rows, ptr interface{}) error {

	// Check that passed value is a pointer to a slice of structures
	v := reflect.ValueOf(ptr)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("passed value should be a pointer to a slice of structures, got: %s", v.Type().String())
	}
	elem := v.Elem()
	if elem.Kind() != reflect.Slice {
		return fmt.Errorf("pointed value should be a slice of structures, got: %s", elem.Type().String())
	}
	sliceElem := elem.Type().Elem()
	if sliceElem.Kind() != reflect.Struct {
		return fmt.Errorf("pointed slice should store structures, got: %s", sliceElem.String())
	}
	// Get zeroed structure
	zeroedSt := reflect.Zero(sliceElem)

	// Store each row
	var i int
	for rows.Next() {
		// Scan row
		results, err := scanRows(rows, zeroedSt.NumField())
		if err != nil {
			return err
		}
		// Append slice
		if !elem.CanSet() {
			return errors.New("structure from pointed slice is not settable")
		}
		elem.Set(reflect.Append(elem, zeroedSt))
		// Store results
		if err := scanIntoStruct(elem.Index(i), results); err != nil {
			return err
		}
		i++
	}
	return nil
}

// Scan and store results into pointed structure
func scanToOneStruct(row *sql.Row, ptr interface{}) error {

	// Check that passed value is a pointer to a structure
	v := reflect.ValueOf(ptr)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("passed value should be a pointer to structure, got: %s", v.Type().String())
	}
	elem := v.Elem()
	if elem.Kind() != reflect.Struct {
		return fmt.Errorf("pointed value should be a structure, got: %s", elem.Type().String())
	}

	// Scan into slice of interface
	results, err := scanRows(row, elem.NumField())
	if err != nil {
		return err
	}
	return scanIntoStruct(elem, results)
}

// Store scanned results to one structure
func scanIntoStruct(st reflect.Value, results []interface{}) error {
	for i, result := range results {
		f := st.Field(i)
		if err := storeToField(f, result, i+1); err != nil {
			return err
		}
	}
	return nil
}

// Scan and store results into pointed value
func scanIntoOnePtr(row *sql.Row, ptr interface{}) error {
	// Scan into interface
	var result interface{}
	err := row.Scan(&result)
	if err != nil {
		return err
	}
	// Store scanned result to pointed value
	return scanIntoPtr(reflect.ValueOf(ptr), result, 1)
}

// Scan and store results into slice of pointed values
func scanIntoPtrs(row *sql.Row, slice interface{}) error {

	// Check that passed value is a slice
	elem := reflect.ValueOf(slice)
	if elem.Kind() != reflect.Slice {
		return fmt.Errorf("passed value must be a slice of pointers, got: %s", elem.Type().String())
	}
	// Scan into slice of interface
	results, err := scanRows(row, elem.Len())
	if err != nil {
		return err
	}

	// Store scanned results into pointed values
	for i, result := range results {
		f := elem.Index(i)
		if f.Kind() != reflect.Interface {
			return fmt.Errorf("passed slice should store pointers in interface{} form, got: %s", f.Type().String())
		}
		pt := f.Elem()
		if err := scanIntoPtr(pt, result, i+1); err != nil {
			return err
		}
	}
	return nil
}

// Store scanned result to pointed value
func scanIntoPtr(ptr reflect.Value, result interface{}, index int) error {
	// Check that passed value is a pointer
	if ptr.Kind() != reflect.Ptr {
		return fmt.Errorf("passed value should be a pointer, got: %s", ptr.Type().String())
	}
	elem := ptr.Elem()
	// Store result
	return storeToField(elem, result, index)
}

// Scan and store results into slice of any type
func scanIntoSlice(rows *sql.Rows, ptr interface{}) error {

	// Check that passed value is a pointer to a slice of any type
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
