package ispend

import (
	"database/sql"

	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

func TestPostgresDB() {
	sslMode := "disable"
	connStr := "user=2beens dbname=ispenddb sslmode=" + sslMode
	db, err := sql.Open("postgres", connStr)
	defer func() {
		if db != nil {
			db.Close()
		}
	}()

	if err != nil {
		log.Errorf("cannot open PS DB connection: %s", err.Error())
		return
	}

	rows, err := db.Query("SELECT * FROM users")
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()
	if err != nil {
		log.Errorf("cannot query PS DB: %s", err.Error())
		return
	}

	columns, err := rows.Columns()
	if err != nil {
		log.Errorf(err.Error())
	}

	log.Debugf("test db - users count: %d", len(columns))
}
