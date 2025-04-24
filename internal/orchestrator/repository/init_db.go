package repository

import (
	"database/sql"
	"log"
)

func InitDB(db *sql.DB) {
	// Создание таблицы expressions
	createExpressionsTableQuery := `
		CREATE TABLE IF NOT EXISTS expressions (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id TEXT NOT NULL,
			status TEXT DEFAULT 'pending',
			result DOUBLE PRECISION DEFAULT 0
		);
	`
	_, err := db.Exec(createExpressionsTableQuery)
	if err != nil {
		log.Fatalf("Ошибка при создании таблицы expressions: %v", err)
	}
	log.Printf("Таблица expressions успешно создана или уже существует.")

	// Создание таблицы tasks
	createTasksTableQuery := `
		CREATE TABLE IF NOT EXISTS tasks (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			expression_id UUID NOT NULL,
			operation TEXT NOT NULL,
			operand1 DOUBLE PRECISION NOT NULL,
			operand2 DOUBLE PRECISION NOT NULL,
			operation_time INTEGER NOT NULL,
			result DOUBLE PRECISION DEFAULT 0,
			depends_on TEXT[] DEFAULT '{}',
			status TEXT DEFAULT 'pending',
			FOREIGN KEY (expression_id) REFERENCES expressions(id) ON DELETE CASCADE
		);
	`
	_, err = db.Exec(createTasksTableQuery)
	if err != nil {
		log.Fatalf("Ошибка при создании таблицы tasks: %v", err)
	}
	log.Printf("Таблица tasks успешно создана или уже существует.")

	// Пример: Создание таблицы для пользователей (если нужно)
	createUsersTableQuery := `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL
		);
	`
	_, err = db.Exec(createUsersTableQuery)
	if err != nil {
		log.Fatalf("Ошибка при создании таблицы users: %v", err)
	}
	log.Printf("Таблица users успешно создана или уже существует.")
}
