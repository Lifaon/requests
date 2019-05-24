package requests_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type (
	BasicFields struct {
		ID        string `db:"id"`
		CreatedAt string `db:"createdAt"`
	}

	TestStruct struct {
		BasicFields `req:"include"`

		Ptr    *bool `db:"ptr"`
		NilPtr *bool `db:"nilptr"`

		ShouldBeIgnored bool `db:"-"`
		NoTag           bool
	}

	EmptyStruct struct {
	}
)

const (
	insert_query_regex = "INSERT INTO user \\(id, createdAt, ptr, nilptr\\) VALUES \\(\\?, \\?, \\?, \\?\\)"
)

var (
	ptr = &[]bool{true}[0]

	oneStruct = TestStruct{
		BasicFields: BasicFields{
			ID:        param_a,
			CreatedAt: param_b,
		},
		Ptr:    ptr,
		NilPtr: nil,
	}
	oneEmptyStruct = EmptyStruct{}

	structs = []TestStruct{
		oneStruct, oneStruct, oneStruct,
	}
	emptyStructs = []EmptyStruct{
		oneEmptyStruct, oneEmptyStruct, oneEmptyStruct,
	}

	intSlice = []int64{42, 101, 1337}
)

// Test InsertStructs method
func TestInsertStructs(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	rq.Table = "user"
	mock.ExpectPrepare(insert_query_regex).WillBeClosed()
	mock.ExpectExec(insert_query_regex).WithArgs(param_a, param_b, ptr, nil).
		WillReturnResult(result_1)
	mock.ExpectExec(insert_query_regex).WithArgs(param_a, param_b, ptr, nil).
		WillReturnResult(result_1)
	mock.ExpectExec(insert_query_regex).WithArgs(param_a, param_b, ptr, nil).
		WillReturnResult(result_1)

	err := rq.InsertStructs(structs)
	assert.NoError(t, err, unexpected_error)
	checkResults(t, mock)
}

// Test InsertStructs method, with parameter not being a slice
func TestInsertStructsNotASlice(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	rq.Table = "user"

	err := rq.InsertStructs(oneStruct)
	assert.Error(t, err, expected_error)
	checkResults(t, mock)
}

// Test InsertStructs method, with empty slice parameter
func TestInsertStructsEmptySlice(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	rq.Table = "user"

	err := rq.InsertStructs([]TestStruct{})
	assert.Error(t, err, expected_error)
	checkResults(t, mock)
}

// Test InsertStructs method, with slice parameter not storing structures
func TestInsertStructsNotASliceOfStructures(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	rq.Table = "user"

	err := rq.InsertStructs(intSlice)
	assert.Error(t, err, expected_error)
	checkResults(t, mock)
}

// Test InsertStructs method, with empty structures in slice parameter
func TestInsertStructsEmptyStructures(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	rq.Table = "user"

	err := rq.InsertStructs(emptyStructs)
	assert.Error(t, err, expected_error)
	checkResults(t, mock)
}

// Test InsertOneStruct method
func TestInsertOneStruct(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	rq.Table = "user"
	mock.ExpectPrepare(insert_query_regex).WillBeClosed().
		ExpectExec().WithArgs(param_a, param_b, ptr, nil).WillReturnResult(result_1)

	err := rq.InsertOneStruct(oneStruct)
	assert.NoError(t, err, unexpected_error)
	checkResults(t, mock)
}

// Test InsertOneStruct method, but with parameter not being a structure
func TestInsertOneStructNotAStructure(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	rq.Table = "user"

	err := rq.InsertOneStruct(structs)
	assert.Error(t, err, expected_error)
	checkResults(t, mock)
}

// Test InsertOneStruct method, but with empty structure parameter
func TestInsertOneStructEmptyStructure(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	rq.Table = "user"

	err := rq.InsertOneStruct(oneEmptyStruct)
	assert.Error(t, err, expected_error)
	checkResults(t, mock)
}
