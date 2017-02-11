/**
 * Global gansoi-scope.
 * @const
 */
var g = {
    /**
     * Get scheme for current site.
     * @return {string} Scheme. Examples: "http" or "https"
     */
    getScheme: function() {
        return window.location.protocol.replace(/:/g, '');
    },

    /**
     * Check if the HTTP connection is encrypted.
     * @return {boolean} True if connection is encrypted, false otherwise
     */
    isEncrypted: function() {
        return (g.getScheme() === 'https');
    },

    /**
     * Get the hostname part of website URL. This includes port (!) if port is
     * non-standard.
     * @return {string} Hostname of website. Examples: "localhost:9000"
     */
    getHost: function() {
        return window.location.host;
    },

    /**
     * Get a textual representation of a duration in milliseconds.
     * @param {!int} duration A duration in milliseconds.
     * @return {!string} A textual representation of duration.
     */
    durationToText: function(duration) {
        const millisecond = 1;

        const second = 1000 * millisecond;
        const minute = 60 * second;
        const hour = 60 * minute;
        const day = 24 * hour;

        var days = Math.floor(duration / day);
        var hours = Math.floor(duration / hour);
        var minutes = Math.floor(duration / minute);
        var seconds = Math.round(duration / second);

        if (duration < (2 * second)) {
            return duration + 'ms';
        }

        if (duration < minute) {
            return seconds + 's';
        }

        if (duration < hour) {
            return minutes + 'm' + seconds % 60 + 's';
        }

        if (duration < day) {
            return hours + 'h' + minutes % 60 + 'm';
        }

        return days + 'd' + hours % 24 + 'h';
    }
};
