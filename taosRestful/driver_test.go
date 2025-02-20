package taosRestful

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Ensure that all the driver interfaces are implemented

var (
	driverName                    = "taosRestful"
	user                          = "root"
	password                      = "taosdata"
	host                          = "127.0.0.1"
	port                          = 6041
	dbName                        = "test_taos_restful"
	dataSourceName                = fmt.Sprintf("%s:%s@http(%s:%d)/", user, password, host, port)
	dataSourceNameWithCompression = fmt.Sprintf("%s:%s@http(%s:%d)/?disableCompression=false", user, password, host, port)
)

type DBTest struct {
	*testing.T
	*sql.DB
}

func NewDBTest(t *testing.T) (dbt *DBTest) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		t.Fatalf("error on:  sql.open %s", err.Error())
		return
	}
	dbt = &DBTest{t, db}
	return
}

func (dbt *DBTest) CreateTables(numOfSubTab int) {
	dbt.mustExec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
	dbt.mustExec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbName))
	dbt.mustExec(fmt.Sprintf("drop table if exists %s.super", dbName))
	dbt.mustExec(fmt.Sprintf("CREATE TABLE %s.super (ts timestamp, value BOOL) tags (degress int)", dbName))
	for i := 0; i < numOfSubTab; i++ {
		dbt.mustExec(fmt.Sprintf("create table %s.t%d using %s.super tags(%d)", dbName, i%10, dbName, i))
	}
}
func (dbt *DBTest) InsertInto(numOfSubTab, numOfItems int) {
	now := time.Now()
	t := now.Add(-100 * time.Minute)
	for i := 0; i < numOfItems; i++ {
		dbt.mustExec(fmt.Sprintf("insert into %s.t%d values(%d, %t)", dbName, i%numOfSubTab, t.UnixNano()/int64(time.Millisecond)+int64(i), i%2 == 0))
	}
}

type TestResult struct {
	ts      string
	value   bool
	degress int
}

func runTests(t *testing.T, tests ...func(dbt *DBTest)) {
	dbt := NewDBTest(t)
	// prepare data
	dbt.Exec("DROP TABLE IF EXISTS test_taos_restful.test")
	var numOfSubTables = 10
	var numOfItems = 200
	dbt.CreateTables(numOfSubTables)
	dbt.InsertInto(numOfSubTables, numOfItems)
	for _, test := range tests {
		test(dbt)
		dbt.Exec("DROP TABLE IF EXISTS test_taos_restful.test")
	}
}
// func (dbt *DBTest) fail(method, query string, err error) {
// 	if len(query) > 300 {
// 		query = "[query too large to print]"
// 	}
// 	dbt.Fatalf("error on %s %s: %s", method, query, err.Error())
// }

func (dbt *DBTest) mustExec(query string, args ...interface{}) (res sql.Result, err error) {
	res, err = dbt.Exec(query, args...)
	return
}

func (dbt *DBTest) mustQuery(query string, args ...interface{}) (rows *sql.Rows, err error) {
	rows, err = dbt.Query(query, args...)
	return
}

// @author: xftan
// @date: 2022/2/8 12:52
// @description: test empty sql
func TestEmptyQuery(t *testing.T) {
	runTests(t, func(dbt *DBTest) {
		// just a comment, no query
		_, err := dbt.mustExec("")
		if err == nil {
			dbt.Fatalf("error is expected")
		}

	})
}

// @author: xftan
// @date: 2022/2/8 12:52
// @description: test error sql
func TestErrorQuery(t *testing.T) {
	runTests(t, func(dbt *DBTest) {
		// just a comment, no query
		_, err := dbt.mustExec("xxxxxxx inot")
		if err == nil {
			dbt.Fatalf("error is expected")
		}
	})
}

type (
	execFunc func(dbt *DBTest, query string, exec bool, err error, expected int64) int64
)

type Obj struct {
	query  string
	err    error
	exec   bool
	fp     execFunc
	expect int64
}

var (
	errUser = errors.New("user error")
	fp      = func(dbt *DBTest, query string, exec bool, eErr error, expected int64) int64 {
		var ret int64 = 0
		if exec == false {
			rows, err := dbt.mustQuery(query)
			if eErr == errUser && err != nil {
				return ret
			}
			if err != nil {
				dbt.Errorf("%s is not expected, err: %s", query, err.Error())
				return ret
			} else {
				var count int64 = 0
				for rows.Next() {
					var row TestResult
					if err := rows.Scan(&(row.ts), &(row.value)); err != nil {
						dbt.Error(err.Error())
						return ret
					}
					count = count + 1
				}
				rows.Close()
				ret = count
				if expected != -1 && count != expected {
					dbt.Errorf("%s is not expected, err: %s", query, errors.New("result is not expected"))
				}
			}
		} else {
			res, err := dbt.mustExec(query)
			if err != eErr {
				dbt.Fatalf("%s is not expected, err: %s", query, err.Error())
			} else {
				count, err := res.RowsAffected()
				if err != nil {
					dbt.Fatalf("%s is not expected, err: %s", query, err.Error())
				}
				if expected != -1 && count != expected {
					dbt.Fatalf("%s is not expected , err: %s", query, errors.New("result is not expected"))
				}
			}
		}
		return ret
	}
)

// @author: xftan
// @date: 2022/2/8 12:58
// @description: test random write and query
func TestAny(t *testing.T) {
	runTests(t, func(dbt *DBTest) {
		now := time.Now()
		tests := make([]*Obj, 0, 100)
		tests = append(tests,
			&Obj{fmt.Sprintf("insert into %s.t%d values(%d, %t)", dbName, 0, now.UnixNano()/int64(time.Millisecond)-1, false), nil, true, fp, int64(1)})
		tests = append(tests,
			&Obj{fmt.Sprintf("insert into %s.t%d values(%d, %t)", dbName, 0, now.UnixNano()/int64(time.Millisecond)-1, false), nil, true, fp, int64(1)})
		tests = append(tests,
			&Obj{fmt.Sprintf("select last_row(*) from %s.t%d", dbName, 0), nil, false, fp, int64(1)})
		tests = append(tests,
			&Obj{fmt.Sprintf("select first(*) from %s.t%d", dbName, 0), nil, false, fp, int64(1)})
		tests = append(tests,
			&Obj{"select error", errUser, false, fp, int64(1)})
		tests = append(tests,
			&Obj{fmt.Sprintf("select * from %s.t%d", dbName, 0), nil, false, fp, int64(-1)})
		tests = append(tests,
			&Obj{fmt.Sprintf("select * from %s.t%d", dbName, 0), nil, false, fp, int64(-1)})

		for _, obj := range tests {
			fp = obj.fp
			fp(dbt, obj.query, obj.exec, obj.err, obj.expect)
		}
	})
}

// @author: xftan
// @date: 2022/2/8 12:59
// @description: test query and insert
func TestCRUD(t *testing.T) {
	runTests(t, func(dbt *DBTest) {
		// Create Data
		now := time.Now()
		res, err := dbt.mustExec(fmt.Sprintf("insert into %s.t%d values(%d, %t)", dbName, 0, now.UnixNano()/int64(time.Millisecond)-1, false))
		if err != nil {
			dbt.Fatalf("insert failed %s", err.Error())
		}
		count, err := res.RowsAffected()
		if err != nil {
			dbt.Fatalf("res.RowsAffected() returned error: %s", err.Error())
		}
		if count != 1 {
			dbt.Fatalf("expected 1 affected row, got %d", count)
		}

		id, err := res.LastInsertId()
		if err == nil {
			dbt.Fatalf("res.LastInsertId() expect error")
		}
		if id != 0 {
			dbt.Fatalf("expected InsertId 0, got %d", id)
		}

		// Read
		rows, err := dbt.mustQuery(fmt.Sprintf("select * from %s.super", dbName))
		if err != nil {
			dbt.Fatalf("select failed")
		}
		for rows.Next() {
			var row TestResult
			err := rows.Scan(&(row.ts), &(row.value), &(row.degress))
			if err != nil {
				dbt.Error(err.Error())
			}
		}
		rows.Close()

		rows, err = dbt.mustQuery(fmt.Sprintf("select last_row(*) from %s.super", dbName))
		if err != nil {
			dbt.Fatalf("select last_row failed")
		} else {
			for rows.Next() {
				var value TestResult
				err := rows.Scan(&(value.ts), &(value.value))
				if err != nil {
					dbt.Error(err.Error())
				}
			}
			rows.Close()
		}

		query2 := "drop table if exists super"
		dbt.mustExec(query2)
		if err != nil {
			dbt.Fatalf(query2)
		}
	})
}

// @author: xftan
// @date: 2022/1/27 16:18
// @description:  test chinese insert and query
func TestChinese(t *testing.T) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		t.Fatalf("error on:  sql.open %s", err.Error())
		return
	}
	defer db.Close()
	_, err = db.Exec("create database if not exists test_chinese")
	if err != nil {
		t.Error(err)
		return
	}
	_, err = db.Exec("drop table if exists test_chinese.chinese")
	if err != nil {
		t.Error(err)
		return
	}
	_, err = db.Exec("create table if not exists test_chinese.chinese(ts timestamp,value nchar(32))")
	if err != nil {
		t.Error(err)
		return
	}
	_, err = db.Exec(`INSERT INTO test_chinese.chinese (ts, value) VALUES (?, ?)`, "1641010332000", "'阴天'")
	if err != nil {
		t.Error(err)
		return
	}
	rows, err := db.Query("select * from test_chinese.chinese")
	if err != nil {
		t.Error(err)
		return
	}
	counter := 0
	for rows.Next() {
		counter += 1
		row := make([]driver.Value, 2)
		err := rows.Scan(&row[0], &row[1])
		if err != nil {
			t.Error(err)
			return
		}
		t.Log(row)
	}
	assert.Equal(t, 1, counter)
}
