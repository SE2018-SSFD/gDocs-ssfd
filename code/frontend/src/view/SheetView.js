import React from 'react';
import {Link, withRouter} from "react-router-dom";
import {commitSheet, getSheet, getSheetCkpt, testWS} from "../api/sheetService";
import {HTTP_URL, MSG_WORDS, WS_URL} from "../api/common";
import {Button, Card, Col, Divider, Drawer, Layout, message, Row, Tooltip} from "antd";
import {
    CheckCircleOutlined,
    EditOutlined,
    FolderOutlined,
    LeftOutlined,
    MenuOutlined,
    StarOutlined
} from "@ant-design/icons";
import {UserAvatar} from "../components/UserAvatar";
import {history} from "../route/history";

const {Header} = Layout
const luckysheet = window.luckysheet;
let socket;
let locking_row = -1, locking_col = -1, locked_row = -1, locked_col = -1;

class SheetView extends React.Component {


    constructor(props) {
        super(props);
        this.fid = 0;
        this.columns = 0;
        this.content = [];
        this.checkpoint_num = 0;
        this.checkpoint = [];
        this.url = "";
        this.state = {
            name: "",
            logVisible: false,
            ckptVisible: false,
        }
    }

    componentDidMount() {
        const query = this.props.location.search;
        const arr = query.split('&');
        this.fid = parseInt(arr[0].substr(4));

        const token = JSON.parse(localStorage.getItem("token"));

        const get_data = {
            token: token,
            fid: this.fid,
        }
        const callback = (data) => {
            console.log(data);
            if (data.success === true) {
                this.checkpoint_num = data.data.checkpoint_num;
                let checkpoint = [];
                for (let i = this.checkpoint_num; i >= 1; i--) {
                    checkpoint.push(i);
                }
                this.checkpoint = checkpoint;
            } else {
                console.log(MSG_WORDS[data.msg]);
            }
        }

        getSheet(HTTP_URL + 'getsheet', get_data, callback)

        testWS(this.fid, this.connectWS);
    }

    openLogDrawer = () => {
        this.setState({
            logVisible: true,
        })
    }

    closeLogDrawer = () => {
        this.setState({
            logVisible: false,
        })
    }

    openCkptDrawer = () => {
        this.setState({
            ckptVisible: true,
        })
    }

    closeCkptDrawer = () => {
        this.setState({
            ckptVisible: false,
        })
    }

    handleRecoverLog = (lid) =>
    {
        console.log(lid);
    }

    handleSave = () =>{
        const callback = (data) =>{
            console.log(data)
            let msg_word = MSG_WORDS[data.msg];
            if (data.success === true) {
                message.success(msg_word).then(() => {
                });
            } else {
                message.error(msg_word).then(() => {
                });
            }

        }
        commitSheet(this.fid,callback)
    }

    handleRecoverCkpt = (cid) =>{
        console.log(cid);
        const callback = (data) =>{
            let msg_word = MSG_WORDS[data.msg];
            if(data.success===true)
            {
                this.cid = data.data.cid;
                this.timestamp = data.data.timestamp;
                this.rows = data.data.rows;
                this.columns = data.data.columns;
                this.content = data.data.content;
                const username = JSON.parse(localStorage.getItem("username"));
                let j = 0, k = 0;
                let celldata = [];
                for (let i = 0; i < this.content.length; i++) {
                    j = Math.floor(i / this.columns);
                    k = i % this.columns;
                    if (this.content[i] !== "") {
                        celldata.push({
                                "r": j,
                                "c": k,
                                "v": this.content[i],
                            }
                        )
                    }
                }
                luckysheet.destroy();
                luckysheet.create({
                    container: "luckysheet",
                    title: this.name,
                    lang: 'zh',
                    gridKey: this.fid,
                    data: [{
                        "name": "Sheet1",
                        color: "",
                        "status": "1",
                        "order": "0",
                        "celldata": celldata,
                        "config": {},
                        "index": 0
                    }],
                    showtoolbar: false,
                    showinfobar: false,
                    showsheetbar: false,
                    userInfo: username,
                    userMenuItem: [
                        {url: "/", "icon": '<i class="fa fa-folder" aria-hidden="true"></i>', "name": "我的表格"},
                    ],
                    myFolderUrl: "/",
                    functionButton:
                        '<button id="log_button" style="padding:3px 6px;font-size: 12px;margin-right: 10px;">Log</button>' +
                        '<button id="ckpt_button" style="padding:3px 6px;font-size: 12px;margin-right: 10px;">CKPT</button>',
                    hook: {
                        // 进入单元格编辑模式之前触发。
                        cellEditBefore: (range) => {
                            console.info('cellEditBefore', range[0]);
                            const row = range[0].row_focus;
                            const col = range[0].column_focus;
                            const username = JSON.parse(localStorage.getItem("username"))
                            const cellLocks = JSON.parse(localStorage.getItem("cellLocks"));
                            for (let i = 0; i < cellLocks.length; i++) {
                                if (row === cellLocks[i].Row && col === cellLocks[i].Col) {
                                    if (username !== cellLocks.Username) {
                                        console.log("other is writing");
                                        message.error(cellLocks.Username + "正在输入，请稍等再点击")
                                    }
                                }
                            }
                            locking_row = row;
                            locking_col = col;
                            const data = {
                                msgType: "acquire",
                                body: {
                                    row: row,
                                    col: col,
                                }
                            }
                            socket.send(JSON.stringify(data))
                        },
                        cellUpdateBefore: function (r, c, value, isRefresh) {
                            console.info('cellUpdateBefore', r, c, value, isRefresh)
                            if (r === locked_row && c === locked_col) {
                                const data = {
                                    msgType: "modify",
                                    body: {
                                        row: r,
                                        col: c,
                                        content: value
                                    }
                                }
                                console.log(data)
                                socket.send(JSON.stringify(data))
                            }
                        },

                        //更新这个单元格后触发
                        // cellUpdated: (r, c, oldValue, newValue, isRefresh) => {
                        //     console.info('cellUpdated', r, c, oldValue, newValue, isRefresh);
                        //     if (r === locked_row && c === locked_col) {
                        //         const data = {
                        //             msgType: "release",
                        //             body: {
                        //                 row: r,
                        //                 col: c,
                        //             }
                        //         }
                        //         console.log(data);
                        //         locked_col = -1;
                        //         locked_row = -1;
                        //         socket.send(JSON.stringify(data))
                        //     }
                        //     let content;
                        //     if (newValue.ct.t === "inlineStr") {
                        //         content = newValue.ct.s[0].v;
                        //     } else if (newValue.ct.t === "n") {
                        //         content = newValue.v.toString();
                        //     } else if (newValue.ct.t === "g") {
                        //         content = newValue.v;
                        //     }
                        //     if (content.indexOf(" 正在输入 ") === -1) {
                        //         const data1 = {
                        //             msgType: "modify",
                        //             body: {
                        //                 row: r,
                        //                 col: c,
                        //                 content: content
                        //             }
                        //         }
                        //         console.log(data1)
                        //         socket.send(JSON.stringify(data1))
                        //         const data2 = {
                        //             msgType: "release",
                        //             body: {
                        //                 row: r,
                        //                 col: c,
                        //             }
                        //         }
                        //         socket.send(JSON.stringify(data2))
                        //     }
                        // },
                    }
                });
                message.success(msg_word).then(() => {
                });
            }
            else{
                message.error(msg_word).then(() => {
                });
            }
        }
        const token = JSON.parse(localStorage.getItem("token"));
        const fid = this.fid;
        const data = {
            token:token,
            fid:fid,
            cid:cid,
        }
        getSheetCkpt(data,callback)
    }

    connectWS = (data) => {
        const token = JSON.parse(localStorage.getItem("token"));
        const username = JSON.parse(localStorage.getItem("username"));

        this.url = WS_URL + 'sheetws?token=' + token + "&fid=" + this.fid;
        if (data.success === false) {
            this.url = data.data;
        } else {
            console.log(MSG_WORDS[data.msg]);
        }

        socket = new WebSocket(this.url);
        socket.addEventListener('open', (event) => {
            console.log('WebSocket open: ', event);
        });
        socket.addEventListener('message', (event) => {
            // console.log('WebSocket message: ', event);
            let data = JSON.parse(event.data);
            console.log(data);
            switch (data.msgType) {
                case "onConn": {
                    let cellLocks = data.body.cellLocks;
                    if (cellLocks === null) {
                        cellLocks = [];
                    }
                    localStorage.setItem("cellLocks", JSON.stringify(cellLocks));
                    this.columns = data.body.columns;
                    this.content = data.body.content;
                    this.name = data.body.name;
                    this.setState({
                        name: data.body.name,
                    })
                    let j = 0, k = 0;
                    let celldata = [];
                    for (let i = 0; i < this.content.length; i++) {
                        j = Math.floor(i / this.columns);
                        k = i % this.columns;
                        if (this.content[i] !== "") {
                            celldata.push({
                                    "r": j,
                                    "c": k,
                                    "v": this.content[i],
                                }
                            )
                        }
                    }

                    for (let i = 0; i < cellLocks.length; i++) {
                        celldata.push({
                            "r": cellLocks[i].Row,
                            "c": cellLocks[i].Col,
                            "v": cellLocks[i].Username + " 正在编辑"
                        })
                    }

                    luckysheet.create({
                        container: "luckysheet",
                        title: this.name,
                        lang: 'zh',
                        gridKey: this.fid,
                        data: [{
                            "name": "Sheet1",
                            color: "",
                            "status": "1",
                            "order": "0",
                            "celldata": celldata,
                            "config": {},
                            "index": 0
                        }],
                        showtoolbar: false,
                        showinfobar: false,
                        showsheetbar: false,
                        userInfo: username,
                        userMenuItem: [
                            {url: "/", "icon": '<i class="fa fa-folder" aria-hidden="true"></i>', "name": "我的表格"},
                        ],
                        myFolderUrl: "/",
                        functionButton:
                            '<button id="log_button" style="padding:3px 6px;font-size: 12px;margin-right: 10px;">Log</button>' +
                            '<button id="ckpt_button" style="padding:3px 6px;font-size: 12px;margin-right: 10px;">CKPT</button>',
                        hook: {
                            // 进入单元格编辑模式之前触发。
                            cellEditBefore: (range) => {
                                console.info('cellEditBefore', range[0]);
                                const row = range[0].row_focus;
                                const col = range[0].column_focus;
                                const username = JSON.parse(localStorage.getItem("username"))
                                const cellLocks = JSON.parse(localStorage.getItem("cellLocks"));
                                for (let i = 0; i < cellLocks.length; i++) {
                                    if (row === cellLocks[i].Row && col === cellLocks[i].Col) {
                                        if (username !== cellLocks.Username) {
                                            console.log("other is writing");
                                            message.error(cellLocks.Username + "正在输入，请稍等再点击")
                                        }
                                    }
                                }
                                locking_row = row;
                                locking_col = col;
                                const data = {
                                    msgType: "acquire",
                                    body: {
                                        row: row,
                                        col: col,
                                    }
                                }
                                socket.send(JSON.stringify(data))
                            },
                            cellUpdateBefore: function (r, c, value, isRefresh) {
                                console.info('cellUpdateBefore', r, c, value, isRefresh)
                                if (r === locked_row && c === locked_col) {
                                    const data = {
                                        msgType: "modify",
                                        body: {
                                            row: r,
                                            col: c,
                                            content: value
                                        }
                                    }
                                    console.log(data)
                                    socket.send(JSON.stringify(data))
                                }
                            },

                            //更新这个单元格后触发
                            // cellUpdated: (r, c, oldValue, newValue, isRefresh) => {
                            //     console.info('cellUpdated', r, c, oldValue, newValue, isRefresh);
                            //     if (r === locked_row && c === locked_col) {
                            //         const data = {
                            //             msgType: "release",
                            //             body: {
                            //                 row: r,
                            //                 col: c,
                            //             }
                            //         }
                            //         console.log(data);
                            //         locked_col = -1;
                            //         locked_row = -1;
                            //         socket.send(JSON.stringify(data))
                            //     }
                            //     let content;
                            //     if (newValue.ct.t === "inlineStr") {
                            //         content = newValue.ct.s[0].v;
                            //     } else if (newValue.ct.t === "n") {
                            //         content = newValue.v.toString();
                            //     } else if (newValue.ct.t === "g") {
                            //         content = newValue.v;
                            //     }
                            //     if (content.indexOf(" 正在输入 ") === -1) {
                            //         const data1 = {
                            //             msgType: "modify",
                            //             body: {
                            //                 row: r,
                            //                 col: c,
                            //                 content: content
                            //             }
                            //         }
                            //         console.log(data1)
                            //         socket.send(JSON.stringify(data1))
                            //         const data2 = {
                            //             msgType: "release",
                            //             body: {
                            //                 row: r,
                            //                 col: c,
                            //             }
                            //         }
                            //         socket.send(JSON.stringify(data2))
                            //     }
                            // },
                        }
                    })
                    break;
                }
                case "acquire": {
                    const row = data.body.row;
                    const col = data.body.col;
                    const username = data.body.username;
                    const me = JSON.parse(localStorage.getItem("username"));
                    const cellLocks = JSON.parse(localStorage.getItem("cellLocks"));
                    cellLocks.push({
                        Row: row,
                        Col: col,
                        Username: username,
                    })
                    localStorage.setItem("cellLocks", JSON.stringify(cellLocks));
                    if (locking_col === col && locking_row === row) {
                        if (me === username) {
                            console.log("acquired");
                            locked_row = row;
                            locked_col = col;
                        } else {
                            console.log("not acquired");
                            message.error(username + " 正在编辑，请稍等再点击");
                            locked_row = -1;
                            locked_col = -1;
                            luckysheet.setCellValue(row, col, username + " 正在编辑 ");
                        }
                    } else {
                        luckysheet.setCellValue(row, col, username + " 正在编辑 ");
                    }
                    break;
                }
                case "modify": {
                    let row = data.body.row;
                    let col = data.body.col;
                    let content = data.body.content;
                    if (row === locked_row && col === locked_col) {
                        console.log("modify_success");
                        const data = {
                            msgType: "release",
                            body: {
                                row: row,
                                col: col,
                            }
                        }
                        socket.send(JSON.stringify(data))
                    } else {
                        console.log("others modify this");
                        luckysheet.setCellValue(row, col, content);
                    }
                    break;
                }
                case "release": {
                    //TODO
                    let row = data.body.row;
                    let col = data.body.col;
                    const cellLocks = JSON.parse(localStorage.getItem("cellLocks"));
                    const username = JSON.parse(localStorage.getItem("username"));
                    for (let i = 0; i < cellLocks.length; i++) {
                        if (cellLocks[i].Row === row && cellLocks[i].Col === col) {
                            if (cellLocks[i].Username === username) {
                                console.log("release success");
                                locked_col = -1;
                                locked_row = -1;
                            } else {
                                console.log("others have release the lock");
                            }
                        }
                    }
                    break;
                }
                default: {
                    break;
                }
            }
        });
        socket.addEventListener('error', function (event) {
            console.log('WebSocket error: ', event);
        });
    }

    render() {
        const luckyCss = {
            margin: '0px',
            padding: '0px',
            position: 'absolute',
            width: '100%',
            height: '93%',
            left: '0px',
            top: '60px',
        }
        const {name} = this.state;

        let logContent = this.checkpoint.map(
            (item) =>
                <Card hoverable style={{width: 240}} title={"编辑记录" + item.toString()}>
                    <Button onClick={()=>this.handleRecoverLog(item)}>恢复到此处</Button>
                </Card>);
        let ckptContent = this.checkpoint.map(
            (item) =>
                <Card hoverable style={{width: 240}}
                      title={"恢复点" + item.toString()}>
                    <Button onClick={()=>this.handleRecoverCkpt(item)}>恢复到此处</Button>
                </Card>
        );

        return (
            <div>
                <Header className="site-layout-sub-header-background" style={{padding: 0}}>
                    <Row justify={"center"}>
                        <Col span={2}>
                            <Link to={{
                                pathname: '/',
                            }}
                            >
                                <LeftOutlined/>
                                {/*<Image src={docs} alt={'docs'} height={50} width={50} preview={false}/>*/}
                            </Link>
                        </Col>
                        <Col span={1}>
                            <StarOutlined/>
                        </Col>
                        <Col span={1}>
                            <FolderOutlined/>
                        </Col>
                        <Col span={1}>
                            <CheckCircleOutlined/>
                        </Col>
                        <Col span={4}>
                            <h1>{name}</h1>
                        </Col>
                        <Col span={1} offset={6}>
                            <MenuOutlined/>
                        </Col>
                        <Col span={1}>
                            <EditOutlined/>
                        </Col>
                        <Divider type={"vertical"}/>
                        <Col span={1}>
                            <Tooltip title="复制url给你的好友吧">
                                <Button type={'primary'}> 分享</Button>
                            </Tooltip>
                        </Col>
                        <Col span={1}>
                            <Button type={'primary'} onClick={this.openLogDrawer}>Log</Button>
                        </Col>
                        <Col span={1}>
                            <Button type={'primary'} onClick={this.openCkptDrawer}>Ckpt</Button>
                        </Col>
                        <Col span={1}>
                            <Button type={'primary'} onClick={this.handleSave}>Save</Button>
                        </Col>
                        <Divider type={"vertical"}/>
                        <Col span={1}>
                            <UserAvatar/>
                        </Col>
                    </Row>
                </Header>
                <div
                    id="luckysheet"
                    style={luckyCss}
                />
                {/*Log*/}
                <Drawer
                    title="Log"
                    onClose={this.closeLogDrawer}
                    visible={this.state.logVisible}
                    bodyStyle={{paddingBottom: 80}}
                    width={320}
                    footer={
                        <div
                            style={{
                                textAlign: 'right',
                            }}
                        >
                            <Button onClick={this.closeLogDrawer} style={{marginRight: 8}}>
                                返回
                            </Button>
                        </div>
                    }
                >
                    {logContent}
                </Drawer>
                <Drawer
                    title="Checkpoint"
                    onClose={this.closeCkptDrawer}
                    visible={this.state.ckptVisible}
                    bodyStyle={{paddingBottom: 80}}
                    width={320}
                    footer={
                        <div
                            style={{
                                textAlign: 'right',
                            }}
                        >
                            <Button onClick={this.closeCkptDrawer} style={{marginRight: 8}}>
                                返回
                            </Button>
                        </div>
                    }
                >
                    {ckptContent}
                </Drawer>
            </div>
        )
    }
}

export default withRouter(SheetView);
