/**
 * Play audio in the browser.
 * @constructor
 */
g.audio = function() {
    var self = this;

    var player = null;

    /**
     * Play a specific audio snippet. If a snippet is already playing, it will
     * be stopped.
     * @param {!string} id The id to play. This must match an id of an
     *                     audio-element
     */
    self.play = function(id) {
        player = document.getElementById('audio-' + id);
        if (player == null) {
            console.error(id + ' not found');
            return;
        }

        self.stop();
        player.play();
    };

    /**
     * Stop all playback.
     */
    self.stop = function() {
        if (player == null) {
            return;
        }

        // Stop is just pause'n'rewind.
        player.pause();
        player.currentTime = 0;
    };

    return self;
};
