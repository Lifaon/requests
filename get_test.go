package requests_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

const (
	select_one_field = "SELECT id FROM user"
)

func getStructRows() *sqlmock.Rows {
	return sqlmock.NewRows(columns).
		AddRow([]byte(param_a), time_now).AddRow([]byte(param_b), time_now)
}

func getSingleFieldRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{col_1}).
		AddRow([]byte(param_a)).AddRow([]byte(param_b))
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
	assert.NoError(t, err, unexpected_error)
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
	assert.NoError(t, err, unexpected_error)
	checkResults(t, mock)
}

// Test GetIntoSlice method
func TestGetIntoSlice(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	var ids []string
	rq.Query = select_one_field
	rq.Arg = &ids
	mock.ExpectPrepare(select_one_field).WillBeClosed().
		ExpectQuery().WillReturnRows(getSingleFieldRows()).RowsWillBeClosed()

	err := rq.GetIntoSlice()
	assert.NoError(t, err, unexpected_error)
	checkResults(t, mock)
}

// Test GetOneField method
func TestGetOneField(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	var id string
	rq.Query = select_one_field
	rq.Arg = &id
	mock.ExpectPrepare(select_one_field).WillBeClosed().
		ExpectQuery().WillReturnRows(getSingleFieldRows()).RowsWillBeClosed()

	err := rq.GetOneField()
	assert.NoError(t, err, unexpected_error)
	checkResults(t, mock)
}
