/**
 * A collection of objects from Gansoi.
 * @constructor
 * @param {!string} identifier THe name of the member used to identify an
 *        object.
 */
g.Collection = function(identifier) {
    var self = this;

    self.data = new Array();

    self.insert = function(obj) {
        self.data.push(obj);
    };

    self.deleteId = function(id) {
        self.data = self.data.filter(function(obj) {
            return obj[identifier] !== id;
        });
    };

    self.upsert = function(obj) {
        var index = self.data.findIndex(function(element) {
            if (element[identifier] === obj[identifier]) {
                return true;
            }

            return false;
        });

        if (index >= 0) {
            // We must use splice, if we simply replace the element Vue will
            // never notice.
            // https://vuejs.org/v2/guide/list.html#Caveats
            self.data.splice(index, 1, obj);
        } else {
            self.insert(obj);
        }
    };

    self.get = function(id) {
        var ret = self.data.find(function(element) {
            if (element[identifier] === id) {
                return true;
            }
        });

        return ret;
    };

    self.log = function(log) {
        switch (log.command) {
            case 'delete':
                self.deleteId(log.data[identifier]);
                break;
            case 'save':
                self.upsert(log.data);
                break;
            default:
                console.dir(log);
        }
    };
};
