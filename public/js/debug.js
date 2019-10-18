function getUsers() {
    $.ajax({
        url: "/users",
        type: "GET",
        dataType: "json",                 // expected format for response
        headers: {
            'X-Ispend-SessionID': getSessionID()
        },
        data: {username: getUsername()},
        complete: function () {
            console.log('completed');
        },
        success: function (data, textStatus, jQxhr) {
            console.log('response: ' + JSON.stringify(data));
            $('#users-list').text(JSON.stringify(data, null, '\t'));
        },
        error: function (jqXhr, textStatus, errorThrown) {
            console.log('response: ' + JSON.stringify(errorThrown));
        },
    });
}

(function () {
    console.log('iSpend debug script loaded ...');
    getUsers();
})();