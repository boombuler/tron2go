var playerThickness;
var canvas;
var canvasContext;

var board = [];
var Nickname = undefined;

function drawBlocks(blocks) {
    $.each(blocks, function(idx, block) {
        if (canvasContext) {
            canvasContext.fillStyle = PLAYER_COLORS[block.PlayerId];
            canvasContext.fillRect(block.X * playerThickness,
                block.Y * playerThickness,
                playerThickness,
                playerThickness);
        }

        board.push(block);
    });
}

function drawBoard() {
    if (canvasContext) {
        canvasContext.clearRect(0, 0, canvas.prop('width'), canvas.prop('height'));
    }
    var b = board;
    board = []
    drawBlocks(b)
}

function updatePlayerList(players) {
    players.sort(function(a, b) { return b.Score - a.Score; })

    var playersDiv = $('#players');
    playersDiv.empty();

    var list = $('<ul>');
    $.each(players, function(idx, player) {
        var listItem = $('<li>').css('color', PLAYER_COLORS[player.Id]);
        $('<span>').addClass('player-name').text(player.Name).appendTo(listItem);
        $('<span>').addClass('player-score').text(player.Score).appendTo(listItem);
        listItem.appendTo(list);
    });
    list.appendTo(playersDiv);
}

function onResize() {
    var gameScreen = $('.game-screen');
    var gameContent = $('.game-content');
    var playersDiv = $('#players');

    var maxCanvasWidth = gameScreen.width() - playersDiv.outerWidth() - 16;
    var maxCanvasHeight = gameScreen.height() - 16;
    playerThickness = Math.floor(Math.min(maxCanvasWidth / FIELD_WIDTH, maxCanvasHeight / FIELD_HEIGHT));
    if (playerThickness < 1) {
        playerThickness = 1;
    }

    canvas.prop('width', FIELD_WIDTH * playerThickness);
    canvas.prop('height', FIELD_HEIGHT * playerThickness);

    playersDiv.css('height', canvas.prop('height'));

    gameContent.css({
        top: Math.max(((gameScreen.height() / 2) - (gameContent.outerHeight() / 2)), 5),
        left: Math.max(((gameScreen.width() / 2) - (gameContent.outerWidth() / 2)), 5)
    });

    drawBoard();
}

function handleTouch(ev) {
    var x = ev.originalEvent.touches[0].clientX;
    var y = ev.originalEvent.touches[0].clientY;

    var centerX = (canvas.prop('width') / 2.0);
    var centerY = (canvas.prop('height') / 2.0);

    var deltaLeft  = x < centerX ? centerX - x : 0;
    var deltaUp    = y < centerY ? centerY - y : 0;
    var deltaRight = x > centerX ? x - centerX : 0;
    var deltaDown  = y > centerY ? y - centerY : 0;

    var maxDelta = Math.max(deltaLeft, deltaUp, deltaDown, deltaRight);
    if (deltaLeft == maxDelta)
        $(document).trigger('move.left');
    else if (deltaRight == maxDelta)
        $(document).trigger('move.right');
    else if (deltaDown == maxDelta)
        $(document).trigger('move.down');
    else if (deltaUp == maxDelta)
        $(document).trigger('move.up');
    ev.preventDefault()
}

function bindInput() {
    $(document).bind('keydown.right', function(){ $(document).trigger('move.right'); });
    $(document).bind('keydown.left', function(){ $(document).trigger('move.left'); });
    $(document).bind('keydown.up', function(){ $(document).trigger('move.up'); });
    $(document).bind('keydown.down', function(){ $(document).trigger('move.down'); });

    canvas.bind('touchstart', handleTouch)
}

function setError(msg) {
    $('body').html('<div class="error-dlg-container"><div class="error-dlg">'+msg+'</div></div>')
}

function connect() {
    conn = new WebSocket(WEBSOCKET_URL);
    conn.onclose = function(evt) {
        setError('Disconnected.');
    }
    conn.onerror = function(evt) {
        setError('ERROR: '+ evt);
    }
    conn.onopen = function(evt) {
        $('.game-screen').show()
        onResize();
        conn.send(JSON.stringify({'Cmd' : 'set.name', 'Name': Nickname}))
    }
    conn.onmessage = function(evt) {
        data = JSON.parse(evt.data)
        $(document).trigger(data.Event, data);
    }

    var sendCommand = function (cmd) {
        conn.send(JSON.stringify({'Cmd': cmd}))
    }

    $(document).bind('move.right', function() { sendCommand('move.right'); });
    $(document).bind('move.left', function() { sendCommand('move.left'); });
    $(document).bind('move.up', function() { sendCommand('move.up'); });
    $(document).bind('move.down', function() { sendCommand('move.down'); });
}

function allowLocalStorage() {
    try {
        return 'localStorage' in window && window['localStorage'] !== null;
    } catch (e) {
        return false;
    }
}

function queryName() {
    if (allowLocalStorage()) {
        Nickname = localStorage.getItem("Nickname")
    }
    if (!Nickname) {
        Nickname = prompt("Name: ");
        if (allowLocalStorage()) {
            localStorage.setItem("Nickname", Nickname)
        }
    }
    connect();
}


$(function() {
    $('.game-screen').hide()
    if (!window['WebSocket']) {
       setError('Your browser does not support WebSockets.');
       return;
    }

    canvas = $('#arena');
    if ((canvas.length < 1) || !canvas[0].getContext) {
       setError('Your browser does not support canvas elements.');
       return;
    }

    canvasContext = canvas[0].getContext('2d');

    $(window).on('resize', onResize);
    $(document).bind('draw.blocks', function(ev, data) { drawBlocks(data.Blocks) });
    $(document).bind('draw.gamestate', function(ev, data) {
        board = [];
        drawBoard();
        drawBlocks(data.Blocks);
        updatePlayerList(data.Players);
    });
    $(document).bind('set.identity', function(ev, data) {
        // alert(JSON.stringify(data)); // Server tells us who we are...
    });
    bindInput();
    queryName();
});
