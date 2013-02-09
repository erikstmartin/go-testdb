package testdb

import (
	"database/sql/driver"
	"errors"
)

type conn struct {
	queries   map[string]query
	queryFunc func(query string) (result driver.Rows, err error)
}

func newConn() *conn {
	return &conn{
		queries: make(map[string]query),
	}
}

func (c *conn) Prepare(query string) (driver.Stmt, error) {
	if c.queryFunc != nil {
		result, err := c.queryFunc(query)

		return &stmt{
			result: result,
			err:    err,
		}, nil
	}

	if q, ok := d.conn.queries[getQueryHash(query)]; ok {
		return &stmt{
			result: q.result,
			err:    q.err,
		}, nil
	}

	return new(stmt), errors.New("Query not stubbed: " + query)
}

func (*conn) Close() error {
	return nil
}

func (*conn) Begin() (driver.Tx, error) {
	return &tx{}, nil
}

func (c *conn) Exec(query string, args []driver.Value) (driver.Result, error) {
	return nil, nil
}
