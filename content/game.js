var playerThickness;
var canvas;

var board = [];

function drawBlocks(blocks) {
	$.each(blocks, function(idx, block) {
		canvas.drawRect({
			fillStyle: PLAYER_COLORS[block.PlayerId],
	  		x: (1+block.X) * playerThickness, 
	  		y: (1+block.Y) * playerThickness,
	  		width: playerThickness,
	  		height: playerThickness,
	  		fromCenter: false
		})
		board.push(block);
	});
}

function drawBoard() {
	canvas.clearCanvas().drawRect({
		strokeStyle: BORDER_COLOR,
	    strokeWidth: 2*playerThickness,
  		fillStyle: BACKGROUND_COLOR,
  		x: 0, y: 0,
  		width: (2+FIELD_WIDTH) * playerThickness,
  		height: (2+FIELD_HEIGHT) * playerThickness,
  		fromCenter: false
	});
	var b = board;
	board = []
	drawBlocks(b)
}

function onResize() {
	playerThickness = Math.floor(Math.min($(document).width() / (2+FIELD_WIDTH), $(document).height() / (2+FIELD_HEIGHT)));
	canvas.attr('width', (2+FIELD_WIDTH) * playerThickness);
	canvas.attr('height', (2+FIELD_HEIGHT) * playerThickness);

	drawBoard();
}

function handleTouch(ev) {
	var x = ev.originalEvent.touches[0].clientX;
	var y = ev.originalEvent.touches[0].clientY;
	
	var centerX = (canvas.attr('width') / 2.0);
	var centerY = (canvas.attr('height') / 2.0);

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

}

function bindInput() {
	$(document).bind('keydown.right', function(){ $(document).trigger('move.right'); });
	$(document).bind('keydown.left', function(){ $(document).trigger('move.left'); });
	$(document).bind('keydown.up', function(){ $(document).trigger('move.up'); });
	$(document).bind('keydown.down', function(){ $(document).trigger('move.down'); });

	canvas.bind('touchstart', handleTouch)
}

function setError(msg) {
	$('body').html('<div><b>'+msg+'</b></div>')
}

function connect() {
	if (!window['WebSocket'])
		return false;

	conn = new WebSocket(WEBSOCKET_URL);
    conn.onclose = function(evt) {
        setError('Disconnected.');
    }
    conn.onerror = function(evt) {
    	setError('ERROR: '+ evt);
    }
    conn.onmessage = function(evt) {
    	data = JSON.parse(evt.data)
    	$(document).trigger(data.Event, data);
    }

    $(document).bind('move.right', function() { conn.send('move.right'); });
    $(document).bind('move.left', function() { conn.send('move.left'); });
    $(document).bind('move.up', function() { conn.send('move.up'); });
    $(document).bind('move.down', function() { conn.send('move.down'); });

    return true;
}
	
$(function() {
    if (!connect()) {
       setError('Your browser does not support WebSockets.');
       return;
    }

	canvas = $('#canvas');

	$(window).on('resize', onResize);
	$(document).bind('draw.blocks', function(ev, data) { drawBlocks(data.Blocks) });

	bindInput();
	onResize();
});