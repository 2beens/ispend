package db

import (
	"github.com/2beens/ispend/internal/models"
)

type SpenderDB interface {
	Open() error
	Close() error

	StoreDefaultSpendKind(kind models.SpendKind) (int, error)
	GetAllDefaultSpendKinds() ([]models.SpendKind, error)
	GetSpendKind(username string, spendingKindID int) (*models.SpendKind, error)
	GetSpendKinds(username string) ([]models.SpendKind, error)
	StoreSpendKind(username string, kind *models.SpendKind) (int, error)

	StoreUser(user *models.User) (int, error)
	GetUser(username string, loadAllData bool) (*models.User, error)
	GetAllUsers(loadAllUserData bool) (models.Users, error)

	StoreSpending(username string, spending models.Spending) (string, error)
	GetSpends(username string) ([]models.Spending, error)
	DeleteSpending(username, spendID string) error
}
