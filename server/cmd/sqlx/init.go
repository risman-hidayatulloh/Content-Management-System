package sqlx

import (
	"database/sql"
	"fullstack-cms/config"

	_ "github.com/lib/pq"
)

func InitMigrationConnection(config *config.Config) (*sql.DB, error) {
	return sql.Open("postgres", config.Connection.Postgresql.DSN)
}
