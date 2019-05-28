package requests_test

import (
	"database/sql/driver"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mlantonn/requests"
	"github.com/stretchr/testify/assert"
)

const (
	// Expected values
	expBool   = true
	expString = "This is a test"
	expInt    = int64(42)
	expFloat  = float64(3.14)
)

var (
	// Expected values
	expByte = []byte(expString)
	expTime = time_now
)

// Test storeToField function, with []byte result type
func TestStoreToFieldBytes(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	var bt []byte
	var btPtr *[]byte
	var s string
	var sPtr *string
	var bo bool
	var boPtr *bool
	rq.Arg = []interface{}{
		&bt, &btPtr, &s, &sPtr, &bo, &boPtr,
	}

	rq.Query = select_query
	mock.ExpectPrepare(select_query).WillBeClosed().
		ExpectQuery().WillReturnRows(setRow(expByte, 6)).RowsWillBeClosed()

	err := rq.GetFields()
	if assert.NoError(t, err) {
		assert.Equal(t, expByte, bt)
		ptrEqualsTo(t, expByte, btPtr)
		assert.Equal(t, expString, s)
		ptrEqualsTo(t, expString, sPtr)
		assert.Equal(t, expBool, bo)
		ptrEqualsTo(t, expBool, boPtr)
	}
	checkResults(t, mock)
}

// Test storeToField function, with int64 result type
func TestStoreToFieldInt(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	var nb int64
	var nbPtr *int64
	rq.Arg = []interface{}{
		&nb, &nbPtr,
	}

	rq.Query = select_query
	mock.ExpectPrepare(select_query).WillBeClosed().
		ExpectQuery().WillReturnRows(setRow(expInt, 2)).RowsWillBeClosed()

	err := rq.GetFields()
	if assert.NoError(t, err) {
		assert.Equal(t, expInt, nb)
		ptrEqualsTo(t, expInt, nbPtr)
	}
	checkResults(t, mock)
}

// Test storeToField function, with float64 result type
func TestStoreToFieldFloat(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	var nb float64
	var nbPtr *float64
	rq.Arg = []interface{}{
		&nb, &nbPtr,
	}

	rq.Query = select_query
	mock.ExpectPrepare(select_query).WillBeClosed().
		ExpectQuery().WillReturnRows(setRow(expFloat, 2)).RowsWillBeClosed()

	err := rq.GetFields()
	if assert.NoError(t, err) {
		assert.Equal(t, expFloat, nb)
		ptrEqualsTo(t, expFloat, nbPtr)
	}
	checkResults(t, mock)
}

// Test storeToField function, with time.Time result type
func TestStoreToFieldTime(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	var tm time.Time
	var tmPtr *time.Time
	rq.Arg = []interface{}{
		&tm, &tmPtr,
	}

	rq.Query = select_query
	mock.ExpectPrepare(select_query).WillBeClosed().
		ExpectQuery().WillReturnRows(setRow(expTime, 2)).RowsWillBeClosed()

	err := rq.GetFields()
	if assert.NoError(t, err) {
		assert.Equal(t, expTime, tm)
		ptrEqualsTo(t, expTime, tmPtr)
	}
	checkResults(t, mock)
}

// Test storeToField function, with nil result type
func TestStoreToFieldNil(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	var bt *[]byte
	var s *string
	var bo *bool
	var i *int64
	var f *float64
	var tm *time.Time
	rq.Arg = []interface{}{
		&bt, &s, &bo, &i, &f, &tm,
	}

	rq.Query = select_query
	mock.ExpectPrepare(select_query).WillBeClosed().
		ExpectQuery().WillReturnRows(setRow(nil, 6)).RowsWillBeClosed()

	err := rq.GetFields()
	if assert.NoError(t, err) {
		assert.Nil(t, bt)
		assert.Nil(t, s)
		assert.Nil(t, bo)
		assert.Nil(t, i)
		assert.Nil(t, f)
		assert.Nil(t, tm)
	}
	checkResults(t, mock)
}

// Test storeToField function, with unsupported storing type, for each type case
func TestStoreToFieldWrongStoringType(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	var res byte
	rq.Arg = []interface{}{
		&res,
	}
	rq.Query = select_query

	testWrongStoringType(t, mock, rq, expByte)
	testWrongStoringType(t, mock, rq, expInt)
	testWrongStoringType(t, mock, rq, expFloat)
	testWrongStoringType(t, mock, rq, expTime)
	testWrongStoringType(t, mock, rq, nil)
}

// Test storeToField function, with unsettable argument, and unsupported result
// type
func TestStoreToFieldWrongResultType(t *testing.T) {
	rq, mock, db := initRequest(t)
	defer db.Close()

	rq.Arg = new([]byte)
	rq.Query = select_query
	mock.ExpectPrepare(select_query).WillBeClosed().
		ExpectQuery().WillReturnRows(setRow(expString, 1)).RowsWillBeClosed()

	err := rq.GetOneField()
	assert.Error(t, err)
	checkResults(t, mock)
}

// Compares ptr (check if nil before checking if equal)
func ptrEqualsTo(t *testing.T, expected, actual interface{}) {
	if assert.NotNil(t, actual) {
		v := reflect.ValueOf(actual).Elem().Interface()
		assert.Equal(t, expected, v)
	}
}

// Create row from value, with given size as number of columns
func setRow(value driver.Value, size int) *sqlmock.Rows {
	values := make([]driver.Value, size)
	columns := make([]string, size)
	for i := range values {
		values[i] = value
		columns[i] = strconv.Itoa(i)
	}
	return sqlmock.NewRows(columns).AddRow(values...)
}

// Tests storeToField with unsupported storing type
func testWrongStoringType(t *testing.T, mock sqlmock.Sqlmock, rq requests.Request, value driver.Value) {
	mock.ExpectPrepare(select_query).WillBeClosed().
		ExpectQuery().WillReturnRows(setRow(value, 1)).RowsWillBeClosed()
	err := rq.GetFields()
	assert.Error(t, err)
	checkResults(t, mock)
}
