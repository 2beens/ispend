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
	var name string
	sqlStatement := `SELECT name FROM spend_kinds WHERE id=$1`
	row := pdb.db.QueryRow(sqlStatement, id)

	err := row.Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		log.Errorf("postgres DB error 10003: " + err.Error())
		return nil, err
	}
	return &SpendKind{
		ID:   id,
		Name: name,
	}, nil
}

func (pdb *PostgresDBClient) GetUserIDByUsername(username string) (int, error) {
	var id int
	sqlStatement := `SELECT id FROM users WHERE username=$1`
	row := pdb.db.QueryRow(sqlStatement, username)
	err := row.Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return -1, ErrNotFound
		}
		log.Errorf("postgres DB error 10021: " + err.Error())
		return -1, err
	}
	return id, nil
}

func (pdb *PostgresDBClient) GetSpendKinds(username string) ([]SpendKind, error) {
	userId, err := pdb.GetUserIDByUsername(username)
	if err != nil {
		return nil, err
	}

	sqlStatement := `SELECT id, name FROM spend_kinds WHERE user_id=$1`
	rows, err := pdb.db.Query(sqlStatement, userId)
	defer pdb.closeRows(rows)
	if err != nil {
		return nil, err
	}

	var spendKinds []SpendKind
	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			return nil, err
		}
		spendKinds = append(spendKinds, SpendKind{
			ID:   id,
			Name: name,
		})
	}

	return spendKinds, nil
}

func (pdb *PostgresDBClient) StoreUser(user *User) error {
	return nil
}

func (pdb *PostgresDBClient) GetUser(username string) (*User, error) {
	var id int
	var email, password string
	sqlStatement := `SELECT * FROM users WHERE username=$1`
	row := pdb.db.QueryRow(sqlStatement, username)

	err := row.Scan(&id, &email, &username, &password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		log.Errorf("postgres DB error 10011: " + err.Error())
		return nil, err
	}

	var spends []Spending
	var spendKinds []SpendKind

	spends, err = pdb.GetSpends(username)
	if err != nil {
		return nil, err
	}

	spendKinds, err = pdb.GetSpendKinds(username)
	if err != nil {
		return nil, err
	}

	return &User{
		Email:      email,
		Username:   username,
		Password:   password,
		Spends:     spends,
		SpendKinds: spendKinds,
	}, nil
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

		spends, err = pdb.GetSpends(username)
		if err != nil {
			return nil, err
		}

		spendKinds, err = pdb.GetSpendKinds(username)
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

func (pdb *PostgresDBClient) GetSpends(username string) ([]Spending, error) {
	userId, err := pdb.GetUserIDByUsername(username)
	if err != nil {
		return nil, err
	}

	rows, err := pdb.db.Query("SELECT * FROM spends WHERE user_id=$1", userId)
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
