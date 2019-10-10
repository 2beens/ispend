window.onload = function () {
    refreshLoggedUserInfo();

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