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
	param_a = int64(42)
	param_b = float64(3.14)

	// Columns
	col_1 = "id"
	col_2 = "createdAt"

	// Queries
	select_query = "SELECT"
	insert_query = "INSERT"
	update_query = "UPDATE"
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
	assert.NoError(t, err)
}

// Test QueryStruct's String method
func TestQueryStructString(t *testing.T) {

	const query = "UPDATE user SET id = '12345' WHERE name = 'test'"
	var rq1, rq2 requests.Request

	// From sub parameters
	rq1.Statement = "UPDATE"
	rq1.Table = "user"
	rq1.Set = "SET id = '12345'"
	rq1.Condition = "WHERE name = 'test'"
	assert.Equal(t, query, rq1.QueryStruct.String())

	// From whole query
	rq2.Query = query
	assert.Equal(t, query, rq2.QueryStruct.String())
}

// Test PrepareStmt method
func TestPrepareStmt(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	rq.Query = select_query
	mock.ExpectPrepare(rq.Query).WillBeClosed()

	stmt, err := rq.PrepareStmt()
	assert.NoError(t, err)
	if assert.NotNil(t, stmt) {
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
	assert.NoError(t, err)
	if assert.NotNil(t, rows) {
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
	assert.Error(t, err)
	if !assert.Nil(t, rows) {
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
	assert.NoError(t, err)
	if assert.NotNil(t, row) {
		var a, b interface{}
		err := row.Scan(&a, &b)
		assert.NoError(t, err)
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
	assert.Error(t, err)
	if !assert.Nil(t, row) {
		var a, b interface{}
		err := row.Scan(&a, &b)
		assert.Error(t, err)
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
	if assert.NoError(t, err) {
		res1, _ := res.LastInsertId()
		res2, _ := res.RowsAffected()
		ret := sqlmock.NewResult(res1, res2)
		assert.Equal(t, result_1, ret)
	}
	checkResults(t, mock)
}

// Test ExecQuery method, but with invalid Query
func TestExecQueryWrongQuery(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	rq.Query = update_query
	mock.ExpectPrepare(rq.Query).WillReturnError(sql_error)

	_, err := rq.ExecQuery(param_a, param_b)
	assert.Error(t, err)
	checkResults(t, mock)
}
