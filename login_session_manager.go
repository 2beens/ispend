package ispend

type LoginSessionHandler struct {
	loginSessions map[string]LoginSession
}

func NewLoginSessionHandler() *LoginSessionHandler {
	return &LoginSessionHandler{
		loginSessions: make(map[string]LoginSession),
	}
}

func (h *LoginSessionHandler) New(username string) string {
	cookieID := GenerateRandomString(45)
	loginSession := LoginSession{
		Username: username,
		CookieID: cookieID,
	}

	h.loginSessions[username] = loginSession

	return cookieID
}

func (h *LoginSessionHandler) Remove(username string) error {
	if _, ok := h.loginSessions[username]; !ok {
		return ErrNotFound
	}
	delete(h.loginSessions, username)
	return nil
}

func (h *LoginSessionHandler) GetByUsername(username string) (*LoginSession, error) {
	if ls, ok := h.loginSessions[username]; ok {
		return &ls, nil
	}
	return nil, ErrNotFound
}

func (h *LoginSessionHandler) GetByCookieID(cookieID string) (*LoginSession, error) {
	for _, ls := range h.loginSessions {
		if ls.CookieID == cookieID {
			return &ls, nil
		}
	}
	return nil, ErrNotFound
}
