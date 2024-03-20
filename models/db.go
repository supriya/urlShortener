package models

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	dbHost = "127.0.0.1"
	dbUser = "root"
	dbPort = "3306"
	dbName = "url_data"
)

func dsn(dbName string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, os.Getenv("MYSQL_ROOT_PASSWORD"), dbHost, dbPort, dbName)
}

func InitDB() (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn(""))
	if err != nil {
		log.Printf("Error opening database\n%s", err)
		return nil, err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	dbCreateQuery := "CREATE DATABASE IF NOT EXISTS " + dbName
	res, err := db.ExecContext(ctx, dbCreateQuery)
	if err != nil {
		log.Printf("Error creating the database\n%s", err)
		return nil, err
	}
	if _, err := res.RowsAffected(); err != nil {
		log.Printf("Error fetching the database creation response\n %s", err)
		return nil, err
	}

	db, err = sql.Open("mysql", dsn(dbName))
	if err != nil {
		log.Printf("Error opening database\n %s", err)
		return nil, err
	}
	// defer db.Close()

	if err := db.Ping(); err != nil {
		return nil, err
	}
	log.Print("Successfully connected to the database.")

	if err := createTableAndIndex(db, ctx); err != nil {
		return nil, err
	}
	return db, nil

}

func createTableAndIndex(db *sql.DB, ctx context.Context) error {
	tableCreateQuery := `CREATE TABLE IF NOT EXISTS urls 
						(id int NOT NULL AUTO_INCREMENT, long_url varchar(1024), short_url varchar(256), expiry_time datetime, created_at datetime, updated_at datetime, UNIQUE (short_url), PRIMARY KEY (id))`

	res, err := db.ExecContext(ctx, tableCreateQuery)
	if err != nil {
		log.Printf("Error creating the table\n%s", err)
		return err
	}
	if _, err := res.RowsAffected(); err != nil {
		log.Printf("Error fetching the table creation response\n%s", err)
		return err
	}

	return nil
}
