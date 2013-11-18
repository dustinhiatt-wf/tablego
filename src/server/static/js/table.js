var ws;
var tableId = '5c645d52fbb74d64a44535c942dfb431';
var numRows = 3000;

function setModal(visible) {
    var modal = $('#dialog-modal');
    if (visible) {
        modal.dialog('open');
    } else {
        modal.dialog('close');
    }
}

function updateCells(cells) {
    for (var i=0; i < cells.length; i++) {
        updateCell(cells[i]);
    }
}

function updateCell(cell) {
    var table = document.getElementById('table');
    console.log(cell)
    table.rows[cell[0]].cells[cell[1]].innerText = cell[2];
}

function saveValue(row, column, value) {
    sendMessage({
        'row': row,
        'column': column,
        'value': value,
        'operation': 'editCellValue',
        'table_id': tableId
    }, false);
}

function sendMessage(msg, displayModal) {
    if (displayModal == true) {
        msg.message_id = Math.floor((Math.random() * 100) + 1);
        messageId = msg.message_id;
        setModal(true);
    }
    ws.send(JSON.stringify(msg));
}

$(document).ready(function (){
    //presentation
    var table = $('#table');

    var colGroup = document.createElement('colgroup');

    for (var i=0; i < 26; i++) {
        var col = document.createElement('col');
        colGroup.appendChild(col);
    }
    table.append(colGroup);

    var tbody = $('tbody');

    for (var i=0; i < numRows; i++) {
        var tr = document.createElement('tr');
        for (var j=0; j < 26; j++) {
            var td = document.createElement('td');
            td.setAttribute('contenteditable', 'true');
            tr.appendChild(td);
        }
        tbody.append(tr)
    }

    var previousValue = null;
    var startElement = null;
    var startIndex = null;
    var stopIndex = null;

    $(this).keydown(function (ev) {
        if (!startElement) {
            return;
        }

        $('#table .ui-selected').removeClass('ui-selected');
        startElement.focus();
    });

    $('#dialog-modal').dialog({
        height: 140,
        modal: true,
        autoOpen: false
    });

    setModal(true);

    var tds = $('td');
    tds.focus(function () {
        var td = $(this);
        previousValue = td.text();
    });
    tds.blur(function () {
        var td = $(this);
        var column = td.index();
        var row = td.parent().index();
        var val = td.text();
        if (val !== previousValue) {
            saveValue(row, column, val);
        }
    });

    table.selectable(
        {
            filter:'td',
            autoRefresh: false,
            start: function (event, ui) {
                if (startElement) {
                    startElement.attr('contenteditable', 'false');
                }
            },
            stop: function (event, ui) {
                var selecteds = $('.ui-selected');
                startElement = selecteds.first();
                var lastElement = selecteds.last();
                startIndex = {
                    'column': startElement.index(),
                    'row': startElement.parent().index()
                };
                stopIndex = {
                    'column': lastElement.index() + 1,
                    'row': lastElement.parent().index() + 1
                };
                startElement.attr('contenteditable', 'true');
            }
        }
    );


    ws = new WebSocket("ws://localhost:8123/ws");
    ws.onopen = function() {
        ws.send(JSON.stringify({
            'operation': 'register',
            'table_id': tableId,
            'start_row': 0,
            'stop_row': numRows + 1,
            'start_column': 0,
            'stop_column': 27
        }));
    };
    ws.onmessage = function(e) {
        var result = jQuery.parseJSON(e.data);

        if (result.hasOwnProperty('operation'))
        {
            switch (result.operation) {
                case 'cellUpdated':
                    updateCell(result.values);
                    break;
                case 'registered':
                    updateCells(result.values);
                    setModal(false);
                    break;
            }
        }

        if (result.hasOwnProperty('message_id')) {
            if (result.message_id && result.message_id == messageId) {
                messageId = null;
                setModal(false);
            }
        }
    };
    ws.onclose = function() {
        console.log('closed');
    };
});
