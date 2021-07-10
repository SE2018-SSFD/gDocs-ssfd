import React from 'react';
import {Link, withRouter} from "react-router-dom";
import {
    commitSheet,
    getAll,
    getSheet,
    getSheetCkpt,
    getSheetLog,
    rollbackSheet,
    testWS,
    uploadImage
} from "../api/sheetService";
import {GET_HTTP_URL, GET_WS_URL, MSG_WORDS} from "../api/common";
import {Button, Card, Col, Divider, Drawer, Layout, message, Row, Tooltip, Upload} from "antd";
import {
    CheckCircleOutlined,
    EditOutlined,
    FolderOutlined,
    LeftOutlined,
    MenuOutlined,
    StarOutlined,
    UploadOutlined
} from "@ant-design/icons";
import {UserAvatar} from "../components/UserAvatar";
import {ColMap, RowMap} from "../utils";
import Modal from "antd/es/modal/Modal";

const {Header} = Layout
const luckysheet = window.luckysheet;
let socket;
let locking_row = -1, locking_col = -1, locked_row = -1, locked_col = -1;

class SheetView extends React.Component {

    normFile = (e) => {
        console.log('Upload event:', e);
        if (Array.isArray(e)) {
            return e;
        }
        return e && e.fileList;
    };

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
            picVisible: false,
            picDrawerVisible: false,
            uploading: false,
            picList: [],
            fileList: [],
        }
    }

    handleUpload = () => {
        const {fileList} = this.state;
        this.setState(({
            uploading: true,
        }))

        const callback = (data) => {
            console.log(data);
            const msg_word = MSG_WORDS[data.msg];
            if (data.success === true) {
                this.setState({
                    fileList: [],
                    uploading: false,
                    picVisible: false,
                    picDrawerVisible: false,
                });
                message.success(msg_word);
            }
        }
        uploadImage(this.fid, fileList[0], callback)
    }

    componentDidMount() {
        const query = this.props.location.search;
        const arr = query.split('&');
        const fid = parseInt(arr[0].substr(4));

        this.fid = fid;

        const callback = (data) => {
            if (data.success === true) {
                this.checkpoint_now = data.data.checkpoint_num;
                const checkPointBrief = data.data.checkPointBrief === null ? [] : data.data.checkPointBrief.reverse()
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
            const checkPointBrief = data.data.checkPointBrief === null ? [] : data.data.checkPointBrief.reverse()
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

    openPicDrawer = () => {
        const callback = (data) => {
            console.log(data);
            const msg_word = MSG_WORDS[data.msg];
            if (data.success === true) {
                this.setState({
                    picList: data.data === null ? [] : data.data,
                    picDrawerVisible: true,
                })
                message.success(msg_word);
            } else {
                message.error(msg_word);
            }
        }
        getAll(this.fid, callback)
    }

    closePicDrawer = () => {
        this.setState({
            picDrawerVisible: false,
        })
    }
    openPic = () => {
        this.setState({
            picVisible: true,
        })
    }

    closePic = () => {
        this.setState({
            picVisible: false,
        })
    }

    handlePic = () => {
        luckysheet.insertImage("data:image/png;base64,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")
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
                for (let i = 0; i < content.length; i++) {
                    j = Math.floor(i / columns);
                    k = i % columns;
                    if (content[i].indexOf("getchunk") !== -1) {
                        console.log("image");
                        luckysheet.insertImage(content[i],0, j, k)
                    }
                }
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
                    for (let i = 0; i < content.length; i++) {
                        j = Math.floor(i / columns);
                        k = i % columns;
                        if (content[i].indexOf("getchunk") !== -1) {
                            console.log("image");
                            luckysheet.insertImage(content[i],0, j, k, )
                        }
                    }
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
                    if (content.indexOf("getchunk") !== -1) {
                        console.log("image")
                        luckysheet.insertImage(content,0,row,col);
                    }
                    break;
                }
                case "release": {
                    let row = data.body.row;
                    let col = data.body.col;
                    let cellLocks = JSON.parse(localStorage.getItem("cellLocks"));
                    const username = JSON.parse(localStorage.getItem("username"));
                    for (let i = cellLocks.length - 1; i >= 0; i--) {
                        if (cellLocks[i].Row === row && cellLocks[i].Col === col) {
                            if (cellLocks[i].Username === username) {
                                console.log("release success");
                                locked_col = -1;
                                locked_row = -1;
                            } else {
                                console.log("others have release the lock");
                                if (luckysheet.getCellValue(row, col) === cellLocks[i].Username + " 正在编辑 ") {
                                    luckysheet.clearCell(row, col);
                                }

                            }
                            cellLocks.splice(i, 1);
                        }
                    }

                    localStorage.setItem("cellLocks", JSON.stringify(cellLocks));
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

        const {name, log, checkpoint, picVisible, picDrawerVisible, uploading, fileList, picList} = this.state;

        const token = JSON.parse(localStorage.getItem("token"));
        const fid = this.fid;
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

        let picContent = picList.map(
            (item) =>
                <Card hoverable style={{width: 300}}
                      title={item}>
                    <p>{GET_HTTP_URL() + "getchunk?fid=" + fid + "&chunk=" + item}</p>
                    <img src={GET_HTTP_URL() + "getchunk?fid=" + fid + "&chunk=" + item}/>
                </Card>
        );

        const props = {
            onRemove: file => {
                this.setState(state => {
                    const index = state.fileList.indexOf(file);
                    const newFileList = state.fileList.slice();
                    newFileList.splice(index, 1);
                    return {
                        fileList: newFileList,
                    };
                });
            },
            beforeUpload: file => {
                this.setState(state => ({
                    fileList: [...state.fileList, file],
                }));
                return false;
            },
        };

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
                        <Col span={1} offset={5}>
                            <MenuOutlined/>
                        </Col>
                        <Col span={1}>
                            <EditOutlined/>
                        </Col>
                        <Divider type={"vertical"}/>
                        <Col span={1}>
                            <Button type={'primary'} onClick={this.openPicDrawer}>图片</Button>
                        </Col>
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
                    title="历史"
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
                    title="存档"
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

                <Drawer
                    title="复制URL到单元格来插入图片"
                    onClose={this.closePicDrawer}
                    visible={this.state.picDrawerVisible}
                    bodyStyle={{paddingBottom: 80}}
                    width={320}
                    footer={
                        <div
                            style={{
                                textAlign: 'right',
                            }}
                        >
                            <Button onClick={this.openPic}>上传图片</Button>
                            <Button onClick={this.closePicDrawer} style={{marginRight: 8}}>
                                返回
                            </Button>
                        </div>
                    }
                >
                    {picContent}
                </Drawer>

                <Modal title="上传图片" visible={picVisible} onCancel={this.closePic}>
                    <Upload {...props}>
                        <Button icon={<UploadOutlined/>}>选择文件</Button>
                    </Upload>
                    <Button
                        type="primary"
                        onClick={this.handleUpload}
                        disabled={fileList.length === 0}
                        loading={uploading}
                        style={{marginTop: 16}}
                    >
                        {uploading ? '上传中' : '开始上传'}
                    </Button>
                </Modal>
            </div>
        )
    }
}

export default withRouter(SheetView);
