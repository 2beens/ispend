function registerPost() {
    // $.ajax({
    //     type: 'POST',
    //     url: '/users',
    //     contentType: "application/json; charset=utf-8",
    //     dataType: 'json',
    //     data: JSON.stringify({ username: '2beens', password: 'dummy' }),
    //     success: function (data, textStatus, jQxhr) {
    //         console.log('reply: ' + data);
    //     },
    //     error: function (jqXhr, textStatus, errorThrown) {
    //         console.error(errorThrown);
    //     }
    // });

    $.ajax({
        url: "/users",
        type: "POST",
        dataType: "json",                 // expected format for response
        contentType: "application/x-www-form-urlencoded; charset=utf-8",  // send as JSON
        data: { username: "John", password: "Boston" },
        complete: function() {
            console.log('completed');
        },
        success: function(data, textStatus, jQxhr) {
            console.log('response: ' + JSON.stringify(data));
        },
        error: function(jqXhr, textStatus, errorThrown) {
            console.log('response: ' + JSON.stringify(data));
        },
    });
}

(function () {
    console.log('iSpend register script loaded ...');
})();