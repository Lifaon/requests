package requests_test

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mlantonn/requests"
	"github.com/stretchr/testify/assert"
)

const (
	// Parameters
	param_a = "first_param"
	param_b = "second_param"

	// Columns
	col_1 = "id"
	col_2 = "createdAt"

	// Queries
	select_query = "SELECT id, createdAt FROM user"
	update_query = "UPDATE user SET id = '12345' WHERE name = 'test'"

	// Error messages
	expected_error   = "expected an error"
	unexpected_error = "did not expect an error"
	nil_error        = "expected result to be nil"
	not_nil_error    = "did not expect result to be nil"
	equal_error      = "result is different from expected"
)

var (
	// SQL results
	result_0 = sqlmock.NewResult(0, 0)
	result_1 = sqlmock.NewResult(1, 1)

	// Errors
	sql_error = errors.New("this would produce an error")

	// SQL Rows
	columns  = []string{col_1, col_2}
	time_now = time.Now()
)

func getSelectRows() *sqlmock.Rows {
	return sqlmock.NewRows(columns).
		AddRow(param_a, time_now).AddRow(param_b, time_now)
}

// Init db, mock & rq structures
func initRequest(t *testing.T) (requests.Request, sqlmock.Sqlmock, *sql.DB) {
	db, mock, err := sqlmock.New()
	if !assert.NoErrorf(t, err, "Could not initialize DB") {
		t.FailNow()
	}
	rq := requests.FromHandler(db)
	return rq, mock, db
}

// Check that expectations were met
func checkResults(t *testing.T, mock sqlmock.Sqlmock) {
	err := mock.ExpectationsWereMet()
	assert.NoErrorf(t, err, "Expectiations were not met")
}

// Test QueryStruct's String method
func TestQueryStructString(t *testing.T) {

	// From sub parameters
	var rq requests.Request
	rq.Statement = "UPDATE"
	rq.Table = "user"
	rq.Set = "SET id = '12345'"
	rq.Condition = "WHERE name = 'test'"
	assert.Equal(t, update_query, rq.QueryStruct.String())

	// From whole query
	var rq2 requests.Request
	rq2.Query = update_query
	assert.Equal(t, update_query, rq2.QueryStruct.String())
}

// Test PrepareStmt method
func TestPrepareStmt(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	rq.Query = select_query
	mock.ExpectPrepare(rq.Query).WillBeClosed()

	stmt, err := rq.PrepareStmt()
	assert.NoError(t, err, unexpected_error)
	if assert.NotNil(t, stmt, not_nil_error) {
		stmt.Close()
	}
	checkResults(t, mock)
}

// Test GetRows method
func TestGetRows(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	rq.Query = select_query
	mock.ExpectPrepare(rq.Query).WillBeClosed().
		ExpectQuery().WillReturnRows(getSelectRows()).RowsWillBeClosed()

	rows, err := rq.GetRows()
	assert.NoError(t, err, unexpected_error)
	if assert.NotNil(t, rows, not_nil_error) {
		rows.Close()
	}
	checkResults(t, mock)
}

// Test GetRows method, but with invalid query
func TestGetRowsWrongQuery(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	rq.Query = select_query
	mock.ExpectPrepare(rq.Query).WillReturnError(sql_error)

	rows, err := rq.GetRows()
	assert.Error(t, err, expected_error)
	if !assert.Nil(t, rows, nil_error) {
		rows.Close()
	}
	checkResults(t, mock)
}

// Test GetOneRow method
func TestGetOneRow(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	rq.Query = select_query
	mock.ExpectPrepare(rq.Query).WillBeClosed().
		ExpectQuery().WillReturnRows(getSelectRows())

	row, err := rq.GetOneRow()
	assert.NoError(t, err, unexpected_error)
	if assert.NotNil(t, row, not_nil_error) {
		var a, b interface{}
		err := row.Scan(&a, &b)
		assert.NoError(t, err, unexpected_error)
	}
	checkResults(t, mock)
}

// Test GetOneRow method
func TestGetOneRowWrongQuery(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	rq.Query = select_query
	mock.ExpectPrepare(rq.Query).WillReturnError(sql_error)

	row, err := rq.GetOneRow()
	assert.Error(t, err, expected_error)
	if !assert.Nil(t, row, not_nil_error) {
		var a, b interface{}
		err := row.Scan(&a, &b)
		assert.Error(t, err, expected_error)
	}
	checkResults(t, mock)
}

// Test ExecQuery method
func TestExecQuery(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	rq.Query = update_query
	mock.ExpectPrepare(rq.Query).WillBeClosed().
		ExpectExec().WillReturnResult(result_1)

	res, err := rq.ExecQuery()
	assert.NoError(t, err, unexpected_error)
	res1, _ := res.LastInsertId()
	res2, _ := res.RowsAffected()
	ret := sqlmock.NewResult(res1, res2)
	assert.Equal(t, result_1, ret, equal_error)
	checkResults(t, mock)
}

// Test ExecQuery method, but with invalid Query
func TestExecQueryWrongQuery(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	rq.Query = update_query
	mock.ExpectPrepare(rq.Query).WillReturnError(sql_error)

	_, err := rq.ExecQuery(param_a, param_b)
	assert.Error(t, err, expected_error)
	checkResults(t, mock)
}
