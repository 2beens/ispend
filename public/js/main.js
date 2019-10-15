window.onload = function () {
    refreshLoggedUserInfo();

	checkSessionOK();

    const path = window.location.pathname;
    console.log('iSpend main script loaded: ' + path);
    
    $('.ispend-navbar-item').each(function(i, obj) {
	    obj.classList.remove('selected');
	});

    if (path == '/spends') {
		$('#navbar-spends-item').addClass('selected');
	} else if (path == '/contact') {
		$('#navbar-contact-item').addClass('selected');
	} else if (path == '/register') {
		$('#navbar-register-item').addClass('selected');
	} else if (path == '/debug') {	
		$('#navbar-debug-item').addClass('selected');
    } else {
    	$('#navbar-home-item').addClass('selected');
    }
};

function checkSessionOK() {
	const lastCheckDateStr = localStorage.getItem("session-check-timestamp");
	const now = new Date();
	if (lastCheckDateStr) {
		const lastCheck = Date.parse(lastCheckDateStr);
		const diffMs = now - lastCheck;
		const diffSeconds = diffMs / 1000;
		// check session after minute
		if (diffSeconds <= 60) {
			console.log(`skipping session check [diffMs = ${diffMs}]`);
			return;
		}
	}

	const username = localStorage.getItem("username");
	const sessionId = localStorage.getItem("sessionId");
	if (!username || !sessionId) {
		return;
	}

	console.log('about to check session ...');
	$.ajax({
		url: "/users/login/check",
		type: "POST",
		dataType: "json",                 // expected format for response
		contentType: "application/x-www-form-urlencoded; charset=utf-8",
		data: {username: username, sessionId: sessionId},
		success: function (data, textStatus, jQxhr) {
			console.log('check session response: ' + JSON.stringify(data));
			if (data && !data.isError) {
				localStorage.setItem("session-check-timestamp", now);
				if (data.message === "true") {
					// session OK
					console.log('session OK');
					return;
				}
				console.warn('session not OK');
				localStorage.setItem("sessionId", "");
				localStorage.setItem("username", "");
				toastr.info('Must login!!');
				refreshLoggedUserInfo();
			}
		},
		error: function (jqXhr, textStatus, errorThrown) {
			console.log('check session error, response: ' + JSON.stringify(errorThrown));
		},
	});
}
