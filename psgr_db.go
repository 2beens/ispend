package ispend

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

type PostgresDBClient struct {
	db      *sql.DB
	sslMode string
	dbUser  string
	dbName  string
}

func NewPostgresDBClient(dbName string, dbUser string, sslMode string) *PostgresDBClient {
	return &PostgresDBClient{
		sslMode: sslMode,
		dbUser:  dbUser,
		dbName:  dbName,
	}
}

func (pdb *PostgresDBClient) Open() error {
	connStr := fmt.Sprintf("user=%s dbname=%s sslmode=%s", pdb.dbUser, pdb.dbName, pdb.sslMode)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	pdb.db = db
	return nil
}

func (pdb *PostgresDBClient) Close() error {
	if pdb.db == nil {
		return errors.New("postgres DB client is nil, cannot close")
	}
	err := pdb.db.Close()
	if err != nil {
		return err
	}
	return nil
}

func (pdb *PostgresDBClient) TestGetAllUsers() {
	rows, err := pdb.db.Query("SELECT * FROM users")
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

	log.Debugf("test db - users, columns count: %d", len(columns))
}

func (pdb *PostgresDBClient) TestSelectRow() {
	username := "admin"
	var column string
	sqlStatement := `SELECT email FROM users WHERE username=$1`
	row := pdb.db.QueryRow(sqlStatement, username)

	log.Debugf("postgres DB, testing [%s]", sqlStatement)

	err := row.Scan(&column)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Debug("zero rows found ...")
		} else {
			log.Warnf("postgres DB, testing error: " + err.Error())
		}
		return
	}
	log.Warnf("postgres DB, testing result: " + column)
}
