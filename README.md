# Проект Калькулятор

Цель проекта совершать математические вычисления по полученным параметрам. Для удобства пользования создан http-сервер
для получения запросов по сети. 

## Возможности

+-*/

## Установка и запуск

Копирование проекта
```bash
git clone https://github.com/kabashin99/calculator.git
```
Запуск сервера
```bash
go run .
```
Запросы отправляются по адресу POST <http://localhost:8080/api/v1/calculate>

## Примеры запросов/ ответов

Запрос 
```bash
curl -X POST http://localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d '{"expression": "3 + 5"}'
```

Ответ 
```bash

```

## Тесты

Запуск тестов 
```bash
go test -cover ./tests
```

## Документация
Документация в формате swagger по методам API <http://localhost:8080/docs>


Автор: Абашин Ярослав