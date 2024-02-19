package db

const (
	DatetimeFormat      = "2006-01-02 15:04:05"
	SeparateParamsInSQL = ","
)

type Config struct {
	Driver   string
	Username string
	Password string
	Host     string
	Port     string
	Database string
}
