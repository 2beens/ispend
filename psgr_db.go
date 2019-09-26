package ispend

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

type PostgresDBClient struct {
	db         *sql.DB
	sslMode    string
	dbHost     string
	dbPort     int
	dbUser     string
	dbName     string
	dbPassword string
}

func NewPostgresDBClient(dbHost string, dbPort int, dbName string, dbUser string, dbPassword string, sslMode string) *PostgresDBClient {
	return &PostgresDBClient{
		sslMode:    sslMode,
		dbHost:     dbHost,
		dbPort:     dbPort,
		dbUser:     dbUser,
		dbPassword: dbPassword,
		dbName:     dbName,
	}
}

func (pdb *PostgresDBClient) Open() error {
	var connStr string
	if pdb.sslMode == "disable" {
		connStr = fmt.Sprintf(
			"host=%s port=%d user=%s dbname=%s sslmode=%s",
			pdb.dbHost, pdb.dbPort, pdb.dbUser, pdb.dbName, pdb.sslMode,
		)
	} else {
		connStr = fmt.Sprintf(
			"host=%s port=%d user=%s dbname=%s password=%s sslmode=%s",
			pdb.dbHost, pdb.dbPort, pdb.dbUser, pdb.dbName, pdb.dbPassword, pdb.sslMode,
		)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		log.Error("failed to ping postgres db")
		return err
	}

	log.Debugf("successfully connected to postgres db at: %s:%d", pdb.dbHost, pdb.dbPort)

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

func (pdb *PostgresDBClient) StoreDefaultSpendKind(kind SpendKind) (int, error) {
	sqlStatement := `
		INSERT INTO default_spend_kinds (name)
		VALUES ($1)
		RETURNING id`
	id := 0
	err := pdb.db.QueryRow(sqlStatement, kind.Name).Scan(&id)
	if err != nil {
		return id, err
	}
	return id, nil
}

func (pdb *PostgresDBClient) GetDefaultSpendKind(name string) (*SpendKind, error) {
	return nil, nil
}

func (pdb *PostgresDBClient) GetAllDefaultSpendKinds() ([]SpendKind, error) {
	return nil, nil
}

func (pdb *PostgresDBClient) GetSpendKind(username string, spendingKindID string) (*SpendKind, error) {
	return nil, nil
}

func (pdb *PostgresDBClient) GetSpendKindByID(id int) (*SpendKind, error) {
	var column string
	sqlStatement := `SELECT name FROM spend_kinds WHERE id=$1`
	row := pdb.db.QueryRow(sqlStatement, id)

	err := row.Scan(&column)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		log.Errorf("postgres DB error 10003: " + err.Error())
		return nil, err
	}
	return &SpendKind{
		ID:   id,
		Name: column,
	}, nil
}

func (pdb *PostgresDBClient) GetSpendKinds(username string) ([]SpendKind, error) {
	return nil, nil
}

func (pdb *PostgresDBClient) StoreUser(user *User) error {
	return nil
}

func (pdb *PostgresDBClient) GetUser(username string) (*User, error) {
	return nil, nil
}

func (pdb *PostgresDBClient) GetAllUsers() (Users, error) {
	rows, err := pdb.db.Query("SELECT * FROM users")
	defer pdb.closeRows(rows)
	if err != nil {
		return nil, err
	}

	var users Users
	for rows.Next() {
		var id int
		var email, username, password string
		err = rows.Scan(&id, &email, &username, &password)
		if err != nil {
			return nil, err
		}

		var spends []Spending
		var spendKinds []SpendKind

		spends, err = pdb.GetSpendings(username)
		if err != nil {
			return nil, err
		}

		users = append(users, &User{
			Email:      email,
			Username:   username,
			Password:   password,
			Spends:     spends,
			SpendKinds: spendKinds,
		})
	}

	return users, nil
}

func (pdb *PostgresDBClient) StoreSpending(username string, spending Spending) error {
	return nil
}

func (pdb *PostgresDBClient) GetSpendings(username string) ([]Spending, error) {
	rows, err := pdb.db.Query("SELECT * FROM spends")
	defer pdb.closeRows(rows)
	if err != nil {
		return nil, err
	}

	var spends []Spending
	for rows.Next() {
		var id, currency string
		var userId, kindId int
		var timestamp time.Time
		var amount float32
		err = rows.Scan(&id, &currency, &amount, &timestamp, &userId, &kindId)
		currency = strings.TrimSpace(currency)
		if err != nil {
			return nil, err
		}
		spendKind, err := pdb.GetSpendKindByID(kindId)
		if err != nil {
			return nil, err
		}
		spends = append(spends, Spending{
			ID:        id,
			Currency:  currency,
			Amount:    amount,
			Kind:      spendKind,
			Timestamp: timestamp,
		})
	}

	return spends, nil
}

func (pdb *PostgresDBClient) closeRows(rows *sql.Rows) {
	if rows != nil {
		err := rows.Close()
		if err != nil {
			log.Error(err)
		}
	}
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
