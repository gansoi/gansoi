package mysql

import (
	"database/sql"
	"strconv"

	"github.com/gansoi/gansoi/plugins"

	// We need the MySQL driver for this.
	_ "github.com/go-sql-driver/mysql"
)

// MySQL retrieves metrics from a MySQL server.
type MySQL struct {
	DSN string `toml:"dsn" json:"dsn" description:"Mysql DSN"`
}

func init() {
	plugins.RegisterAgent("mysql", MySQL{})
}

// Check implements plugins.Agent.
func (m *MySQL) Check(result plugins.AgentResult) error {
	db, err := sql.Open("mysql", m.DSN)
	if err != nil {
		return err
	}
	defer db.Close()

	rows, err := db.Query("SHOW GLOBAL STATUS")
	if err != nil {
		return err
	}

	defer rows.Close()

	var name, value string

	for rows.Next() {
		e := rows.Scan(&name, &value)
		if e == nil {
			i, e := strconv.ParseInt(value, 10, 64)
			if e != nil {
				// Error, value is not integer
				result.AddValue(name, value)
			} else {
				result.AddValue(name, i)
			}
		}
	}

	return nil
}

// Ensure compliance
var _ plugins.Agent = (*MySQL)(nil)
