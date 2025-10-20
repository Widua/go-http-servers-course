package database

import (
	"database/sql"
	"fmt"
)

type DatabaseConfig struct {
	Db_connection *sql.DB
	Queries       *Queries
}

var DB_Config DatabaseConfig

func InitializeDatabase(dbUrl string) {
	db, err := sql.Open("postgres", dbUrl)

	if err != nil {
		panic("Error while connecting to database")
	}
	DB_Config = DatabaseConfig{Db_connection: db, Queries: New(db)}
	fmt.Printf("Successfully connected to database: %v\n", dbUrl)
}
