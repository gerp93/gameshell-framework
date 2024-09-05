package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var dbcs string

func SetDatabaseConnectionString() {
	serverHost := os.Getenv("GFB_SQL_HOST")
	userName := os.Getenv("GFB_SQL_USER")
	userPassword := os.Getenv("GFB_SQL_PASSWORD")
	databaseName := os.Getenv("GFB_SQL_DATABASE")
	dbcs = fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true", userName, userPassword, serverHost, databaseName)
}

func Ping() error {
	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		log.Println(err)
		return errors.New("failed to connect to database")
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Println(err)
		return errors.New("failed to ping database")
	}
	return nil
}
