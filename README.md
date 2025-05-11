# Калькулятор: микросервисное приложение

Проект представляет собой микросервисную систему на Go для обработки арифметических выражений с использованием gRPC, 
REST API, JWT и хранения данных в SQLite.

---

## Технологии
- **Язык:** Go 1.23.4
- **API:** REST (JSON), gRPC
- **База данных:** SQLite
- **Безопасность:** JWT, bcrypt
- **Тестирование:** `go test`, интеграционные тесты с тегом `integration`

---

## Архитектура проекта

**Основные компоненты:**

- `orchestrator/service`: бизнес-логика (регистрация, аутентификация, обработка выражений)
- `orchestrator/repository`: доступ к базе данных
- `orchestrator/grpc`: gRPC-сервер
- `internal/models`: структуры данных
- `internal/proto`: определения gRPC-протоколов

*Диаграмма архитектуры:*
![Диаграмма архитектуры](calculator_app/static/calc_diag.jpg) 

---


## Структура БД

**Таблицы:**

- `users`: логин, хэш пароля
- `tasks`: арифметические подзадачи, статус, зависимости, результат
- `expressions`: исходные выражения, итоговый результат и статус

---

## Безопасность

- Пароли хэшируются с помощью **bcrypt**
- Аутентификация реализована через **JWT**
- Все защищённые эндпоинты требуют заголовка `Authorization: Bearer <token>`

---

##  Основной функционал

- Поддержка операций: `+`, `-`, `*`, `/`, включая вложенные скобки
- Сервис разбивает выражение на подзадачи и обрабатывает их с помощью агентов
- Все данные пользователей и результаты сохраняются

---

##  REST API Оркестратора

- `POST /api/v1/register`: регистрация пользователя
- `POST /api/v1/login`: вход и получение JWT
- `POST /api/v1/calculate`: отправка выражения на вычисление
- `GET /api/v1/expressions`: список выражений пользователя
- `GET /api/v1/expressions/{id}`: информация по конкретному выражению

---

**Агент (Worker):**

Агент:

- Получает задачи через gRPC у оркестратора
- Выполняет операции с задержкой (зависит от конфигурации)
- Отправляет результат обратно через gRPC

---

## Статусы task и expression

### `task`:
  - `pending` - создана новая задача 
  - `processing` - задача взята в обработку
  - `completed` - задача завершена
  - `division_by_zero` - ошибка задачи , деление на ноль
  - `unknown_operation` - неизвестная операция 
  - `internal_error` - внутренняя ошибка

### `expression`:
  - `pending` - создано новое выражение
  - `done` - выполнена
  - `division_by_zero` - ошибка выражения, деление на ноль
  - `unknown_operation` - неизвестная операция
  - `internal_error` - внутренняя ошибка 

## Установка и запуск

Скопируйте проект
```bash
git clone https://github.com/kabashin99/calculator.git
cd calculator
go mod tidy
```


### Запуск оркестратора (сервера)

- **Linux/macOS: / Windows (PowerShell)**
из корневой директории проекта : 
```bash
go run cmd/orchestrator/main.go

```

Лог успешного запуска:
```
2025/05/11 15:19:59 Config loaded completed
Running SQLite DB migrations...
Migrating table: users
Migrating table: expressions
Migrating table: tasks
DB migrations completed
2025/05/11 15:19:59 HTTP сервер запущен на localhost:8080
2025/05/11 15:19:59 gRPC сервер запущен на порту 50051

```

### Запуск агента (воркер)

В отдельном терминале запустите агента. 

- **Linux/macOS: / Windows (PowerShell)**
  из корневой директории проекта :
  ```bash
  go run cmd/agent/main.go
  
  ```
Сообщение при успешном запуске агента
```bash
2025/05/11 15:20:31 Config loaded completed
2025/05/11 15:20:31 Agent started with 4 workers     
2025/05/11 15:20:31 gRPC агент запущен на порту 50051

```

### Конфигурация
Файл *config/config.txt* 
```
# Конфигурация оркестратора
TIME_ADDITION_MS=100  #  время выполнения операции сложения в миллисекундах 
TIME_SUBTRACTION_MS=100  # время выполнения операции вычитания в миллисекундах 
TIME_MULTIPLICATION_MS=200  #  время выполнения операции умножения в 
TIME_DIVISION_MS=200  # время выполнения операции деления в миллисекундах

# Конфигурация агента
COMPUTING_POWER=4  # Количество горутин 
```


## Примеры запросов/ ответов

### 1. Регистрация нового пользователя
_Запрос:_
```bash
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"login":"user1","password":"pwd"}'
```
_Ответ:_
#### Удачный ответ, http код 200
```json
{"status":"user 'test5' created successfully"}
```
#### Ошибка попытка повторной регистрации , http код 409
```json
Registration failed: failed to register user: user already exists
``` 

#### Ошибка некорректный запрос , http код 400
```json
Invalid request
```

### 2. Вход зарегистрированного пользователя
_Запрос:_
```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"login":"user1","password":"pwd"}'
```
_Ответ:_
#### Удачный ответ , http код 200
```json
    "expires_at": "2025-05-12T12:49:14+03:00",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDcwNDMzNTQsImlhdCI6MTc0Njk1Njk1NCwibG9naW4iOiJ0ZXN0MiJ9.SpxcS9a5jgR0LMtC-fq9AcYtRN5jg7zgdQ5iYnAlou4"
```

#### Ошибка некорректный запрос , http код 400
```json
Invalid request
```

#### Ошибка ошибка аутентификации , http код 401
```json
 Authentication failed
```

### 3. Отправка выражения
Для удачного запроса в Authorization перенести токен , пример :
'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDcwNDMzNTQsImlhdCI6MTc0Njk1Njk1NCwibG9naW4iOiJ0ZXN0MiJ9.SpxcS9a5jgR0LMtC-fq9AcYtRN5jg7zgdQ5iYnAlou4'

_Запрос:_
```bash
curl -X POST http://localhost:8080/api/v1/calculate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{"expression":"2+2*2"}'
```

_Ответ:_
#### Удачный ответ , http код 201
```json
{"id":"550cf23a-4cd3-40d8-b1df-820d44c23479"}
```

#### Ошибка некорректный запрос , http код 400
```
Invalid request
```

#### Ошибка ошибка аутентификации , http код 401
```
 Authentication failed
```

#### Ошибка пользователь не найден , http код 403
```
 User not found
```

#### Ошибка выражения , http код 422
```
 invalid request
```

#### Ошибка сервера, http код 500
```
Internal server error
```



### 4. Список выражений
Для удачного запроса в Authorization перенести токен , пример :
'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDcwNDMzNTQsImlhdCI6MTc0Njk1Njk1NCwibG9naW4iOiJ0ZXN0MiJ9.SpxcS9a5jgR0LMtC-fq9AcYtRN5jg7zgdQ5iYnAlou4'

_Запрос:_
```bash
curl http://localhost:8080/api/v1/expressions \
  -H "Authorization: Bearer <TOKEN>"
```

_Ответ:_
#### Удачный ответ, http код 200
```json
{"expressions":
  [
    {"id":"ee409ffe-dd05-430b-bd3f-80b239b14a2d","status":"division_by_zero","result":0,"owner":"test2"},
    {"id":"fd3c0722-e236-4c9b-a837-1fd7df54173a","status":"done","result":33,"owner":"test2"}
  ]
}
```
#### Ошибка аутентификации, http код 401
```
 Authentication failed
```

#### Ошибка пользователь не найден, http код 403
```
 User not found
```

#### Ошибка сервера, http код 500
```
Internal server error
```

## 5. Информация по выражению
Для удачного запроса в Authorization перенести токен , пример :
'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDcwNDMzNTQsImlhdCI6MTc0Njk1Njk1NCwibG9naW4iOiJ0ZXN0MiJ9.SpxcS9a5jgR0LMtC-fq9AcYtRN5jg7zgdQ5iYnAlou4'

Внести в адрес http://localhost:8080/api/v1/expressions/{id} id выражения 

_Запрос:_
```bash
curl http://localhost:8080/api/v1/expressions/<id> \
  -H "Authorization: Bearer <TOKEN>"
```

_Ответ:_
```json
{"expression":{"id":"fd980e11-f026-420c-aee7-8b71b2f2e0f3","status":"done","result":33,"owner":"test2"}}

```
#### Ошибка аутентификации, http код 401
```
 Authentication failed
```

#### Ошибка пользователь не найден, http код 403
```
 User not found
```

#### Ошибка id не найден, http код 404
```
expression not found
```

#### Ошибка сервера, http код 500
```
Internal server error
```


## Тестирование

Юнит-тесты:
```bash
go test -cover ./...

```

Интеграционные тесты:
```bash
go test -tags=integration ./cmd/orchestrator

```
## Примечание
Проект протестирован на Windows.
Время выполнения операций управляется через конфигурацию.

Автор: Абашин Ярослав
Telegram: @kabashin
