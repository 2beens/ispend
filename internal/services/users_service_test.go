package services_test

import (
	"sync"
	"testing"
	"time"

	"github.com/2beens/ispend/internal/db"
	"github.com/2beens/ispend/internal/metrics"
	"github.com/2beens/ispend/internal/models"
	"github.com/2beens/ispend/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAllUsers(t *testing.T) {
	usersService := getUserServiceTest()
	allUsers, err := usersService.GetAllUsers()
	require.NoError(t, err)
	assert.Len(t, allUsers, 2)
}

func TestStoreAndRetrieveUser(t *testing.T) {
	usersService := getUserServiceTest()
	allUsersBefore, err := usersService.GetAllUsers()
	require.NoError(t, err)

	spendKind := &models.SpendKind{
		ID:   1,
		Name: "sk1",
	}
	spend := &models.Spending{
		ID:        "spend1",
		Currency:  "rsd",
		Amount:    120,
		Kind:      spendKind,
		Timestamp: time.Now(),
	}
	username := "user1"
	user := &models.User{
		Email:      "email1",
		Username:   username,
		Password:   "pass1",
		Spends:     []models.Spending{*spend},
		SpendKinds: []models.SpendKind{*spendKind},
	}
	err = usersService.AddUser(user)
	require.NoError(t, err)

	allUsersAfter, err := usersService.GetAllUsers()
	require.NoError(t, err)

	assert.Equal(t, len(allUsersBefore)+1, len(allUsersAfter))
	retrievedUser, err := usersService.GetUser(username)
	require.NoError(t, err)
	assert.Equal(t, username, retrievedUser.Username)
	assert.Len(t, retrievedUser.Spends, 1)
	assert.Len(t, retrievedUser.SpendKinds, 1)
}

func TestStoreAndRetrieveUser_Multithreaded(t *testing.T) {
	usersService := getUserServiceTest()
	allUsersBefore, err := usersService.GetAllUsers()
	require.NoError(t, err)
	require.Len(t, allUsersBefore, 2)

	storeUserTestFunc := func(username string) error {
		spendKind := &models.SpendKind{
			ID:   1,
			Name: "testID",
		}
		spend := &models.Spending{
			ID:        "testID",
			Currency:  "testC",
			Amount:    120,
			Kind:      spendKind,
			Timestamp: time.Now(),
		}
		user := &models.User{
			Email:      "testEmail",
			Username:   username,
			Password:   "testPass",
			Spends:     []models.Spending{*spend},
			SpendKinds: []models.SpendKind{*spendKind},
		}
		return usersService.AddUser(user)
	}

	usersCount := 5
	var wg sync.WaitGroup
	for i := 1; i <= usersCount; i++ {
		wg.Add(1)
		username := "username" + string(i)
		go func(t *testing.T) {
			err := storeUserTestFunc(username)
			assert.NoError(t, err)
			wg.Done()
		}(t)
	}

	wg.Wait()

	allUsers, err := usersService.GetAllUsers()
	assert.NoError(t, err)
	assert.Len(t, allUsers, len(allUsersBefore)+usersCount)

	for i := 1; i <= usersCount; i++ {
		username := "username" + string(i)
		user, err := usersService.GetUser(username)
		assert.NoError(t, err)
		assert.Equal(t, username, user.Username)
	}
}

func getUserServiceTest() *services.UsersService {
	inMemDB := db.NewInMemoryDB()
	graphiteClient := metrics.NewGraphiteNop("test.graphite.host", 1000)
	us := services.NewUsersService(inMemDB, graphiteClient)
	return us
}
