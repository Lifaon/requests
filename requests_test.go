package requests_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mlantonn/requests"
	"github.com/stretchr/testify/assert"
)

const (
	a = "first"
	b = "second"
)

var (
	result_0 = sqlmock.NewResult(0, 0)
	result_1 = sqlmock.NewResult(1, 1)

	time_now = time.Now()
	mockRows = sqlmock.NewRows([]string{"id", "createdAt"}).
			AddRow(a, time_now).AddRow(b, time_now)
)

func initRequest(t *testing.T) (requests.Request, sqlmock.Sqlmock, *sql.DB) {
	db, mock, err := sqlmock.New()
	if !assert.NoErrorf(t, err, "Could not initialize DB: %v", err) {
		t.FailNow()
	}
	rq := requests.FromHandler(db)

	return rq, mock, db
}

func checkResults(t *testing.T, mock sqlmock.Sqlmock) {
	err := mock.ExpectationsWereMet()
	assert.NoErrorf(t, err, "Expectiations were not met")
}

func TestQueryStructString(t *testing.T) {
	const query = "UPDATE user SET id = '12345' WHERE name = 'test'"
	var rq requests.Request
	rq.Statement = "UPDATE"
	rq.Table = "user"
	rq.Set = "SET id = '12345'"
	rq.Condition = "WHERE name = 'test'"
	assert.Equal(t, query, rq.QueryStruct.String())

	var rq2 requests.Request
	rq2.Query = query
	assert.Equal(t, query, rq2.QueryStruct.String())
}

func TestPrepareStatement(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	rq.Query = "SELECT id FROM user"
	mock.ExpectPrepare(rq.Query).WillBeClosed()

	stmt, _ := rq.PrepareStmt()
	if stmt != nil {
		stmt.Close()
	}

	checkResults(t, mock)
}

func TestExecQuery(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	rq.Query = "INSERT INTO user"
	mock.ExpectPrepare(rq.Query).
		ExpectExec().WithArgs(a, b)
	rq.ExecQuery(a, b)

	checkResults(t, mock)
}

func TestExecQueryWrongQuery(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	rq.Query = "INSERT INTO user"

	_, err := rq.ExecQuery(a, b)
	assert.Error(t, err, "Prepare should return an error")
	checkResults(t, mock)
}

func TestExecQueryWrongParams(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	rq.Query = "INSERT INTO user"
	mock.ExpectPrepare(rq.Query).WillBeClosed()

	_, err := rq.ExecQuery(a)
	assert.Error(t, err, "Exec should return an error")
	checkResults(t, mock)
}

func TestGetOneRow(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	rq.Query = "SELECT id, createdAt FROM user"
	mock.ExpectPrepare(rq.Query).WillBeClosed().
		ExpectQuery()

	row, err := rq.GetOneRow()
	assert.NoError(t, err, "Query should work")
	assert.NotNil(t, row, "Returned row should not be nil")
	checkResults(t, mock)
}

func TestGetRows(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	tim := time.Now()
	rq.Query = "SELECT id FROM user"

	mock.ExpectPrepare("^SELECT (.+) FROM user$").WillBeClosed().
		ExpectQuery().WillReturnRows(mockRows).RowsWillBeClosed()

	res, err := rq.GetRows(tim)
	assert.NoError(t, err, "Query should work")
	if assert.NotNil(t, res, "Returned rows should not be nil") {
		res.Close()
	}
	checkResults(t, mock)
}
