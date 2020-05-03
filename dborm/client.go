package dborm

import (
	"database/sql"
	"fmt"
)

type DBorm struct {
	DB *sql.DB
}

func NewDBorm() *DBorm {
	return &DBorm{}
}

func (orm *DBorm) Connect(DSN string) error {
	db, err := sql.Open("mysql", DSN)
	if err != nil {
		return fmt.Errorf("Unable to connect to DB", err)
	}
	err = db.Ping() // вот тут будет первое подключение к базе
	if err != nil {
		fmt.Errorf("Unable to ping BD")
	}
	orm.DB = db
	return nil
}
