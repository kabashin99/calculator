#!/bin/bash

# Запуск оркестратора
go run cmd/orchestrator/main.go &

# Запуск агентов
COMPUTING_POWER=4 ORCHESTRATOR_URL=http://localhost:8080 go run cmd/agent/main.go &
COMPUTING_POWER=2 ORCHESTRATOR_URL=http://localhost:8080 go run cmd/agent/main.go &

wait
