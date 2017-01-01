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

var Collection = function(identifier) {
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

var checks = new Collection('id');
var nodes = new Collection('name');

var agents = new Array();
agents.get = function(name) {
    return agents.find(function(agent) {
        if (agent.name === name) {
            return true;
        }

        return false;
    })
};

var listChecks = Vue.component('list-checks', {
    data: function() {
        return {
            checks: checks
        };
    },

    methods: {
        deleteCheck: function(button, id) {
            button.disabled = true;

            Vue.http.delete('/api/checks/' + id);
        },

        editCheck: function(id) {
            router.push('/check/edit/' + id);
        }
    },

    template: '#template-checks'
});

var listNodes = Vue.component('list-nodes', {
    data: function() {
        return {
            nodes: nodes
        };
    },

    template: '#template-nodes'
});

var init = g.waitGroup(function() {
    g.live();

    const app = new Vue({
        el: '#app',
        router: router
    });
});

console.log(init);

Vue.http.get('/api/agents').then(function(response) {
    init.add(1);
    response.body.forEach(function(check) {
        agents.push(check);
        init.done();
    });
});

Vue.http.get('/api/checks').then(function(response) {
    init.add(1);
    response.body.forEach(function(check) {
        checks.upsert(check);
        init.done();
    });
});

/**
 * Keep live updates from service.
 */
g.live = function() {
    var socket;

    /**
     * Open the websocket connection.
     */
    var open = function() {
        if (g.isEncrypted()) {
            socket = new WebSocket('wss://' + g.getHost() + '/api/live');
        } else {
            socket = new WebSocket('ws://' + g.getHost() + '/api/live');
        }

        socket.onclose = onclose;
        socket.onmessage = onmessage;
    };

    /**
     * onmessage callback.
     * @param {!MessageEvent} event
     */
    var onmessage = function(event) {
        var data = JSON.parse(event.data);

        switch (data.type) {
            case 'nodeinfo':
                nodes.log(data);
                break;
            case 'checkresult':
                break;
            case 'check':
                checks.log(data);
                break;
            default:
                console.log(data);
        }
    };

    /**
     * Use as onclose callback from websocket, will try to reconnect after
     * two and a half second.
     * @param {!CloseEvent} event
     */
    var onclose = function(event) {
        setTimeout(function() {
            open();
        }, 2500);
    };

    // Open the connection right away.
    open();
};

var editCheck = Vue.component('edit-check', {
    data: function() {
        return {
            title: 'Add check',
            agents: agents,
            check: {
                arguments: {},
                agent: 'http',
                id: '',
                expressions: []
            },
            results: {results:{}},
        };
    },

    created: function() {
        // fetch the data when the view is created and the data is
        // already being observed
        this.fetchData()
    },

    watch: {
        '$route': 'fetchData'
    },

    methods: {
        addExpression: function() {
            this.check.expressions.push('');
        },

        removeExpression: function(index) {
            this.check.expressions.splice(index, 1);
        },

        fetchData: function() {
            var check = checks.get(this.$route.params.id);

            if (check != undefined) {
                this.title = "Edit " + this.$route.params.id;
                this.check = check;
            }
        },

        testCheck: function() {
            this.$http.post("/api/test", this.check).then(function(response) {
                this.results = response.body;
            });
        },

        addCheck: function() {
            this.$http.post("/api/checks", this.check).then(function(response) {
                router.push('/checks');
            });
        }
    },

    computed: {
        arguments: function () {
            var agentId = this.check.agent;
            var agent = agents.get(agentId);

            console.log(agentId);
            console.log(agent);

            return agent.arguments;
        }
    },

    template: '#template-edit-check'
});

const router = new VueRouter({
    routes: [
        { path: '/', component: { template: '<h1>Hello, world.</h1>' } },
        { path: '/overview', component: { template: '#template-overview' } },
        { path: '/gansoi', component: listNodes },
        { path: '/checks', component: listChecks },
        { path: '/check/edit/:id', component: editCheck }
    ]
});
