/**
 * Will wait for a number of done()'s before calling cb.
 * @param {!func} cb
 * @constructor
 */
g.waitGroup = function(cb) {
    var count = 0;

    /**
     * Add delta units of work to finish before calling cb
     * @param {!number} delta The number to add.
     */
    this.add = function(delta) {
        count += delta;
    };

    /**
     * Mark one job as done.
     */
    this.done = function() {
        count--;
        if (count === 0) {
            cb();
        }
    };

    return this;
};
