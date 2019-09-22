function login() {
    var username = $('#form_username').val();
    var password = $('#form_password').val();
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
            if (data && !data.is_error) {
                toastr.success(data.message, `Login [${username}] success!`);
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

(function () {
    console.log('iSpend sidebar script loaded ...');
})();