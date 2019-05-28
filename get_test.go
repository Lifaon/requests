package requests_test

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func getStructRows() *sqlmock.Rows {
	return sqlmock.NewRows(columns).
		AddRow(param_a, time_now).AddRow(param_a, time_now)
}

func getSingleFieldRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{col_1}).
		AddRow(param_a).AddRow(param_a)
}

// Test GetIntoStructs method
func TestGetIntoStructs(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	var users []BasicFields
	rq.Query = select_query
	rq.Arg = &users
	mock.ExpectPrepare(select_query).WillBeClosed().
		ExpectQuery().WillReturnRows(getStructRows()).RowsWillBeClosed()

	err := rq.GetIntoStructs()
	assert.NoError(t, err)
	checkResults(t, mock)
}

// Test GetIntoStructs method, but without argument to store into
func TestGetIntoStructsNoArg(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	err := rq.GetIntoStructs()
	assert.Error(t, err)
	checkResults(t, mock)
}

// Test GetIntoOneStruct method
func TestGetIntoOneStruct(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	var user BasicFields
	rq.Query = select_query
	rq.Arg = &user
	mock.ExpectPrepare(select_query).WillBeClosed().
		ExpectQuery().WillReturnRows(getStructRows()).RowsWillBeClosed()

	err := rq.GetIntoOneStruct()
	assert.NoError(t, err)
	checkResults(t, mock)
}

// Test GetIntoOneStruct method, but without argument to store into
func TestGetIntoOneStructNoArg(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	err := rq.GetIntoOneStruct()
	assert.Error(t, err)
	checkResults(t, mock)
}

// Test GetIntoSlice method
func TestGetIntoSlice(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	var ids []int64
	rq.Query = select_query
	rq.Arg = &ids
	mock.ExpectPrepare(select_query).WillBeClosed().
		ExpectQuery().WillReturnRows(getSingleFieldRows()).RowsWillBeClosed()

	err := rq.GetIntoSlice()
	assert.NoError(t, err)
	checkResults(t, mock)
}

// Test GetIntoSlice method, but without argument to store into
func TestGetIntoSliceNoArg(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	err := rq.GetIntoSlice()
	assert.Error(t, err)
	checkResults(t, mock)
}

// Test GetOneField method
func TestGetOneField(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	var id int64
	rq.Query = select_query
	rq.Arg = &id
	mock.ExpectPrepare(select_query).WillBeClosed().
		ExpectQuery().WillReturnRows(getSingleFieldRows()).RowsWillBeClosed()

	err := rq.GetOneField()
	assert.NoError(t, err)
	checkResults(t, mock)
}

// Test GetOneField method, but without argument to store into
func TestGetOneFieldNoArg(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	err := rq.GetOneField()
	assert.Error(t, err)
	checkResults(t, mock)
}

// Test GetFields method
func TestGetFields(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	var id int64
	var createdAt time.Time
	rq.Query = select_query
	rq.Arg = []interface{}{
		&id,
		&createdAt,
	}
	mock.ExpectPrepare(select_query).WillBeClosed().
		ExpectQuery().WillReturnRows(getStructRows()).RowsWillBeClosed()

	err := rq.GetFields()
	assert.NoError(t, err)
	checkResults(t, mock)
}

// Test GetFields method, but without argument to store into
func TestGetFieldsNoArg(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	err := rq.GetFields()
	assert.Error(t, err)
	checkResults(t, mock)
}
