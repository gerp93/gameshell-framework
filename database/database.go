package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func GetDatabaseConnectionString() string {
	serverHost := os.Getenv("GFB_SQL_HOST")
	userName := os.Getenv("GFB_SQL_USER")
	userPassword := os.Getenv("GFB_SQL_PASSWORD")
	databaseName := os.Getenv("GFB_SQL_DATABASE")
	return fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true", userName, userPassword, serverHost, databaseName)
}

func Ping(dbcs string) error {
	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return err
	}
	return nil
}
