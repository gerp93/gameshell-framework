package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

var dbcs string
var allUsers map[uuid.UUID]User = make(map[uuid.UUID]User)

func Setup() (err error) {
	// get connection string
	userName := os.Getenv("CARD_JUDGE_SQL_USER")
	userPassword := os.Getenv("CARD_JUDGE_SQL_PASSWORD")
	serverHost := os.Getenv("CARD_JUDGE_SQL_HOST")
	databaseName := os.Getenv("CARD_JUDGE_SQL_DATABASE")
	dbcs = fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true", userName, userPassword, serverHost, databaseName)

	// ping to test connection
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

	// load all users into memory
	users, err := getUsers()
	if err != nil {
		return err
	}

	for _, user := range users {
		allUsers[user.Id] = user
	}

	return nil
}

func Query(sqlString string, params ...any) (rows *sql.Rows, err error) {
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

	rows, err = statement.Query(params...)
	if err != nil {
		log.Println(err)
		return nil, errors.New("failed to query statement in database")
	}

	return rows, nil
}

func Execute(sqlString string, params ...any) (err error) {
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

	_, err = statement.Exec(params...)
	if err != nil {
		log.Println(err)
		return errors.New("failed to execute statement in database")
	}

	return nil
}
