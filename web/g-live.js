/**
 * Keep live updates from service.
 * @constructor
 */
g.live = function() {
    var socket;
    var reconnect = 0;
    var audio = new g.audio();

    /**
     * Open the websocket connection.
     */
    var open = function() {
        if (g.isEncrypted()) {
            socket = new WebSocket('wss://' + g.getHost() + '/api/live');
        } else {
            socket = new WebSocket('ws://' + g.getHost() + '/api/live');
        }

        socket.onopen = onopen;
        socket.onclose = onclose;
        socket.onmessage = onmessage;
    };

    /**
     * onmessage callback.
     * @param {!MessageEvent} event
     */
    var onmessage = function(event) {
        var data = JSON.parse(event.data);

        if (subscriptions.hasOwnProperty(data.type)) {
            subscriptions[data.type].forEach(function(collection) {
                collection.log(data);
            });
        }
    };

    var onopen = function(event) {
        if (reconnect > 0) {
            audio.play('connected');
        }

        reconnect = 0;
    };

    /**
     * Use as onclose callback from websocket, will try to reconnect after
     * two and a half second.
     * @param {!CloseEvent} event
     */
    var onclose = function(event) {
        // Play a audible message at every tenth reconnect attempt.
        if (reconnect % 10 == 0) {
            audio.play('connection-lost');
        }

        reconnect++;

        setTimeout(function() {
            open();
        }, 2500);
    };

    // Open the connection right away.
    open();

    /**
     * @type {Object.<string, g.Collection>}
     */
    var subscriptions = {};

    /**
     * Subscribe a collection.
     * @param {!string} id The type id to subscribe to.
     * @param {!g.Collection} collection The collection to update.
     */
    this.subscribe = function(id, collection) {
        if (subscriptions.hasOwnProperty(id)) {
            subscriptions[id].push(collection);
        } else {
            subscriptions[id] = [collection];
        }
    };

    return this;
};
