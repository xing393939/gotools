let callerMap = {};
let calleeMap = {};
let app = new Vue({
    el: '#app',
    data() {
        return {
            props: {
                label: 'label',
                children: 'children',
                isLeaf: 'leaf'
            },
            data: [],
            filterText: '',
            callerMap: {},
            calleeMap: {},
        };
    },
    created() {
        axios.get(`./data.gv`).then(response => {
            this.setData(response.data);
        }).catch(function (error) {
            console.log(error);
        });
    },
    watch: {
        filterText(val) {
            this.$refs.tree.filter(val);
        }
    },
    methods: {
        filterNode(value, data) {
            if (!value) return true;
            return data.label.indexOf(value) !== -1;
        },
        setTextarea() {
            let self = this;
            this.$prompt('输入dotgraph内容', {
                confirmButtonText: '确定',
                cancelButtonText: '取消',
                inputType: 'textarea',
            }).then((v) => {
                this.setData(v.value);
                this.setRoots();
            })
        },
        setData(v) {
            let lines = v.split("\n");
            callerMap = {};
            calleeMap = {};
            lines.forEach((v2) => {
                let edge = v2.match(/"(.*)" "(.*)"/);
                if (!edge || edge.length != 3) {
                    return
                }
                if (callerMap[edge[1]]) {
                    callerMap[edge[1]].push(edge[2])
                } else {
                    callerMap[edge[1]] = [edge[2]];
                }
                if (calleeMap[edge[2]]) {
                    calleeMap[edge[2]].push(edge[1])
                } else {
                    calleeMap[edge[2]] = [edge[1]];
                }
            });
            this.setRoots();
        },
        setRoots() {
            let arr = [];
            for (let nodeName in callerMap) {
                if (Object.keys(calleeMap).indexOf(nodeName) === -1) {
                    arr.push(nodeName);
                }
            }
            arr.sort();
            let roots = [];
            arr.forEach((nodeName) => {
                roots.push({label: nodeName});
            });
            this.data = roots;
        },
        loadNode(node, resolve) {
            if (node.level === 0) {
                return resolve(this.data);
            }
            let nodeName = node.data.label;
            let nodes = [];
            let temps = [];
            if (!callerMap[nodeName]) {
                resolve(nodes);
                return
            }
            callerMap[nodeName].forEach((v2) => {
                if (temps.indexOf(v2) > -1) {
                    return
                }
                temps.push(v2);
                let leaf = true;
                if (callerMap[v2]) {
                    leaf = false;
                }
                let node = {label: v2, leaf: leaf};
                if (nodes.indexOf(node) === -1) {
                    nodes.push(node)
                }
            });
            resolve(nodes);
        }
    }
});