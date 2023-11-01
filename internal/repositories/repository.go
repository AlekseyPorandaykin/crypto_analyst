package repositories

import (
	"fmt"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

const SeparateParamsInSQL = ","

type Config struct {
	Driver   string
	Username string
	Password string
	Host     string
	Port     string
	Database string
}

func CreateDB(conf Config) (*sqlx.DB, error) {
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
	return conn, nil
}

func strToSQLString(params []string) string {
	res := ""
	for _, param := range params {
		if len(res) > 0 {
			res += SeparateParamsInSQL
		}
		res += fmt.Sprintf("'%s'", param)
	}
	return res
}
