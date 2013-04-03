package testdb

type tx struct {
}

func (*tx) Commit() error {
	return nil
}

func (*tx) Rollback() error {
	return nil
}
