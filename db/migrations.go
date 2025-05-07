package db

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
)

func RunMigrations(db *sql.DB) error {
	fmt.Println("Running SQLite DB migrations...")

	statements := []string{
		`CREATE TABLE IF NOT EXISTS users (
            login TEXT PRIMARY KEY,
            password TEXT NOT NULL,
            created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			last_login DATETIME
        );`,
		`CREATE TABLE IF NOT EXISTS expressions (
            id TEXT PRIMARY KEY,
			status TEXT NOT NULL,
			result REAL,
			owner TEXT NOT NULL,
			FOREIGN KEY (owner) REFERENCES users(login)
        );`,
		`CREATE TABLE IF NOT EXISTS tasks (
            id TEXT PRIMARY KEY,
			arg1 REAL NOT NULL,
			arg2 REAL NOT NULL,
			operation TEXT NOT NULL,
			operation_time INTEGER,
			result REAL,
			depends_on TEXT,
			user_login TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_login) REFERENCES users(login)
        );`,
	}

	for _, stmt := range statements {
		tableName := extractTableName(stmt)
		fmt.Printf("Migrating table: %s\n", tableName)

		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("migration failed for table %s: %w", tableName, err)
		}
	}

	fmt.Println("DB migrations completed")
	return nil
}

func extractTableName(stmt string) string {
	re := regexp.MustCompile(`(?i)CREATE\s+TABLE\s+IF\s+NOT\s+EXISTS\s+(\w+)`)
	matches := re.FindStringSubmatch(stmt)
	if len(matches) >= 2 {
		return strings.TrimSpace(matches[1])
	}
	return "unknown"
}
