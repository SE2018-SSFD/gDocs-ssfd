import React from 'react';
import {Link, withRouter} from "react-router-dom";
import {commitSheet, getSheet, getSheetCkpt, getSheetLog, rollbackSheet, testWS} from "../api/sheetService";
import {MSG_WORDS, GET_WS_URL} from "../api/common";
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
import {ColMap, RowMap} from "../utils";

const {Header} = Layout
const luckysheet = window.luckysheet;
let socket;
let locking_row = -1, locking_col = -1, locked_row = -1, locked_col = -1;

class SheetView extends React.Component {

    constructor(props) {
        super(props);
        this.fid = 0;
        this.url = "";
        this.checkpoint_now = 0;
        this.state = {
            checkpoint_num: 0,
            checkpoint: [],
            log: [],
            name: "",
            logVisible: false,
            ckptVisible: false,
        }
    }

    componentDidMount() {
        const query = this.props.location.search;
        const arr = query.split('&');
        const fid = parseInt(arr[0].substr(4));

        this.fid = fid;

        const callback = (data) => {
            if (data.success === true) {
                this.checkpoint_now = data.data.checkpoint_num;
                const checkPointBrief = data.data.checkPointBrief === null? []:data.data.checkPointBrief.reverse()
                this.setState({
                    checkpoint_num: data.data.checkpoint_num,
                    checkpoint: checkPointBrief
                })
            } else {
                console.log(MSG_WORDS[data.msg]);
            }
        }
        getSheet(fid, callback)

        testWS(fid, this.connectWS);
    }

    getSheetCallback = (data) => {
        if (data.success === true) {
            const checkPointBrief = data.data.checkPointBrief === null? []:data.data.checkPointBrief.reverse()
            this.setState({
                checkpoint_num: data.data.checkpoint_num,
                checkpoint: checkPointBrief
            })
        } else {
            console.log(MSG_WORDS[data.msg]);
        }
    }


    componentWillUnmount() {
        socket.close();
    }

    openLogDrawer = () => {
        const callback = (data) => {
            const msg_word = MSG_WORDS[data.msg];
            if (data.success === true) {
                const log = data.data;
                let log_new = [];
                for (let i = log.length - 1; i >= 0; i--) {
                    if (log[i].new !== log[i].old && log[i].lid !== 0) {
                        log_new.push(log[i]);
                    }
                }
                this.setState({
                    logVisible: true,
                    log: log_new
                })
                message.success(msg_word);
            } else {
                message.error(msg_word);
            }
        }
        getSheetLog(this.fid, this.checkpoint_now, callback)
    }

    closeLogDrawer = () => {
        this.setState({
            logVisible: false,
        })
    }

    openCkptDrawer = () => {
        getSheet(this.fid, this.getSheetCallback)
        this.setState({
            ckptVisible: true,
        })
    }

    closeCkptDrawer = () => {
        this.setState({
            ckptVisible: false,
        })
    }

    handleRecoverLog = (lid) => {
        console.log(lid);
    }

    handleSave = () => {
        const callback = (data) => {
            let msg_word = MSG_WORDS[data.msg];
            if (data.success === true) {
                getSheet(this.fid, this.getSheetCallback);
                message.success(msg_word).then(() => {
                });
            } else {
                message.error(msg_word).then(() => {
                });
            }

        }
        commitSheet(this.fid, callback)
    }

    handleRollback = () => {
        const callback = (data) => {
            console.log(data);
            let msg_word = MSG_WORDS[data.msg];
            if (data.success === true) {

                let checkpoint = this.state.checkpoint;
                let new_checkpoint = [];

                for (let i = 0; i < checkpoint.length; i++) {
                    if (checkpoint[i].cid <= this.checkpoint_now) {
                        new_checkpoint.push(checkpoint[i]);
                    }
                }
                this.setState({
                    checkpoint_num: this.checkpoint_now,
                    checkpoint: new_checkpoint,
                })

                message.success(msg_word).then(() => {
                });
            } else {
                message.error(msg_word).then(() => {
                });
            }
        }

        rollbackSheet(this.fid, this.checkpoint_now, callback)
    }

    handleRecoverCkpt = (cid) => {
        const callback = (data) => {
            let msg_word = MSG_WORDS[data.msg];
            if (data.success === true) {
                this.cid = data.data.cid;
                this.timestamp = data.data.timestamp;
                this.rows = data.data.rows;
                const columns = data.data.columns;
                this.checkpoint_now = cid;
                const content = data.data.content;
                const username = JSON.parse(localStorage.getItem("username"));
                this.setState({
                    ckptVisible: false
                })

                let j = 0, k = 0;
                let celldata = [];
                for (let i = 0; i < content.length; i++) {
                    j = Math.floor(i / columns);
                    k = i % columns;
                    if (content[i] !== "") {
                        celldata.push({
                                "r": j,
                                "c": k,
                                "v": content[i],
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
                    // functionButton:
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
                                socket.send(JSON.stringify(data))
                            }
                        },
                    }
                });
                message.success(msg_word).then(() => {
                });
            } else {
                message.error(msg_word).then(() => {
                });
            }
        }
        const token = JSON.parse(localStorage.getItem("token"));
        const fid = this.fid;
        const data = {
            token: token,
            fid: fid,
            cid: cid,
        }
        getSheetCkpt(data, callback)
    }

    connectWS = (data) => {
        const token = JSON.parse(localStorage.getItem("token"));
        const username = JSON.parse(localStorage.getItem("username"));

        if (data.success === false) {
            this.url = data.data;
            if (data.msg !== 19) {
                message.error(MSG_WORDS[data.msg])
            }
        } else {
            this.url = GET_WS_URL() + 'sheetws?token=' + token + "&fid=" + this.fid;
        }

        socket = new WebSocket(this.url);
        socket.addEventListener('open', (event) => {
            console.log('WebSocket open: ', event);
        });
        socket.addEventListener('message', (event) => {
            console.log('WebSocket message: ', event);
            let data = JSON.parse(event.data);
            switch (data.msgType) {
                case "onConn": {
                    let cellLocks = data.body.cellLocks;
                    if (cellLocks === null) {
                        cellLocks = [];
                    }
                    localStorage.setItem("cellLocks", JSON.stringify(cellLocks));
                    const columns = data.body.columns;
                    const content = data.body.content;
                    this.name = data.body.name;
                    this.setState({
                        name: data.body.name,
                    })
                    let j = 0, k = 0;
                    let celldata = [];
                    for (let i = 0; i < content.length; i++) {
                        j = Math.floor(i / columns);
                        k = i % columns;
                        if (content[i] !== "") {
                            celldata.push({
                                    "r": j,
                                    "c": k,
                                    "v": content[i],
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
                        // functionButton:
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
                                        if (username !== cellLocks[i].Username) {
                                            message.error(cellLocks[i].Username + "正在输入，请稍等再点击")
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
                                    socket.send(JSON.stringify(data))
                                }
                            },

                        }
                    })
                    break;
                }
                case "acquire": {
                    const row = data.body.row;
                    const col = data.body.col;
                    const username = data.body.username;
                    const me = JSON.parse(localStorage.getItem("username"));

                    if (locking_col === col && locking_row === row) {
                        if (me === username) {
                            console.log("acquired");
                            locked_row = row;
                            locked_col = col;
                        } else {
                            message.error(username + " 正在编辑，请稍等再点击");
                            locked_row = -1;
                            locked_col = -1;
                            luckysheet.setCellValue(row, col, username + " 正在编辑 ");
                        }
                    } else {
                        const cellLocks = JSON.parse(localStorage.getItem("cellLocks"));
                        cellLocks.push({
                            Row: row,
                            Col: col,
                            Username: username,
                        })
                        localStorage.setItem("cellLocks", JSON.stringify(cellLocks));
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

        const {name, log, checkpoint} = this.state;
        console.log(this.state.checkpoint_num,this.checkpoint_now);

        let logContent = log.map(
            (item) =>
                <Card hoverable style={{width: 300}}
                      title={item.username + " - " + new Date(item.timestamp).toLocaleString()}>
                    <div>
                        {
                            item.old === "" ?
                                (<p> 设置 {ColMap(item.col) + RowMap(item.row)} 为 {item.new}</p>) :
                                item.new === "" ?
                                    (<p> 清空 {ColMap(item.col) + RowMap(item.row)}</p>) :
                                    (<p> 将 {ColMap(item.col) + RowMap(item.row)} 从 {item.old} 修改为 {item.new}</p>)
                        }
                        {/*<Button onClick={() => this.handleRecoverLog(item)}>恢复到此处</Button>*/}
                    </div>
                </Card>);

        let ckptContent = checkpoint.map(
            (item) =>
                <Card hoverable style={{width: 300}}
                      title={"存档" + item.cid + " - " + new Date(item.timestamp).toLocaleString()}>
                    <Button onClick={() => this.handleRecoverCkpt(item.cid)}>查看存档</Button>
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
                                <Button type={'primary'} onClick={() => {
                                    navigator.permissions.query({name: "clipboard-write"}).then(result => {
                                        if (result.state === "granted" || result.state === "prompt") {
                                            /* write to the clipboard now */
                                            navigator.clipboard.writeText(window.location.href).then(function () {
                                                console.log('Async: Copying to clipboard was successful!');
                                                message.success("已复制到剪切板")
                                            }, function (err) {
                                                console.error('Async: Could not copy text: ', err);
                                                message.error("复制剪切板失败，请手动操作")
                                            });
                                        }
                                    })
                                }}> 分享</Button>
                            </Tooltip>
                        </Col>
                        <Col span={1}>
                            <Button type={'primary'} onClick={this.openLogDrawer}>历史</Button>
                        </Col>
                        <Col span={1}>
                            <Button type={'primary'} onClick={this.openCkptDrawer}>存档</Button>
                        </Col>
                        <Col span={1}>
                            {this.state.checkpoint_num === this.checkpoint_now ?
                                (
                                    <Button type={'primary'} onClick={this.handleSave}>保存</Button>
                                ) :
                                (
                                    <Button type={'primary'} onClick={this.handleRollback}>恢复</Button>
                                )
                            }

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
