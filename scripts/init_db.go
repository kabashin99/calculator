package main

import (
	"database/sql"
	"log"
	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "./database.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS expressions (
            id TEXT PRIMARY KEY,
            status TEXT NOT NULL DEFAULT 'pending',
            result REAL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );

        CREATE TABLE IF NOT EXISTS tasks (
            id TEXT PRIMARY KEY,
            expression_id TEXT NOT NULL,
            arg1 REAL NOT NULL,
            arg2 REAL NOT NULL,
            operation TEXT NOT NULL,
            status TEXT NOT NULL DEFAULT 'pending',
            result REAL,
            FOREIGN KEY (expression_id) REFERENCES expressions(id)
        );

        CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
        CREATE INDEX IF NOT EXISTS idx_expressions_status ON expressions(status);
    `)
	if err != nil {
		log.Fatal("Failed to create tables:", err)
	}

	log.Println("Database initialized successfully")
}
