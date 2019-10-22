package internal

type LoginSessionManager struct {
	loginSessions map[string]LoginSession
}

func NewLoginSessionHandler() *LoginSessionManager {
	return &LoginSessionManager{
		loginSessions: make(map[string]LoginSession),
	}
}

func (manager *LoginSessionManager) New(username string) string {
	sessionID := GenerateRandomString(45)
	loginSession := LoginSession{
		Username:  username,
		SessionID: sessionID,
	}

	manager.loginSessions[username] = loginSession

	return sessionID
}

func (manager *LoginSessionManager) Remove(username string) error {
	if _, ok := manager.loginSessions[username]; !ok {
		return ErrNotFound
	}
	delete(manager.loginSessions, username)
	return nil
}

func (manager *LoginSessionManager) GetByUsername(username string) (*LoginSession, error) {
	if ls, ok := manager.loginSessions[username]; ok {
		return &ls, nil
	}
	return nil, ErrNotFound
}

func (manager *LoginSessionManager) GetBySessionID(sessionID string) (*LoginSession, error) {
	for _, ls := range manager.loginSessions {
		if ls.SessionID == sessionID {
			return &ls, nil
		}
	}
	return nil, ErrNotFound
}

func (manager *LoginSessionManager) IsUserLoggedIn(sessionID, username string) bool {
	session, err := manager.GetBySessionID(sessionID)
	if err != nil {
		return false
	}
	if session.Username != username {
		return false
	}
	return true
}

func (manager *LoginSessionManager) IsUserNotLoggedIn(sessionID, username string) bool {
	return !manager.IsUserLoggedIn(sessionID, username)
}
