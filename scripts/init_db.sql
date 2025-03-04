-- Создание таблицы выражений
CREATE TABLE IF NOT EXISTS expressions (
                                           id TEXT PRIMARY KEY,
                                           status TEXT NOT NULL DEFAULT 'pending', -- pending/processing/done/error
                                           result REAL,
                                           created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы задач
CREATE TABLE IF NOT EXISTS tasks (
                                     id TEXT PRIMARY KEY,
                                     expression_id TEXT NOT NULL,
                                     arg1 REAL NOT NULL,
                                     arg2 REAL NOT NULL,
                                     operation TEXT NOT NULL, -- +, -, *, /
                                     status TEXT NOT NULL DEFAULT 'pending', -- pending/completed/error
                                     result REAL,
                                     FOREIGN KEY (expression_id) REFERENCES expressions(id)
    );

-- Индексы для ускорения поиска
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_expressions_status ON expressions(status);