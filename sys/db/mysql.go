package db

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"nemo/sys/log"
)

type ConnPool struct {
	db             *sql.DB
	driverName     string
	dataSourceName string
}

// Search multiple rows
func (p *ConnPool) Query(query string, args ...interface{}) (rows *sql.Rows) {
	rows, err := p.db.Query(query, args...)
	if err != nil {
		p.fail("QUERY", query, err, args...)
		return nil
	}
	return rows
}

// Search only one row.
func (p *ConnPool) QueryRow(query string, args ...interface{}) (row *sql.Row) {
	return p.db.QueryRow(query, args...)
}

// Execute with no rows return, but return result.
func (p *ConnPool) Exec(query string, args ...interface{}) (res sql.Result, err error) {
	res, err = p.db.Exec(query, args...)
	if err != nil {
		p.fail("EXEC", query, err, args...)
		return
	}
	count, err := res.RowsAffected()
	if err != nil {
		p.fail("EXEC", query, errors.New(fmt.Sprintf("Affected rows:%s", err.Error())), args...)
		return
	}
	if count != 1 {
		p.fail("EXEC", query, errors.New(fmt.Sprintf("Affected rows:%d", count)), args...)
		return
	}
	return
}

// Open connection pool.
func (p *ConnPool) open() error {
	var err error
	if p.db, err = sql.Open(p.driverName, p.dataSourceName); err != nil {
		return err
	}
	if err = p.db.Ping(); err != nil {
		return err
	}
	return nil
}

// Close pool.
func (p *ConnPool) Close() {
	p.db.Close()
}

func (p *ConnPool) fail(method, query string, err error, args ...interface{}) {
	log.Errorf("Failed to execute SQL [%s][%s] [Parameter]:%v [Error]:%s\n", method, query, args, err.Error())
}

// Create mysql connection pool.
func InitMySQLPool(DSN string) *ConnPool {
	dataSourceName := fmt.Sprintf("%s", DSN)
	db := &ConnPool{
		driverName:     "mysql",
		dataSourceName: dataSourceName,
	}
	if err := db.open(); err != nil {
		log.Fatalf("Failed to initialize database, Error:%v", err)
	}
	return db
}
