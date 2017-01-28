/**
 * Keep live updates from service.
 */
g.live = function() {
    var socket;

    /**
     * Open the websocket connection.
     */
    var open = function() {
        if (g.isEncrypted()) {
            socket = new WebSocket('wss://' + g.getHost() + '/api/live');
        } else {
            socket = new WebSocket('ws://' + g.getHost() + '/api/live');
        }

        socket.onclose = onclose;
        socket.onmessage = onmessage;
    };

    /**
     * onmessage callback.
     * @param {!MessageEvent} event
     */
    var onmessage = function(event) {
        var data = JSON.parse(event.data);

        switch (data.type) {
            case 'nodeinfo':
                nodes.log(data);
                break;
            case 'checkresult':
                checkresults.log(data);
                break;
            case 'check':
                checks.log(data);
                break;
            case 'evaluation':
                evaluations.log(data);
                break;
            default:
                console.log(data);
        }
    };

    /**
     * Use as onclose callback from websocket, will try to reconnect after
     * two and a half second.
     * @param {!CloseEvent} event
     */
    var onclose = function(event) {
        setTimeout(function() {
            open();
        }, 2500);
    };

    // Open the connection right away.
    open();
};
