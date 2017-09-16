/**
 * This is a collection of wrappers around a key/value storage.
 */
g.storage = {
    /**
     * Get a value from storage.
     * @param {!string} key The key to set.
     * @param {*} defaultValue An optional value to return if the key was not
     *                         found.
     */
    get: function(key, defaultValue) {
        var value = localStorage.getItem(key);
        if (null === value) {
            value = defaultValue;
        } else {
            try {
                value = JSON.parse(value)
            } catch (e) {
                console.log("Error parsing '" + value + "' from localstorage key '" + key + "', removing id");
                g.storage.unset(key);

                value = defaultValue;
            }
        }

        return value;
    },

    /**
     * Set a value to storage.
     * @param {!string} key The key to retrieve.
     * @param {*} value The value to set. The value will be JSON-encoded.
     */
    set: function(key, value) {
        localStorage.setItem(key, JSON.stringify(value));
    },

    /**
     * Remove a value from storage.
     * @param {!string} key The key to remove.
     */
    unset: function(key) {
        localStorage.removeItem(key);
    }
};
