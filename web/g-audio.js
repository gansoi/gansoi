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

    self.preloadInflight = {};

    /**
     * Preload a specific audio snippet.
     * @param {!string} id The id to play. This must match an id of an
     *                     audio-element
     */
    self.preload = function(id) {
        var element = document.getElementById('audio-' + id);
        if (element == null) {
            console.error(id + ' not found');
            return;
        }

        // If the src property is a "blob:"-url it makes no sense to preload.
        // We exit in silence, since many users may try to reload the same
        // audio snippet.
        if (element.src.startsWith('blob:')) {
            return;
        }

        // Multiple preloads could easily be called before the first one
        // succeed. We keep track of what's in flight, so we can easily cancel
        // repeated requests.
        if (self.preloadInflight[element.src]) {
            return;
        }
        self.preloadInflight[element.src] = true;

        // We use XMLHttpRequest to retrieve the binary data as a Blob and then
        // use URL's createObjectURL to get a unique URL for the ressource. We
        // then write that URL back to the src tag.
        var xhr = new XMLHttpRequest();
        xhr.onload = function() {
            element.src = URL.createObjectURL(xhr.response);

            self.preloadInflight[element.src] = false;
        };
        xhr.onerror = function() {
            self.preloadInflight[element.src] = false;
        };

        xhr.open('GET', element.src);
        xhr.responseType = 'blob';
        xhr.send();
    };

    return self;
};
