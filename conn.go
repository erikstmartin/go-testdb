package testdb

import (
	"database/sql/driver"
	"errors"
)

type conn struct {
	queries      map[string]query
	queryFunc    func(query string, args []driver.Value) (driver.Rows, error)
	execFunc     func(query string, args []driver.Value) (driver.Result, error)
	beginFunc    func() (driver.Tx, error)
	commitFunc   func() error
	rollbackFunc func() error
}

func newConn() *conn {
	return &conn{
		queries: make(map[string]query),
	}
}

func (c *conn) Prepare(query string) (driver.Stmt, error) {
	s := new(stmt)

	if c.queryFunc != nil {
		s.queryFunc = func(args []driver.Value) (driver.Rows, error) {
			return c.queryFunc(query, args)
		}
	}

	if c.execFunc != nil {
		s.execFunc = func(args []driver.Value) (driver.Result, error) {
			return c.execFunc(query, args)
		}
	}

	if q, ok := c.queries[getQueryHash(query)]; ok {
		if s.queryFunc == nil && q.rows != nil {
			s.queryFunc = func(args []driver.Value) (driver.Rows, error) {
				if q.rows != nil {
					if rows, ok := q.rows.(*rows); ok {
						return rows.clone(), nil
					}
					return q.rows, nil
				}
				return nil, q.err
			}
		}

		if s.execFunc == nil && q.result != nil {
			s.execFunc = func(args []driver.Value) (driver.Result, error) {
				if q.result != nil {
					return q.result, nil
				}
				return nil, q.err
			}
		}
	}

	if s.queryFunc == nil && s.execFunc == nil {
		return new(stmt), errors.New("Query not stubbed: " + query)
	}

	return s, nil
}

func (*conn) Close() error {
	return nil
}

func (c *conn) Begin() (driver.Tx, error) {
	if c.beginFunc != nil {
		return c.beginFunc()
	}

	t := &Tx{}
	if c.commitFunc != nil {
		t.SetCommitFunc(c.commitFunc)
	}
	if c.rollbackFunc != nil {
		t.SetRollbackFunc(c.rollbackFunc)
	}

	return t, nil
}

func (c *conn) Query(query string, args []driver.Value) (driver.Rows, error) {
	if c.queryFunc != nil {
		return c.queryFunc(query, args)
	}
	if q, ok := c.queries[getQueryHash(query)]; ok {
		if rows, ok := q.rows.(*rows); ok {
			return rows.clone(), q.err
		}
		return q.rows, q.err
	}
	return nil, errors.New("Query not stubbed: " + query)
}

func (c *conn) Exec(query string, args []driver.Value) (driver.Result, error) {
	if c.execFunc != nil {
		return c.execFunc(query, args)
	}

	if q, ok := c.queries[getQueryHash(query)]; ok {
		if q.result != nil {
			return q.result, nil
		} else if q.err != nil {
			return nil, q.err
		}
	}

	return nil, errors.New("Exec call not stubbed: " + query)
}

// Stubbing functions
// Set your own function to be executed when db.Query() is called. As with StubQuery() you can use the RowsFromCSVString() method to easily generate the driver.Rows, or you can return your own.
func (c *conn) SetQueryFunc(f func(query string) (result driver.Rows, err error)) {
	c.SetQueryWithArgsFunc(func(query string, args []driver.Value) (result driver.Rows, err error) {
		return f(query)
	})
}

// Set your own function to be executed when db.Query() is called. As with StubQuery() you can use the RowsFromCSVString() method to easily generate the driver.Rows, or you can return your own.
func (c *conn) SetQueryWithArgsFunc(f func(query string, args []driver.Value) (result driver.Rows, err error)) {
	c.queryFunc = f
}

// Stubs the global driver.Conn to return the supplied driver.Rows when db.Query() is called, query stubbing is case insensitive, and whitespace is also ignored.
func (c *conn) StubQuery(q string, rows driver.Rows) {
	c.queries[getQueryHash(q)] = query{
		rows: rows,
	}
}

// Stubs the global driver.Conn to return the supplied error when db.Query() is called, query stubbing is case insensitive, and whitespace is also ignored.
func (c *conn) StubQueryError(q string, err error) {
	c.queries[getQueryHash(q)] = query{
		err: err,
	}
}

// Set your own function to be executed when db.Exec is called. You can return an error or a Result object with the LastInsertId and RowsAffected
func (c *conn) SetExecFunc(f func(query string) (driver.Result, error)) {
	c.SetExecWithArgsFunc(func(query string, args []driver.Value) (driver.Result, error) {
		return f(query)
	})
}

// Set your own function to be executed when db.Exec is called. You can return an error or a Result object with the LastInsertId and RowsAffected
func (c *conn) SetExecWithArgsFunc(f func(query string, args []driver.Value) (driver.Result, error)) {
	c.execFunc = f
}

// Stubs the global driver.Conn to return the supplied Result when db.Exec is called, query stubbing is case insensitive, and whitespace is also ignored.
func (c *conn) StubExec(q string, r *Result) {
	c.queries[getQueryHash(q)] = query{
		result: r,
	}
}

// Stubs the global driver.Conn to return the supplied error when db.Exec() is called, query stubbing is case insensitive, and whitespace is also ignored.
func (c *conn) StubExecError(q string, err error) {
	c.StubQueryError(q, err)
}

// Set your own function to be executed when db.Begin() is called. You can either hand back a valid transaction, or an error.
func (c *conn) SetBeginFunc(f func() (driver.Tx, error)) {
	c.beginFunc = f
}

// Stubs the global driver.Conn to return the supplied tx and error when db.Begin() is called.
func (c *conn) StubBegin(tx driver.Tx, err error) {
	c.SetBeginFunc(func() (driver.Tx, error) {
		return tx, err
	})
}

// Set your own function to be executed when tx.Commit() is called on the default transcation.
func (c *conn) SetCommitFunc(f func() error) {
	c.commitFunc = f
}

// Stubs the default transaction to return the supplied error when tx.Commit() is called.
func (c *conn) StubCommitError(err error) {
	c.SetCommitFunc(func() error {
		return err
	})
}

// Set your own function to be executed when tx.Rollback() is called on the default transcation.
func (c *conn) SetRollbackFunc(f func() error) {
	c.rollbackFunc = f
}

// Stubs the default transaction to return the supplied error when tx.Rollback() is called.
func (c *conn) StubRollbackError(err error) {
	c.SetRollbackFunc(func() error {
		return err
	})
}
