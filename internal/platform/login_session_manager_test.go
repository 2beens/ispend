package platform_test

import (
	"testing"

	"github.com/2beens/ispend/internal/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoginSessionManager(t *testing.T) {
	sessionManager := platform.NewLoginSessionHandler()
	require.NotNil(t, sessionManager)

	username1 := "u1"
	username2 := "u2"
	username3 := "u3-not-added"

	sessionId1 := sessionManager.New(username1)
	assert.True(t, len(sessionId1) > 0)
	sessionId2 := sessionManager.New(username2)
	assert.True(t, len(sessionId2) > 0)
	sessionId3 := "not-existing-sessionID"

	// assert logged in
	isLoggedUser1 := sessionManager.IsUserLoggedIn(sessionId1, username1)
	assert.True(t, isLoggedUser1)
	isLoggedUser2 := sessionManager.IsUserLoggedIn(sessionId2, username2)
	assert.True(t, isLoggedUser2)
	isLoggedUser3 := sessionManager.IsUserLoggedIn(sessionId3, username3)
	assert.False(t, isLoggedUser3)
	isLoggedUser12 := sessionManager.IsUserLoggedIn(sessionId1, username2)
	assert.False(t, isLoggedUser12)

	// get by session id
	session1, err := sessionManager.GetBySessionID(sessionId1)
	assert.NoError(t, err)
	assert.NotNil(t, session1)
	assert.Equal(t, sessionId1, session1.SessionID)
	assert.Equal(t, username1, session1.Username)
	session2, err := sessionManager.GetBySessionID(sessionId2)
	assert.NoError(t, err)
	assert.NotNil(t, session2)
	assert.Equal(t, sessionId2, session2.SessionID)
	assert.Equal(t, username2, session2.Username)
	session3, err := sessionManager.GetBySessionID(sessionId3)
	assert.Equal(t, platform.ErrNotFound, err)
	assert.Nil(t, session3)

	// get by username
	session1, err = sessionManager.GetByUsername(username1)
	assert.NoError(t, err)
	assert.NotNil(t, session1)
	assert.Equal(t, sessionId1, session1.SessionID)
	assert.Equal(t, username1, session1.Username)
	session2, err = sessionManager.GetByUsername(username2)
	assert.NoError(t, err)
	assert.NotNil(t, session2)
	assert.Equal(t, sessionId2, session2.SessionID)
	assert.Equal(t, username2, session2.Username)
	session3, err = sessionManager.GetByUsername(username3)
	assert.Equal(t, platform.ErrNotFound, err)
	assert.Nil(t, session3)

	// remove session
	err = sessionManager.Remove(username1)
	assert.NoError(t, err)
	session1, err = sessionManager.GetByUsername(username1)
	assert.Equal(t, platform.ErrNotFound, err)
	assert.Nil(t, session1)
	session1, err = sessionManager.GetBySessionID(sessionId1)
	assert.Equal(t, platform.ErrNotFound, err)
	assert.Nil(t, session1)
}
