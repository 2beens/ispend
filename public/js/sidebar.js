function login() {
    const username = $('#form_username').val();
    const password = $('#form_password').val();
    if (!username || !password) {
        console.error('username | password empty');
        return;
    }

    $.ajax({
        url: "/users/login",
        type: "POST",
        dataType: "json",                 // expected format for response
        contentType: "application/x-www-form-urlencoded; charset=utf-8",
        data: {username: username, password: password},
        complete: function () {
            console.log('login request complete');
            $('#form_username').val('');
            $('#form_password').val('');
        },
        success: function (data, textStatus, jQxhr) {
            console.log('response: ' + JSON.stringify(data));
            if (data && !data.isError) {
                localStorage.setItem("sessionId", data.data);
                localStorage.setItem("username", username);
                toastr.success(data.message, `Login [${username}] success!`);
                refreshLoggedUserInfo();
            } else {
                toastr.error(data.message, 'Login error');
            }
        },
        error: function (jqXhr, textStatus, errorThrown) {
            console.log('response: ' + JSON.stringify(errorThrown));
            toastr.error(JSON.stringify(errorThrown), 'Login error');
        },
    });
}

function logout() {
    const username = localStorage.getItem("username");
    const sessionId = localStorage.getItem("sessionId");
    if (!username || !sessionId) {
        console.error('username | sessionId empty');
        return;
    }

    $.ajax({
        url: "/users/logout",
        type: "POST",
        dataType: "json",                 // expected format for response
        contentType: "application/x-www-form-urlencoded; charset=utf-8",
        data: {username: username, sessionId: sessionId},
        complete: function () {
            console.log('logout request complete');
        },
        success: function (data, textStatus, jQxhr) {
            console.log('response: ' + JSON.stringify(data));
            if (data && !data.isError) {
                localStorage.setItem("sessionId", "");
                localStorage.setItem("username", "");
                toastr.success(data.message, `Logout [${username}] success!`);
                refreshLoggedUserInfo();
            } else {
                toastr.error(data.message, 'Logout error');
                if (data && data.message && data.message.includes("session not found")) {
                    localStorage.setItem("sessionId", "");
                    localStorage.setItem("username", "");
                    refreshLoggedUserInfo();
                }
            }
        },
        error: function (jqXhr, textStatus, errorThrown) {
            console.log('response: ' + JSON.stringify(errorThrown));
            toastr.error(JSON.stringify(errorThrown), 'Logout error');
        },
    });
}

function refreshLoggedUserInfo() {
    const username = localStorage.getItem("username");
    const sessionId = localStorage.getItem("sessionId");
    if (!username || !sessionId) {
        $('#loginForm').css("display", "block");
        $('#loggedUserInfo').css("display", "none");
        $('#usernameInfo').text("-> " + username);

        $('#navbar-register-item').css("display", "block");
        $('#navbar-spends-item').css("display", "none");
    } else {
        $('#loginForm').css("display", "none");
        $('#loggedUserInfo').css("display", "block");
        $('#usernameInfo').text("-> " + username);

        $('#navbar-register-item').css("display", "none");
        $('#navbar-spends-item').css("display", "block");
    }
}

// window.onload = function () {
//     console.log('iSpend sidebar script loaded ...');
//     refreshLoggedUserInfo();
// };
