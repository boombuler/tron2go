var Tron = window.Tron || {};

Tron.Player = function() {
    var playerId;
    var playerName;


    var _allowLocalStorage = function() {
        try {
            return 'localStorage' in window && window['localStorage'] !== null;
        } catch (e) {
            return false;
        }
    };

    return {
        loadConfig: function() {
            if (_allowLocalStorage()) {
                playerName = localStorage.getItem("Nickname")
            }
        },

        getId: function() {
            return playerId;
        },
        setId: function(id) {
            playerId = id;
        },

        getName: function() {
            return playerName;
        },
        setName: function(name) {
            playerName = name;
            if (_allowLocalStorage()) {
                localStorage.setItem("Nickname", playerName)
            }
        }
    };
}();


Tron.PlayerList = function() {
    return {
        update: function(players) {
            players.sort(function(a, b) { return b.Score - a.Score; })

            var playersDiv = $('#players');
            playersDiv.empty();

            var list = $('<ul>');
            $.each(players, function(idx, player) {
                var listItem = $('<li>').css('color', PLAYER_COLORS[player.Id]);
                if (player.Id == Tron.Player.getId()) {
                    listItem.addClass('player-myself');
                }

                $('<span>').addClass('player-name').text(player.Name).appendTo(listItem);
                $('<span>').addClass('player-score').text(player.Score).appendTo(listItem);

                listItem.appendTo(list);
            });
            list.appendTo(playersDiv);
        }
    };
}();


Tron.Client = function() {
    var conn;


    var serverMessageHandler = {
        "draw.blocks": function(data) {
            Tron.ArenaCanvas.drawBlocks(data.Blocks);
        },
        "draw.gamestate": function(data) {
            Tron.PlayerList.update(data.Players);

            Tron.ArenaCanvas.clear();
            Tron.ArenaCanvas.drawBlocks(_translateBoardString(data.Board));
        },
        "set.identity": function(data) {
            Tron.Player.setId(data.Id);
        },
        "draw.scoreboard": function(data) {
            Tron.PlayerList.update(data.Players);
        }
    };

    var _translateBoardString = function(str) {
        var flds = str.split(",");
        var result = new Array();
        for(var x = 0; x < FIELD_WIDTH; x++) {
            for(var y = 0; y < FIELD_HEIGHT; y++) {
                var spid = flds[(x * FIELD_HEIGHT) + y]
                if (spid != '') {
                    var pid = parseInt(spid);
                    result.push({X: x, Y:y, PlayerId: pid});
                }
            }
        }
        return result;
    };

    var _sendSetNameCommand = function(name) {
        conn.send(JSON.stringify({ "Cmd" : "set.name", "Name": name }));
    };

    var _sendMoveCommand = function(direction) {
        conn.send(JSON.stringify({ "Cmd": "move." + direction }));
    };

    return {
        connect: function(url, roomId) {
            conn = new WebSocket(url +'?'+ roomId);
            conn.binaryType = 'arraybuffer';
            conn.onclose = function(evt) {
                setError('Disconnected');
            }
            conn.onerror = function(evt) {
                setError('ERROR: ' + evt);
            }
            conn.onopen = function(evt) {
                $('.game-screen').show();
                onResize();

                _sendSetNameCommand(Tron.Player.getName());
            }
            conn.onmessage = function(evt) {
                var data = JSON.parse(decodeServerMsg(evt.data));
                if (serverMessageHandler[data.Event]) {
                    serverMessageHandler[data.Event](data);
                }
            }

            $(document).bind('move.right', function() { _sendMoveCommand('right'); });
            $(document).bind('move.left', function() { _sendMoveCommand('left'); });
            $(document).bind('move.up', function() { _sendMoveCommand('up'); });
            $(document).bind('move.down', function() { _sendMoveCommand('down'); });
        }
    };
}();


Tron.Input = function() {
    var _handleTouch = function(ev) {
        var x = ev.originalEvent.touches[0].pageX;
        var y = ev.originalEvent.touches[0].pageY;

        var arena = $('#arena');
        var arenaOffset = arena.offset();
        var centerX = arenaOffset.left + (arena.prop('width') / 2.0);
        var centerY = arenaOffset.top + (arena.prop('height') / 2.0);

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
    };

    return {
        init: function() {
            $(document).bind('keydown.right', function(){ $(document).trigger('move.right'); });
            $(document).bind('keydown.left', function(){ $(document).trigger('move.left'); });
            $(document).bind('keydown.up', function(){ $(document).trigger('move.up'); });
            $(document).bind('keydown.down', function(){ $(document).trigger('move.down'); });

            $(document).bind('touchstart', _handleTouch);
        }
    };
}();


Tron.ArenaCanvas = function() {
    var canvas;
    var ctx;

    var playerThickness = 1;

    var boardCols;
    var boardRows;
    var board = [];


    var _repaint = function() {
        var b = board;
        Tron.ArenaCanvas.clear();
        Tron.ArenaCanvas.drawBlocks(b)
    };

    return {
        init: function(el, cols, rows) {
            boardCols = cols;
            boardRows = rows;

            canvas = $(el);
            if ((canvas.length < 1) || !canvas[0].getContext) {
               return false;
            }

            ctx = canvas[0].getContext('2d');
            Tron.ArenaCanvas.clear();

            return true;
        },

        onResize: function(maxWidth, maxHeight) {
            playerThickness = Math.floor(Math.min(maxWidth / boardCols, maxHeight / boardRows));
            if (playerThickness < 1) {
                playerThickness = 1;
            }

            canvas.prop('width', boardCols * playerThickness);
            canvas.prop('height', boardRows * playerThickness);

            _repaint();
        },

        clear: function() {
            board = [];
            if (ctx) {
                ctx.clearRect(0, 0, canvas.prop('width'), canvas.prop('height'));
            }
        },

        drawBlocks: function(blocks) {
            $.each(blocks, function(idx, block) {
                if (ctx) {
                    ctx.fillStyle = PLAYER_COLORS[block.PlayerId];
                    ctx.fillRect(block.X * playerThickness,
                        block.Y * playerThickness,
                        playerThickness,
                        playerThickness);
                }

                board.push(block);
            });
        }
    };
}();


function onResize() {
    var gameScreen = $('.game-screen');
    var gameContent = $('.game-content');
    var playersDiv = $('#players');

    var maxCanvasWidth = gameScreen.width() - playersDiv.outerWidth() - 16;
    var maxCanvasHeight = gameScreen.height() - 16;
    Tron.ArenaCanvas.onResize(maxCanvasWidth, maxCanvasHeight);

    playersDiv.css('height', $('#arena').prop('height'));

    gameContent.css({
        top: Math.max(((gameScreen.height() / 2) - (gameContent.outerHeight() / 2)), 5),
        left: Math.max(((gameScreen.width() / 2) - (gameContent.outerWidth() / 2)), 5)
    });
}

function setError(msg) {
    $('body').html('<div class="error-dlg-container"><div class="error-dlg">'+msg+'</div></div>')
}

function queryName() {
    if (!Tron.Player.getName()) {
        Tron.Player.setName(prompt("Name: "));
    }

    Tron.Client.connect(WEBSOCKET_URL, 0);
}


$(function() {
    $('.game-screen').hide();

    if (!window['WebSocket']) {
       setError('Your browser does not support WebSockets');
       return;
    }

    if (!Tron.ArenaCanvas.init('#arena', FIELD_WIDTH, FIELD_HEIGHT)) {
       setError('Your browser does not support canvas elements');
       return;
    }

    Tron.Player.loadConfig();
    Tron.Input.init();

    $(window).on('resize', onResize);

    queryName();
});
