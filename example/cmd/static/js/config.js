

let reqtpl = `
<div class="button-group fill">
<button class="send" @click=sendReq(item)>{{item.id}}</button>
<button class="button" @click="editReq(cls,item)"><i class="fa fa-pencil-square-o"
    aria-hidden="true"></i></button>
<button class="button" @click="deleteReq(cls, item)"><i class="fa fa-trash-o"
    aria-hidden="true"></i></button>
</div>
`


Vue.component('reqbtn', {
    props: ['item', 'cls'],
    methods: {
        sendReq(item) {
            // console.log("send", item)
            config.client.Send(item.req, JSON.stringify(item.params)).then(()=>{
                // this.$notify({
                //     title: '成功',
                //     message: '发送成功',
                //     position: 'top-left',
                //     duration: 500,
                //     type: 'success'
                //   });
            }).catch((e)=>{
                this.$notify({
                    title: '失败',
                    message: '发送失败:' + e,
                    position: 'top-left',
                    duration: 2000,
                    // offset: 100,
                    type: 'error'
                  });
            })
        },

        editReq(cls, item) {
            if (cls === 1) {
                cls = 'common'
            }
            // console.log('editReq', cls, item)
            reqhd.ShowByName(cls, item)
        },

        deleteReq(cls, item) {
            config.deleteReq(cls, item)
        },
    },
    template: reqtpl
})






let config = new Vue({
    el: "#config",
    data: {
        servername: '',
        servers: [],
        trees: [],
        subtrees: [],
        mixtree:null,
        reqclass: '',
        classList: [],
        bizreqs: [],
        commreqs: [],
        reqs: null,
        pbmsgs: null,
        client: new RequestAgent,
        File: {},
        state: {},
        loginLimit:10,
        runMode: 'stress',
        account:"",

        // file:new FileAgent("./cache/server.json", "Server"),
    },
    methods: {
        onRunModeChange(val){
            // console.log(val, this.runMode)
        },
        onAccountChange(){
            let tag = this.servername + "." + "account"
            this.File["Server"].Set(tag, this.account)
            tabs.setServer()
        },
        onServerNameChange(){
            // console.log(this.servername)
            this.account = this.GetServerConfig().account
            tabs.setServer()
        },
        onClickEditApp(id) {
            dialog.Show("应用配置", this.File["App"].Gets())
        },
        onClickEditServer() {
            dialog.Show("服务器配置", this.File["Server"].Gets()[this.servername])
        },
        onClickReload() {
            this.client.Reload().then(()=>{

            })
        },
        onClickLogin(){
            stressDialog.RunTree(this.trees[0])
        },
        onClickClearLog() {
            logbox.clearAll()
        },
        onClickCreateReq(cls) {
            if (cls === 'biz') {
                cls = this.reqclass
            }
            reqhd.Show(cls)
        },
        onClickStopRun() {
            bar.changeLevel('error')
            this.client.Stop()
        },

        onClickSelectZone() {
            zoneDialog.Show()
        },
        onSelectClass() {
            this.bizreqs = this.reqs[this.reqclass]
            this.File['State'].Set('reqClass', this.reqclass)
        },

        onClickStartStress() {
            stressDialog.Show(this.mixtree)
        },


        createReq(msgname, classname, reqname, json) {
            let data = {
                id: reqname,
                req: msgname,
                params: json,
            }
            // data[reqname] = 


            let key = classname + '.' + reqname
            // console.log(key, data)
            this.File["Req"].Set(key, data).then(() => {
                this.refreshReqs()
                this.changeBizReq(this.reqclass)
            })
        },

        deleteReq(cls, item) {
            if (cls === 1) {
                cls = 'common'
            } else {
                cls = this.reqclass
            }
            let key = cls + '.' + item.id
            // console.log('deleteReq', key, item)
            this.File["Req"].Del(key).then(() => {
                this.refreshReqs()
            })
        },

        refreshReqs() {
            this.reqs = this.File["Req"].Gets()
            this.commreqs = this.reqs['common']
            let keys = Object.keys(this.reqs)
            if (keys.length !== 0) {
                this.classList = keys.slice(1, keys.length)
            }

            reqhd.SetReqs(this.reqs)
            // console.log('refreshReqs', this.commreqs)
        },

        changeBizReq(name) {
            if (name === '') {
                // 初始化
                if (this.classList.length !== 0) {
                    name = this.classList[0]
                }
            }
            this.bizreqs = this.reqs[name]
            this.reqclass = name
        },

        GetServerConfig(name) {
            if (name == undefined){
                name = this.servername
            }
            return this.File["Server"].Gets()[name]
        },

        onJsonModified(name, json) {
            // console.log('onJsonModified', name)
            if (name === '应用配置') {
                this.File["App"].Sets(json)
                return
            }

            if (name === '服务器配置') {
                this.account = json.account
                this.File["Server"].Set(this.servername, json)
                return
            }
        },

        restoreStat(){
            this.state = this.File['State'].Gets()
            if (this.state.hasOwnProperty('lastServer')){
                this.servername = this.state['lastServer']
            }

            if (this.state.hasOwnProperty('reqClass')){
                this.reqclass = this.state['reqClass']
            }

        },
        saveLastloginServer(){
            this.File['State'].Set('lastServer', this.servername)
        },
        saveLogFilter(name){
            this.File['State'].Set('logFilter.'+name, true)
            this.state = this.File['State'].Gets()
        },
        setSelectedTree(id){
            this.File['Btree'].Set('data.selectedTree', id)
        },

        reloadTree(){
            let agent = new FileAgent("./conf/robot.b3", "Btree")
            agent.then(res =>{
                this.File['Btree'] = res
                this.onTreeLoaded()
            })
        },

        saveZone(zoneId){
            let data = this.GetServerConfig()
            data.zone = zoneId
            this.File["Server"].Set(this.servername, data)

        },
        skipShowLog(log){
            if (this.state.hasOwnProperty('logFilter')){
                let filter = this.state.logFilter
                return filter[log.meta.name] === true
            }
            return false
        },

        getSubtreeWeight() {
            let arr = []
            for (const tr of this.subtrees){
                
                let obj = {
                    'weight': Number(tr.properties.weight),
                    'id':tr.id,
                }
                arr.push(obj)
            }
            return arr
        },
        getAppConfigJs(){
            let limitConf = {
                'GamerLoginC2S': Number(this.loginLimit),
            }
            return {
                'rate_limit':limitConf,
            }
        },
        onTreeLoaded(){
            trees = []
            subtrees = []
            tag = this.File["App"].Gets().show_tags
            for (const tr of this.File["Btree"].Gets().data.trees){
                data = tr.properties
                if (data.hasOwnProperty('subtree')){
                    if (data.hasOwnProperty('weight')== false){
                        data['weight'] = 3
                        
                    }
                    // console.log(data, data.hasOwnProperty('weight'))
                    subtrees.push(tr)
                }

                if (data.hasOwnProperty('mixtree')){
                    this.mixtree = tr
                }else if (data.tag ==1 || tag=="" || data.tag ==tag){
                    trees.push(tr)
                }
                
            }
            this.subtrees = subtrees
            this.trees = trees
            // console.log('tress', this.trees[0])
            treejson = this.File["Btree"].Gets()
        }
    },

    mounted() {
        let files = [
            new FileAgent("./conf/server.json", "Server"),
            new FileAgent("./conf/state.json", "State"),
            new FileAgent("./conf/app.json", "App"),
            new FileAgent("./conf/req_short_cut.json", "Req"),
            new FileAgent("./conf/robot.b3", "Btree"),
        ]

        this.client.Gets().then((data) => {
            // console.log("msglist", data)
            this.pbmsgs = data
            reqhd.SetMsgs(data)
        }).catch(e=>{
            console.log("failed",e)
        })

        // console.log("start load all ")
        let ct= Promise.all(files).then(fas => {
            for (const fa of fas) {
                this.File[fa.name] = fa
            }

            this.servers = this.File["Server"].Gets()
            this.servername = Object.keys(this.servers)[0]
            this.onTreeLoaded()
            // console.log('load tree', treejson)
            
            // let d = {}
            // src = this.File['Req'].Gets()
            // let total = 0
            // for (let e of Object.keys(this.File['Req'].Gets())){
            //     let dd = {}
            //     d[e]=dd
            //     for (let ee of Object.keys(this.File['Req'].Gets()[e])){
                    
            //         let item = src[e][ee]
            //         total += 1
            //         // console.log(this.File['Req'].Gets()[e][ee])
            //         dd[ee]={}
            //         this.client.GetInfo(item.req).then((res)=>{
            //             // console.log(item, res)
            //             total -= 1
            //             item.params = res
            //             dd[ee]= {
            //                 id: item.id,
            //                 req: item.req,
            //                 params: item.params,
            //             }

            //             if (total <= 0 ){
            //                 console.log(d, dd)
            //                 this.File['Req'].Sets(d)
            //             }
            //         })
            //     }
                
            // }
            // console.log(d)
            
            
            this.restoreStat()
            this.refreshReqs()
            this.changeBizReq(this.reqclass)
            
            this.account = this.GetServerConfig().account
            tabs.setServer()
            //console.log(this.GetServerConfig())

            // console.log("all file ok")
        }).catch(e => {
            console.log(e)
        })

        ct.finally()
    }
})


