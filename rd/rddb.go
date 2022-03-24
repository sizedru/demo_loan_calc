package rd

import (
	"database/sql"
	"log"
)

// InitDB инициализация БД
func InitDB(user, password, database, host string) (db *sql.DB) {
	var err error
	var connStr, connDb, connHost string
	if user == "" {
		connStr = "..............."
		connDb = "................."
		connHost = "............"
	} else {
		connStr = user + ":" + password
		connDb = database
		connHost = host
	}
	db, err = sql.Open("mysql", connStr+"@tcp("+connHost+":3306)/"+connDb+"?charset=utf8")
	if err != nil {
		log.Panic(err)
	}
	if err = db.Ping(); err != nil {
		log.Panic(err)
	}
	return
}

// CloseDB закрытие соединения с БД
func CloseDB(db *sql.DB) {
	db.Close()
}

// QueryDB return query result rows
func QueryDB(db *sql.DB, query string) (*sql.Rows, error) {
	//trx, _ := db.Begin()
	//defer trx.Rollback()
	//rows, err := trx.Query(query)
	rows, err := db.Query(query)
	if err != nil {
		rows = nil
	}
	return rows, err
}
