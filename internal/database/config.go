package database

import (
	"database/sql"
	"fmt"
)

type DatabaseConfig struct {
	Db_connection *sql.DB
	Queries       *Queries
}

func InitializeDatabase(dbUrl string) DatabaseConfig {
	db, err := sql.Open("postgres", dbUrl)

	if err != nil {
		panic("Error while connecting to database")
	}
	fmt.Printf("Successfully connected to database: %v\n", dbUrl)
	return DatabaseConfig{Db_connection: db, Queries: New(db)}
}
