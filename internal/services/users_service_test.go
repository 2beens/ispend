package services_test

import (
	"testing"

	"github.com/2beens/ispend/internal/db"
	"github.com/2beens/ispend/internal/metrics"
	"github.com/2beens/ispend/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAllUsers(t *testing.T) {
	usersService := getUserServiceTest()
	allUsers, err := usersService.GetAllUsers()
	require.NoError(t, err)
	assert.NotNil(t, allUsers)
}

func getUserServiceTest() *services.UsersService {
	db := db.NewInMemoryDB()
	graphiteClient := metrics.NewGraphiteNop("test.graphite.host", 1000)
	us := services.NewUsersService(db, graphiteClient)
	return us
}
