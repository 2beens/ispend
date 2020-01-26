function htmlToElement(html) {
    const template = document.createElement('template');
    template.innerHTML = html.trim();
    return template.content.firstChild;
}

function getLoggedUser() {
    const username = localStorage.getItem("username");
    const sessionId = localStorage.getItem("sessionId");
    const isLogged = username !== "" && sessionId !== "";
    return {username: username, sessionId: sessionId, isLogged: isLogged};
}

function clearLoginData() {
    localStorage.setItem("sessionId", "");
    localStorage.setItem("username", "");
    refreshLoggedUserInfo();
}

function getSessionID() {
    return localStorage.getItem("sessionId");
}

function getUsername() {
    return localStorage.getItem("username");
}
