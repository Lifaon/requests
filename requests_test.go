package requests_test

import (
	"database/sql"
	"testing"

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

func TestPrepareStatement(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	mock.ExpectPrepare("SELECT")

	rq.Query = "SELECT id FROM user"
	stmt, _ := rq.PrepareStmt()
	if stmt != nil {
		stmt.Close()
	}

	checkResults(t, mock)
}

func TestExecQuery(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	rq.Query = "INSERT INTO user (id, name) VALUES (?, ?)"
	// mock.ExpectPrepare(rq.Query).WillBeClosed()
	// mock.ExpectExec(rq.Query).WithArgs(a, b).WillReturnResult(result_1)

	_, err := rq.ExecQuery(a, b)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	checkResults(t, mock)
}

func TestExecQueryWrongQuery(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	rq.Query = "INSERT INTO user (id) VALUES (?, ?)"

	mock.ExpectPrepare(rq.Query).WillReturnError(nil)

	rq.ExecQuery(a, b)

	checkResults(t, mock)
}

func TestExecQueryWrongParams(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	mock.ExpectPrepare("INSERT").WillBeClosed()
	mock.ExpectExec("INSERT").WithArgs(a).WillReturnError(nil)

	rq.Query = "INSERT INTO user (id, name) VALUES (?, ?)"
	rq.ExecQuery(a)

	checkResults(t, mock)
}
