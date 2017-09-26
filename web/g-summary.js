/**
 * Can hold a summary of current Gansoi state.
 * @param {!g.audio} audio The audio controller to use for
 *                                   announcements.
 * @constructor
 */
g.Summary = function(audio) {
    var self = this;

    var lastMessage = '';

    self.checks = '-';
    self.states = {
        'unknown': '-',
        'up': '-',
        'down': '-'
    };

    self.log = function(log) {
        self.checks = log.data.checks;
        self.states.unknown = log.data.states.unknown;
        self.states.up = log.data.states.up;
        self.states.down = log.data.states.down;

        var message = self.message();
        if (message != self.lastMessage) {
            audio.play(message);

            self.lastMessage = message;
        }
    };

    self.message = function() {
        if (self.checks === self.states.up) {
            return 'checks-up';
        }

        if (self.states.down === 1) {
            return 'check-down';
        }

        if (self.states.down > 1) {
            return 'checks-down';
        }
    };

    return this;
}
