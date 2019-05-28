package requests_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mlantonn/requests"
	"github.com/stretchr/testify/assert"
)

type (
	InvalidStruct struct {
		ID   int64
		Name string
	}
)

func TestScanIntoStructsInvalidArgs(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	rq.Query = select_query

	testScanIntoStructs(t, mock, rq, "")
	testScanIntoStructs(t, mock, rq, new(string))
	testScanIntoStructs(t, mock, rq, &[]string{})
	testScanIntoStructs(t, mock, rq, &[]EmptyStruct{})
	testScanIntoStructs(t, mock, rq, &[]InvalidStruct{})
}

func testScanIntoStructs(t *testing.T, mock sqlmock.Sqlmock, rq requests.Request, arg interface{}) {
	rq.Arg = arg
	mock.ExpectPrepare(select_query).WillBeClosed().
		ExpectQuery().WillReturnRows(getSelectRows()).RowsWillBeClosed()
	err := rq.GetIntoStructs()
	assert.Error(t, err)
	checkResults(t, mock)
}

func TestScanIntoOneStructInvalidArgs(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	rq.Query = select_query

	testScanIntoOneStruct(t, mock, rq, false, "")
	testScanIntoOneStruct(t, mock, rq, false, new(string))
	testScanIntoOneStruct(t, mock, rq, true, &EmptyStruct{})
	testScanIntoOneStruct(t, mock, rq, true, &InvalidStruct{})
}

func testScanIntoOneStruct(t *testing.T, mock sqlmock.Sqlmock, rq requests.Request, scan bool, arg interface{}) {
	rq.Arg = arg
	if scan {
		mock.ExpectPrepare(select_query).WillBeClosed().
			ExpectQuery().WillReturnRows(getSelectRows()).RowsWillBeClosed()
	} else {
		mock.ExpectPrepare(select_query).WillBeClosed().
			ExpectQuery()
	}
	err := rq.GetIntoOneStruct()
	assert.Error(t, err)
	checkResults(t, mock)
}

func TestScanIntoOnePtrInvalidArgs(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	rq.Query = select_query

	testScanIntoOnePtr(t, mock, rq, false, "")
	testScanIntoOnePtr(t, mock, rq, true, new(int64))
}

func testScanIntoOnePtr(t *testing.T, mock sqlmock.Sqlmock, rq requests.Request, scanError bool, arg interface{}) {
	rq.Arg = arg
	if scanError {
		mock.ExpectPrepare(select_query).WillBeClosed().
			ExpectQuery().WillReturnRows(getStructRows()).RowsWillBeClosed()
	} else {
		mock.ExpectPrepare(select_query).WillBeClosed().
			ExpectQuery().WillReturnRows(getSingleFieldRows()).RowsWillBeClosed()
	}
	err := rq.GetOneField()
	assert.Error(t, err)
	checkResults(t, mock)
}

func TestScanIntoPtrsInvalidArgs(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	rq.Query = select_query

	testScanIntoPtrs(t, mock, rq, false, "")
	testScanIntoPtrs(t, mock, rq, true, []*int64{})
}

func testScanIntoPtrs(t *testing.T, mock sqlmock.Sqlmock, rq requests.Request, scan bool, arg interface{}) {
	rq.Arg = arg
	if scan {
		mock.ExpectPrepare(select_query).WillBeClosed().
			ExpectQuery().WillReturnRows(getStructRows()).RowsWillBeClosed()
	} else {
		mock.ExpectPrepare(select_query).WillBeClosed().
			ExpectQuery()
	}
	err := rq.GetFields()
	assert.Error(t, err)
	checkResults(t, mock)
}

func TestScanIntoSliceInvalidArgs(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	rq.Query = select_query

	testScanIntoSlice(t, mock, rq, false, "")
	testScanIntoSlice(t, mock, rq, false, new(string))
	testScanIntoSlice(t, mock, rq, true, &[]int64{})
	testScanIntoSlice(t, mock, rq, false, &[]string{})
}

func testScanIntoSlice(t *testing.T, mock sqlmock.Sqlmock, rq requests.Request, scanError bool, arg interface{}) {
	rq.Arg = arg
	if scanError {
		mock.ExpectPrepare(select_query).WillBeClosed().
			ExpectQuery().WillReturnRows(getStructRows()).RowsWillBeClosed()
	} else {
		mock.ExpectPrepare(select_query).WillBeClosed().
			ExpectQuery().WillReturnRows(getSingleFieldRows()).RowsWillBeClosed()
	}
	err := rq.GetIntoSlice()
	assert.Error(t, err)
	checkResults(t, mock)
}
