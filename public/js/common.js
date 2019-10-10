function getLoggedUser() {
    const username = localStorage.getItem("username");
    const sessionId = localStorage.getItem("sessionId");
    const isLogged = username !== "" && sessionId !== "";
    return {username: username, sessionId: sessionId, isLogged: isLogged};
}