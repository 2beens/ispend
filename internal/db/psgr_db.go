package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/2beens/ispend/internal/models"
	"github.com/2beens/ispend/internal/platform"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

type PostgresDBClient struct {
	db          *sql.DB
	sslMode     string
	dbHost      string
	dbPort      int
	dbUser      string
	dbName      string
	dbPassword  string
	pingTimeout int
}

func NewPostgresDBClient(dbHost string, dbPort int, dbName string, dbUser string, dbPassword string, sslMode string, pingTimeout int) *PostgresDBClient {
	return &PostgresDBClient{
		sslMode:     sslMode,
		dbHost:      dbHost,
		dbPort:      dbPort,
		dbUser:      dbUser,
		dbPassword:  dbPassword,
		dbName:      dbName,
		pingTimeout: pingTimeout,
	}
}

func (pdb *PostgresDBClient) Open() error {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		pdb.dbHost, pdb.dbPort, pdb.dbUser, pdb.dbPassword, pdb.dbName, pdb.sslMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	pingDoneCh := make(chan error, 1)
	go func() {
		err = db.Ping()
		if err != nil {
			log.Error("failed to ping postgres usersService")
		} else {
			log.Trace("successful postgres DB ping!")
		}
		pingDoneCh <- err
	}()

	select {
	case pingErr := <-pingDoneCh:
		if pingErr == nil {
			log.Debugf("successfully connected to postgres usersService at: %s:%d", pdb.dbHost, pdb.dbPort)
			pdb.db = db
			return nil
		}
		return pingErr
	case <-time.After(time.Duration(pdb.pingTimeout) * time.Second):
		return fmt.Errorf("timeout [after %d seconds] fatal error - cannot ping database", pdb.pingTimeout)
	}
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

func (pdb *PostgresDBClient) StoreDefaultSpendKind(kind models.SpendKind) (int, error) {
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

func (pdb *PostgresDBClient) GetAllDefaultSpendKinds() ([]models.SpendKind, error) {
	rows, err := pdb.db.Query("SELECT * FROM default_spend_kinds")
	defer pdb.closeRows(rows)
	if err != nil {
		return nil, err
	}

	var spendKinds []models.SpendKind
	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			return nil, err
		}

		spendKinds = append(spendKinds, models.SpendKind{
			ID:   id,
			Name: name,
		})
	}

	return spendKinds, nil
}

func (pdb *PostgresDBClient) GetSpendKind(username string, spendingKindID int) (*models.SpendKind, error) {
	// TODO: can maybe use just GetSpendKindByID(id int) instead of this one

	userId, err := pdb.GetUserIDByUsername(username)
	if err != nil {
		return nil, err
	}

	var name string
	sqlStatement := `SELECT name FROM spend_kinds WHERE id=$1 AND user_id=$2`
	row := pdb.db.QueryRow(sqlStatement, spendingKindID, userId)

	err = row.Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, platform.ErrNotFound
		}
		log.Errorf("postgres DB error 10023: " + err.Error())
		return nil, err
	}
	return &models.SpendKind{
		ID:   spendingKindID,
		Name: name,
	}, nil
}

func (pdb *PostgresDBClient) GetSpendKindByID(id int) (*models.SpendKind, error) {
	var name string
	sqlStatement := `SELECT name FROM spend_kinds WHERE id=$1`
	row := pdb.db.QueryRow(sqlStatement, id)

	err := row.Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, platform.ErrNotFound
		}
		log.Errorf("postgres DB error 10003: " + err.Error())
		return nil, err
	}
	return &models.SpendKind{
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
			return -1, platform.ErrNotFound
		}
		log.Errorf("postgres DB error 10021: " + err.Error())
		return -1, err
	}
	return id, nil
}

func (pdb *PostgresDBClient) GetSpendKinds(username string) ([]models.SpendKind, error) {
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

	var spendKinds []models.SpendKind
	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			return nil, err
		}
		spendKinds = append(spendKinds, models.SpendKind{
			ID:   id,
			Name: name,
		})
	}

	return spendKinds, nil
}

func (pdb *PostgresDBClient) SpendKindExistsForUser(userId int, kindName string) (bool, error) {
	var id int
	sqlStatement := `SELECT id FROM spend_kinds WHERE user_id=$1 AND name=$2`
	row := pdb.db.QueryRow(sqlStatement, userId, kindName)
	err := row.Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, platform.ErrNotFound
		}
		log.Errorf("postgres DB error 10022: " + err.Error())
		return false, err
	}
	return true, nil
}

func (pdb *PostgresDBClient) StoreSpendKind(username string, kind *models.SpendKind) (int, error) {
	userId, err := pdb.GetUserIDByUsername(username)
	if err != nil {
		return -1, err
	}

	sqlStatement := `
		INSERT INTO spend_kinds (user_id, name)
		VALUES ($1, $2)
		RETURNING id`
	id := -1
	err = pdb.db.QueryRow(sqlStatement, userId, kind.Name).Scan(&id)
	if err != nil {
		return id, err
	}
	return id, nil
}

func (pdb *PostgresDBClient) StoreUser(user *models.User) (int, error) {
	sqlStatement := `
		INSERT INTO users (email, username, password)
		VALUES ($1, $2, $3)
		RETURNING id`
	id := 0
	err := pdb.db.QueryRow(sqlStatement, user.Email, user.Username, user.Password).Scan(&id)
	if err != nil {
		return id, err
	}

	for i := range user.SpendKinds {
		spendKindID, err := pdb.StoreSpendKind(user.Username, &user.SpendKinds[i])
		if err != nil {
			log.Errorf("postgres DB client store user - store spend kind error: %s", err)
			continue
		}
		user.SpendKinds[i].ID = spendKindID
	}

	for i := range user.Spends {
		spendId, err := pdb.StoreSpending(user.Username, user.Spends[i])
		if err != nil {
			log.Errorf("postgres DB client store user - store spending error: %s", err)
			continue
		}
		user.Spends[i].ID = spendId
	}

	return id, nil
}

func (pdb *PostgresDBClient) GetUser(username string, loadAllData bool) (*models.User, error) {
	var id int
	var email, password string
	sqlStatement := `SELECT * FROM users WHERE username=$1`
	row := pdb.db.QueryRow(sqlStatement, username)

	err := row.Scan(&id, &email, &username, &password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, platform.ErrNotFound
		}
		log.Errorf("postgres DB error 10011: " + err.Error())
		return nil, err
	}

	var spends []models.Spending
	var spendKinds []models.SpendKind
	if loadAllData {
		spends, err = pdb.GetSpends(username)
		if err != nil {
			return nil, err
		}

		spendKinds, err = pdb.GetSpendKinds(username)
		if err != nil {
			return nil, err
		}
	}

	return &models.User{
		Email:      email,
		Username:   username,
		Password:   password,
		Spends:     spends,
		SpendKinds: spendKinds,
	}, nil
}

func (pdb *PostgresDBClient) GetAllUsers(loadAllUserData bool) (models.Users, error) {
	rows, err := pdb.db.Query("SELECT * FROM users")
	defer pdb.closeRows(rows)
	if err != nil {
		return nil, err
	}

	var users models.Users
	for rows.Next() {
		var id int
		var email, username, password string
		err = rows.Scan(&id, &email, &username, &password)
		if err != nil {
			return nil, err
		}

		var spends []models.Spending
		var spendKinds []models.SpendKind
		if loadAllUserData {
			spends, err = pdb.GetSpends(username)
			if err != nil {
				return nil, err
			}

			spendKinds, err = pdb.GetSpendKinds(username)
			if err != nil {
				return nil, err
			}
		}

		users = append(users, &models.User{
			Email:      email,
			Username:   username,
			Password:   password,
			Spends:     spends,
			SpendKinds: spendKinds,
		})
	}

	return users, nil
}

func (pdb *PostgresDBClient) StoreSpending(username string, spending models.Spending) (string, error) {
	userId, err := pdb.GetUserIDByUsername(username)
	if err != nil {
		return "", err
	}

	var spendKindId int
	spendKindExists, err := pdb.SpendKindExistsForUser(userId, spending.Kind.Name)
	if err != nil {
		log.Errorf("StoreSpending error 12998: %s", err)
	}
	if spendKindExists {
		spendKindId = spending.Kind.ID
	} else {
		spendKindId, err = pdb.StoreSpendKind(username, spending.Kind)
		if err != nil {
			return "", err
		}
	}

	sqlStatement := `
		INSERT INTO spends (currency, amount, spend_timestamp, user_id, kind_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`
	id := 0
	err = pdb.db.QueryRow(
		sqlStatement, spending.Currency, spending.Amount, spending.Timestamp, userId, spendKindId,
	).Scan(&id)
	if err != nil {
		return "", err
	}
	return strconv.Itoa(id), nil
}

func (pdb *PostgresDBClient) GetSpends(username string) ([]models.Spending, error) {
	userId, err := pdb.GetUserIDByUsername(username)
	if err != nil {
		return nil, err
	}

	rows, err := pdb.db.Query("SELECT * FROM spends WHERE user_id=$1", userId)
	defer pdb.closeRows(rows)
	if err != nil {
		return nil, err
	}

	var spends []models.Spending
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
		spends = append(spends, models.Spending{
			ID:        id,
			Currency:  currency,
			Amount:    amount,
			Kind:      spendKind,
			Timestamp: timestamp,
		})
	}

	return spends, nil
}

func (pdb *PostgresDBClient) DeleteSpending(username, spendID string) error {
	log.Tracef("DB tries to delete spending [user: %s] [id: %s]...", username, spendID)
	userId, err := pdb.GetUserIDByUsername(username)
	if err != nil {
		return errors.New("user not found")
	}

	sqlStatement := `
		DELETE FROM spends
		WHERE id=$1 AND user_id=$2;`
	res, err := pdb.db.Exec(sqlStatement, spendID, userId)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if count <= 0 {
		return exec.ErrNotFound
	}

	log.Tracef("DB deleted spending [user: %s] [id: %s]", username, spendID)
	return nil
}

func (pdb *PostgresDBClient) closeRows(rows *sql.Rows) {
	if rows != nil {
		err := rows.Close()
		if err != nil {
			log.Error(err)
		}
	}
}
