package requests

import (
	"database/sql"
	"fmt"
	"reflect"
)

// InsertStructs inserts a slice of structures into the given table
func (rq Request) InsertStructs(slice interface{}) error {

	// Check that passed value is a slice
	elem := reflect.ValueOf(slice)
	if elem.Kind() != reflect.Slice {
		return fmt.Errorf("passed value should be a slice of structures, got: %s", elem.Type().String())
	}

	// Create insert query
	if elem.Len() == 0 {
		return fmt.Errorf("passed slice is empty")
	}
	if err := rq.prepareInsert(elem.Type().Elem()); err != nil {
		return err
	}

	// prepare statement
	stmt, err := rq.PrepareStmt()
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Insert each structure
	for i := 0; i < elem.Len(); i++ {
		if err := insertStruct(stmt, elem.Index(i)); err != nil {
			return err
		}
	}
	return nil
}

// InsertOneStruct inserts a single structure into the given table
func (rq Request) InsertOneStruct(structure interface{}) error {

	// Check that passed value is a structure
	elem := reflect.ValueOf(structure)
	if elem.Kind() != reflect.Struct {
		return fmt.Errorf("passed value should be a structure, got: %s", elem.Type().String())
	}

	// Create insert query
	if err := rq.prepareInsert(elem.Type()); err != nil {
		return err
	}

	// Prepare statement
	stmt, err := rq.PrepareStmt()
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Insert structure
	return insertStruct(stmt, elem)
}

func (rq *Request) prepareInsert(elem reflect.Type) error {
	if elem.NumField() == 0 {
		return fmt.Errorf("passed structure has no field")
	}

	columns := "("
	values := "VALUES ("
	for i := 0; i < elem.NumField(); i++ {
		for i := 0; i < elem.NumField(); i++ {
			f := elem.Field(i)
			if f.Type.Kind() == reflect.Struct {
				for i := 0; i < f.Type.NumField(); i++ {
					columns += f.Type.Field(i).Tag.Get("db")
				}
			} else {
				columns += elem.Field(i).Tag.Get("db")
				values += "?"
			}
		}
		if i < elem.NumField()-1 {
			columns += ", "
			values += ", "
		}
	}
	columns += ")"
	values += ")"
	(*rq).Statement = "INSERT INTO"
	(*rq).Set = columns
	(*rq).Condition = values
	return nil
}

func insertStruct(stmt *sql.Stmt, elem reflect.Value) error {

	if elem.Kind() != reflect.Struct {
		return fmt.Errorf("passed value should be a structure, got: %s", elem.Type().String())
	}

	values := make([]interface{}, elem.NumField())
	for i := range values {
		v := elem.Field(i)
		if v.Kind() != reflect.Ptr {
			values[i] = v.Interface()
		} else if v.IsNil() {
			values[i] = nil
		} else {
			values[i] = v.Elem().Interface()
		}
	}
	_, err := stmt.Exec(values...)
	return err
}
