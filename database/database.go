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

func Ping() (err error) {
	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		log.Println(err)
		return errors.New("failed to open database connection")
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Println(err)
		return errors.New("failed to ping database")
	}
	return nil
}

func Query(sqlString string, params ...interface{}) (rows *sql.Rows, err error) {
	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		log.Println(err)
		return nil, errors.New("failed to open database connection")
	}
	defer db.Close()

	statement, err := db.Prepare(sqlString)
	if err != nil {
		log.Println(err)
		return nil, errors.New("failed to prepare database statement")
	}
	defer statement.Close()

	rows, err = statement.Query(params)
	if err != nil {
		log.Println(err)
		return nil, errors.New("failed to query statement in database")
	}

	return rows, nil
}

func Execute(sqlString string, params ...interface{}) (err error) {
	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		log.Println(err)
		return errors.New("failed to open database connection")
	}
	defer db.Close()

	statement, err := db.Prepare(sqlString)
	if err != nil {
		log.Println(err)
		return errors.New("failed to prepare database statement")
	}
	defer statement.Close()

	_, err = statement.Exec(params)
	if err != nil {
		log.Println(err)
		return errors.New("failed to execute statement in database")
	}

	return nil
}
