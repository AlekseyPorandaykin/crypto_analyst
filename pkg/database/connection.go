package database

import (
	"fmt"

	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	Driver             string
	Username           string
	Password           string
	Host               string
	Port               string
	Database           string
	MaxOpenConnections int
	MaxIdleConnections int
}

func CreateConnection(conf Config) (*sqlx.DB, error) {
	switch conf.Driver {
	case "postgres":
		return CreatePostgresConnection(conf)
	default:
		return nil, fmt.Errorf("not found connection for driver: %s", conf.Driver)
	}
}

func CreatePostgresConnection(conf Config) (*sqlx.DB, error) {
	conn, err := sqlx.Connect(
		"pgx",
		fmt.Sprintf(
			"%s://%s:%s@%s:%s/%s",
			conf.Driver,
			conf.Username,
			conf.Password,
			conf.Host,
			conf.Port,
			conf.Database,
		),
	)
	if err != nil {
		return nil, err
	}
	if conf.MaxOpenConnections > 0 {
		conn.SetMaxOpenConns(conf.MaxOpenConnections)
	}
	if conf.MaxIdleConnections > 0 {
		conn.SetMaxIdleConns(conf.MaxIdleConnections)
	}
	return conn, nil
}
