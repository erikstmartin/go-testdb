package testdb

import (
	"database/sql/driver"
)

type stmt struct {
	result driver.Rows
	err    error
}

func (*stmt) Close() error {
	return nil
}

func (*stmt) NumInput() int {
	return 0
}

func (*stmt) Exec(args []driver.Value) (driver.Result, error) {
	return nil, nil
}

func (s *stmt) Query(args []driver.Value) (driver.Rows, error) {
	return s.result, s.err
}
