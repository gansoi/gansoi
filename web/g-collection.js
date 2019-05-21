/**
 * A collection of objects from Gansoi.
 * @constructor
 * @param {!string} identifier THe name of the member used to identify an
 *        object.
 */
g.Collection = function(identifier) {
    var self = this;

    self.dataset = new timeline.DataSet([], {
        fieldId: identifier
    });

    // This is for easy Vue.js iteration.
    self.data = self.dataset._data;

    self.insert = function(obj) {
        Vue.set(self.data, obj[identifier], obj);

        self.dataset.add(obj);
    };

    self.deleteId = function(id) {
        Vue.delete(self.data, id);

        self.dataset.remove(id);
    };

    self.upsert = function(obj) {
        Vue.set(self.data, obj[identifier], obj);

        self.dataset.update(obj);
    };

    self.get = function(id) {
        return self.dataset.get(id);
    };

    self.query = function(field, value) {
        return self.dataset.get({
            filter: function (item) {
                return (item[field] == value);
            }
        });
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
