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
    }
};
