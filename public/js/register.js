function registerPost() {
    var username = $('#username').val();
    var password = $('#password').val();
    if (!username || !password) {
        console.error('username | password empty');
        return;
    }

    $.ajax({
        url: "/users",
        type: "POST",
        dataType: "json",                 // expected format for response
        contentType: "application/x-www-form-urlencoded; charset=utf-8",  // send as JSON
        data: {username: username, password: password},
        complete: function () {
            console.log('completed');
        },
        success: function (data, textStatus, jQxhr) {
            console.log('response: ' + JSON.stringify(data));
        },
        error: function (jqXhr, textStatus, errorThrown) {
            console.log('response: ' + JSON.stringify(data));
        },
    });
}

(function () {
    console.log('iSpend register script loaded ...');
})();