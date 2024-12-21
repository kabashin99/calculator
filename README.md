# Проект Калькулятор

Цель проекта: HTTP-сервер, который обрабатывает входящие арифметические выражения и возвращает результаты вычислений.

## Функциональность

Сервер принимает на вход строку, содержащую арифметическое выражение. Строка может включать:

• Цифры (рациональные числа), представленные в виде односимвольных идентификаторов.

• Арифметические операции: сложение (+), вычитание (-), умножение (*) и деление (/).

• Скобки ( и ), которые используются для задания приоритета выполнения операций.

## Установка и запуск

Скопируйте проект
```bash
git clone https://github.com/kabashin99/calculator.git
```

Установите все необходимые зависимости
```bash
go mod tidy
```

Запустите сервер
```bash
go run .
```

Сообщение при успешном запуске сервера
```bash
2024/12/20 21:40:10 Сервер запущен на :8080
```

Запросы отправляются по адресу POST <http://localhost:8080/api/v1/calculate>

В случае сообщения об ошибке: 
```bash
2024/12/20 22:40:26 Ошибка при запуске сервера: listen tcp :8080: bind: Only one usage of each socket address (protocol/network address/port) is normally permitted.
exit status 1 
```
Поменяйте в файле `main.go` порт *8080* на любой свободный


## Примеры запросов/ ответов

Запрос 
```bash
curl -X POST http://localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d '{"expression": "3 + 5"}'
```

Запрос (для Windows если не работает *curl*)
```powershell
$headers = @{"Content-Type" = "application/json"}
$body = '{"expression": "3 + 5"}'
Invoke-WebRequest -Uri "http://127.0.0.1:8080/api/v1/calculate" `
    -Method Post `
    -Headers $headers `
    -Body $body
```

Ответ 
```
StatusCode        : 200
StatusDescription : OK
Content           : {"result":"8.000000"}
```

## Тесты

Запуск тестов 
```bash
go test -cover calculator_app/internal/...
```

## Документация
Документация в формате swagger по методам API <http://localhost:8081/swagger/> (доступна после запуска сервера!)
В случае сообщения об ошибке:
```bash
2024/12/20 22:40:26 Ошибка при запуске сервера: listen tcp :8081: bind: Only one usage of each socket address (protocol/network address/port) is normally permitted.
exit status 1 
```
Поменяйте в файле `main.go` порт *8081* на любой свободный

Автор: Абашин Ярослав
Telegram: @kabashin
