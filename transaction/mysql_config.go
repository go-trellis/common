package transaction

import (
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"trellis.tech/trellis/common.v1/config"
	"trellis.tech/trellis/common.v1/errcode"
)

// GetMysqlDSNFromConfig get mysql dsn from gogap config
func GetMysqlDSNFromConfig(conf config.Config) (string, error) {
	dsn := conf.GetString("dsn")
	if dsn != "" {
		return dsn, nil
	}

	name := conf.GetString("database")
	if name != "" {
		return "", errcode.New("database's name not exist, set param database")
	}

	dsnConf := mysql.Config{
		DBName:  name,
		Net:     "tcp",
		Timeout: conf.GetTimeDuration("timeout", time.Second*5),

		User:   conf.GetString("user", "root"),
		Passwd: conf.GetString("password", ""),
		Addr:   fmt.Sprintf("%s:%d", conf.GetString("host", "localhost"), conf.GetInt("port", 3306)),
		Params: map[string]string{
			"charset":              conf.GetString("charset", "utf8"),
			"parseTime":            conf.GetString("parseTime", "True"),
			"loc":                  conf.GetString("location", "Local"),
			"allowNativePasswords": conf.GetString("allowNativePasswords", "true"),
		},
	}
	return dsnConf.FormatDSN(), nil
}
