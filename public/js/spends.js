function getSpendKinds(user, callback) {
    $.ajax({
        url: '/spending/kind/' + user.username,
        type: 'GET',
        dataType: 'json',                 // expected format for response
        contentType: 'application/x-www-form-urlencoded; charset=utf-8',
        complete: function () {
            console.log('get spend kinds request complete');
        },
        success: function (data, textStatus, jQxhr) {
            console.log('response: ' + JSON.stringify(data));
            if (data && !data.isError) {
                toastr.success(data.message, `Get spend kinds [${user.username}] success!`);
                callback(data.data);
            } else {
                toastr.error(data.message, 'Get spend kinds error: ' + data.message);
                callback({});
            }
        },
        error: function (jqXhr, textStatus, errorThrown) {
            console.log('response: ' + JSON.stringify(errorThrown));
            toastr.error(JSON.stringify(errorThrown), 'Get spend kinds error');
            callback({});
        },
    });
}

function getSpends(user, callback) {
    $.ajax({
        url: '/spending/all/' + user.username,
        type: 'GET',
        dataType: 'json',                 // expected format for response
        contentType: 'application/x-www-form-urlencoded; charset=utf-8',
        headers: {
            'X-Ispend-SessionID': getSessionID()
        },
        complete: function () {
            console.log('get spends request complete');
        },
        success: function (data, textStatus, jQxhr) {
            console.log('response: ' + JSON.stringify(data));
            if (data && !data.isError) {
                toastr.success(data.message, `Get spends [${user.username}] success!`);
                callback(data.data);
            } else {
                toastr.error(data.message, 'Get spends error: ' + data.message);
                callback({});
            }
        },
        error: function (jqXhr, textStatus, errorThrown) {
            console.log('response: ' + JSON.stringify(errorThrown));
            toastr.error(JSON.stringify(errorThrown), 'Get spends error');
            callback({});
        },
    });
}

function deleteSpend(spendID, callback) {
    const user = getLoggedUser();
    if (!user.isLogged) {
        console.error('not logged in');
    }
    console.log('trying to delete: ' + spendID);

    $.ajax({
        url: `/spending/${user.username}/${spendID}`,
        type: "DELETE",
        dataType: "json",                 // expected format for response
        contentType: "application/x-www-form-urlencoded; charset=utf-8",  // send as JSON
        headers: {
            'X-Ispend-SessionID': getSessionID()
        },
        success: function (data, textStatus, jQxhr) {
            console.log('response: ' + JSON.stringify(data));
            if (data && !data.isError) {
                callback(spendID, true);
            } else {
                console.error('delete spending error: ' + data.message);
                callback(spendID, false);
            }
        },
        error: function (jqXhr, textStatus, errorThrown) {
            console.log('response: ' + JSON.stringify(errorThrown));
            callback(spendID, false);
        },
    });
}

function handleSpendingRemoved(spendId, isRemoved) {
    if (!isRemoved) {
        toastr.error('Spending not deleted!', 'Delete spending');
        return;
    }

    document.getElementById(`spending-row-${spendId}`).remove();
    toastr.success('Spending deleted!', 'Delete spending');
}

function postNewSpending(user, spending, callback) {
    $.ajax({
        url: "/spending",
        type: "POST",
        dataType: "json",                 // expected format for response
        contentType: "application/x-www-form-urlencoded; charset=utf-8",  // send as JSON
        headers: {
            'X-Ispend-SessionID': getSessionID()
        },
        data: {username: user.username, currency: spending.currency, amount: spending.amount, kind_id: spending.skId},
        complete: function () {
            console.log('new spending request complete');
        },
        success: function (data, textStatus, jQxhr) {
            console.log('response: ' + JSON.stringify(data));
            if (data && !data.isError) {
                callback(true, data.data);
            } else {
                console.error('add new spending error: ' + data.message);
                callback(false);
            }
        },
        error: function (jqXhr, textStatus, errorThrown) {
            console.log('response: ' + JSON.stringify(errorThrown));
            callback(false);
        },
    });
}

function addSpending() {
    const user = getLoggedUser();
    if (!user.isLogged) {
        console.error('not logged in');
    }

    const amount = $('#amount').val();
    const currency = $('#currency').val();
    const sk = document.getElementById('spend-kinds');
    const skId = sk.options[sk.selectedIndex].value;
    if (!amount || !currency || !skId) {
        toastr.error('Error, please check parameters.', 'Add spending');
        return;
    } else {
        toastr.success('OK', 'Add spending');
    }

    const spending = {amount: amount, currency: currency, skId: skId};
    postNewSpending(user, spending, function(success, spendId) {
        if (success) {
            toastr.success('Spending added!', 'Add new spending');
            const skName = sk.options[sk.selectedIndex].text;
            spending.id = spendId;
            addSpendKindToSpendsTable(spending, skName);
        } else {
            toastr.error('Spending was not added!', 'Add new spending');
        }
    });
}

function addSpendKindToSpendsTable(s, kindName) {
    const spendsTable = document.getElementById('spends-table');

    const tdAmount = document.createElement('td');
    tdAmount.appendChild(document.createTextNode(s.amount + ' ' + s.currency));
    const tdKind = document.createElement('td');
    tdKind.appendChild(document.createTextNode(kindName));

    const tdDelete = document.createElement('td');
    document.createElement("input");
    tdDelete.appendChild(htmlToElement(`<input type="button" onclick="deleteSpend('${s.id}', handleSpendingRemoved)" value="Delete">`));

    const newRow = document.createElement('tr');
    newRow.setAttribute('id', `spending-row-${s.id}`);
    newRow.appendChild(tdAmount);
    newRow.appendChild(tdKind);
    newRow.appendChild(tdDelete);

    spendsTable.appendChild(newRow);
}

(function () {
    console.log('iSpend spends script loaded ...');
    const user = getLoggedUser();
    if (!user.isLogged) {
        console.error('not logged in');
    }

    getSpendKinds(user, function(spendKinds) {
        console.log(JSON.stringify(spendKinds));
        const spendKindsDropdown = document.getElementById('spend-kinds');
        spendKinds.forEach(function(sk, i) {
            console.log('processing: ' + `<option value="${sk.id}">${sk.name}</option>`);
            const skOption = document.createElement('option');
            skOption.value = sk.id;
            skOption.text = sk.name;
            spendKindsDropdown.appendChild(skOption);
        });
    });

    getSpends(user, function(spends) {
        const spendsTable = document.getElementById('spends-table');
        while (spendsTable.firstChild) {
            spendsTable.removeChild(spendsTable.firstChild);
        }
        if (!spends) {
            return;
        }
        spends.forEach(function(s, i) {
            addSpendKindToSpendsTable(s, s.kind.name);
        });
        spendsTable.innerHTML = `<tr><th>Amount [currency]</th><th>Kind</th><th></th></tr>` + spendsTable.innerHTML;
    });

})();