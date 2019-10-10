function registerPost() {
    var email = $('#email').val();
    var username = $('#username').val();
    var password = $('#password').val();
    if (!username || !password) {
        toastr.error('username | password empty', 'Register error');
        return;
    }

    $.ajax({
        url: "/users",
        type: "POST",
        dataType: "json",                 // expected format for response
        contentType: "application/x-www-form-urlencoded; charset=utf-8",  // send as JSON
        data: {email: email, username: username, password: password},
        complete: function () {
            console.log('register request complete');
            $('#username').val('');
            $('#password').val('');
        },
        success: function (data, textStatus, jQxhr) {
            console.log('response: ' + JSON.stringify(data));
            if (data && !data.isError) {
                toastr.success(data.message, `Register [${username}] success!`);
            } else {
                toastr.error(data.message, 'Register error');
            }
        },
        error: function (jqXhr, textStatus, errorThrown) {
            console.log('response: ' + JSON.stringify(errorThrown));
            toastr.error(JSON.stringify(errorThrown), 'Register error');
        },
    });
}

(function () {
    console.log('iSpend register script loaded ...');
    toastr.success('Page loaded ....');
})();