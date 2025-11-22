#!/bin/bash

set -e

echo "🔍 Проверка CI workflow локально..."
echo ""

echo "1️⃣  Проверка Go версии..."
go version
echo ""

echo "2️⃣  Загрузка зависимостей..."
go mod download
echo "✅ Зависимости загружены"
echo ""

echo "3️⃣  Запуск тестов..."
go test -tags=integration ./internal/service/... -v
echo "✅ Тесты прошли"
echo ""

echo "4️⃣  Проверка линтера..."
if command -v golangci-lint &> /dev/null; then
    golangci-lint run --timeout=5m ./...
    echo "✅ Линтер прошел"
else
    echo "⚠️  golangci-lint не установлен"
    echo "   Установите: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
fi
echo ""

echo "✅ Все проверки завершены!"

