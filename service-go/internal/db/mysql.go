package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// NewDBConnection создает новое подключение к MySQL
func NewDBConnection() (*sql.DB, error) {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		// Дефолтное значение, если переменная не задана (для локальной разработки)
		dsn = "user:password@tcp(localhost:3306)/go_test?charset=utf8mb4&parseTime=True&loc=Local"
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("ошибка при открытии базы данных: %w", err)
	}

	// Настройка пула соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Проверка подключения
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ошибка при подключении к базе данных: %w", err)
	}

	return db, nil
}
