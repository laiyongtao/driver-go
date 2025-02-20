package taosSql

import (
	"context"
	"database/sql/driver"
	"fmt"
	"reflect"

	"github.com/taosdata/driver-go/v2/errors"
)

type taosSqlStmt struct {
	tc         *taosConn
	id         uint32
	pSql       string
	paramCount int
}

func (stmt *taosSqlStmt) Close() error {
	return nil
}

func (stmt *taosSqlStmt) NumInput() int {
	return stmt.paramCount
}

func (stmt *taosSqlStmt) Exec(args []driver.Value) (driver.Result, error) {
	if stmt.tc == nil || stmt.tc.taos == nil {
		return nil, errors.ErrTscInvalidConnection
	}
	return stmt.tc.Exec(stmt.pSql, args)
}

func (stmt *taosSqlStmt) Query(args []driver.Value) (driver.Rows, error) {
	if stmt.tc == nil || stmt.tc.taos == nil {
		return nil, errors.ErrTscInvalidConnection
	}
	return stmt.tc.Query(stmt.pSql, args)
}

func (stmt *taosSqlStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	if stmt.tc == nil {
		return nil, errors.ErrTscInvalidConnection
	}
	driverArgs, err := namedValueToValue(args)

	if err != nil {
		return nil, err
	}

	rs, err := stmt.Query(driverArgs)
	if err != nil {
		return nil, err
	}
	return rs, err
}

func (stmt *taosSqlStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	if stmt.tc == nil {
		return nil, errors.ErrTscInvalidConnection
	}

	driverArgs, err := namedValueToValue(args)
	if err != nil {
		return nil, err
	}

	return stmt.Exec(driverArgs)
}

type converter struct{}

// ConvertValue mirrors the reference/default converter in database/sql/driver
// with _one_ exception.  We support uint64 with their high bit and the default
// implementation does not.  This function should be kept in sync with
// database/sql/driver defaultConverter.ConvertValue() except for that
// deliberate difference.
func (c converter) ConvertValue(v interface{}) (driver.Value, error) {

	if driver.IsValue(v) {
		return v, nil
	}

	if vr, ok := v.(driver.Valuer); ok {
		sv, err := callValuerValue(vr)
		if err != nil {
			return nil, err
		}
		if !driver.IsValue(sv) {
			return nil, fmt.Errorf("non-Value type %T returned from Value", sv)
		}

		return sv, nil
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr:
		// indirect pointers
		if rv.IsNil() {
			return nil, nil
		} else {
			return c.ConvertValue(rv.Elem().Interface())
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint(), nil
	case reflect.Float32, reflect.Float64:
		return rv.Float(), nil
	case reflect.Bool:
		return rv.Bool(), nil
	case reflect.Slice:
		ek := rv.Type().Elem().Kind()
		if ek == reflect.Uint8 {
			return rv.Bytes(), nil
		}
		return nil, fmt.Errorf("unsupported type %T, a slice of %s", v, ek)
	case reflect.String:
		return rv.String(), nil
	}
	return nil, fmt.Errorf("unsupported type %T, a %s", v, rv.Kind())
}

var valuerReflectType = reflect.TypeOf((*driver.Valuer)(nil)).Elem()

// callValuerValue returns vr.Value(), with one exception:
// If vr.Value is an auto-generated method on a pointer type and the
// pointer is nil, it would panic at runtime in the panicwrap
// method. Treat it like nil instead.
//
// This is so people can implement driver.Value on value types and
// still use nil pointers to those types to mean nil/NULL, just like
// string/*string.
//
// This is an exact copy of the same-named unexported function from the
// database/sql package.
func callValuerValue(vr driver.Valuer) (v driver.Value, err error) {
	if rv := reflect.ValueOf(vr); rv.Kind() == reflect.Ptr &&
		rv.IsNil() &&
		rv.Type().Elem().Implements(valuerReflectType) {
		return nil, nil
	}
	return vr.Value()
}
