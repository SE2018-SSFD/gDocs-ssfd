import React from 'react';
import {withRouter} from "react-router-dom";
import {getSheet, testWS} from "../api/sheetService";
import {HTTP_URL, MSG_WORDS, WS_URL} from "../api/common";
import {message} from "antd";

const luckysheet = window.luckysheet;
let socket;
let locking_row = -1, locking_col = -1, locked_row = -1, locked_col = -1;

class SheetView extends React.Component {

    constructor(props) {
        super(props);
        this.fid = 0;
        this.cellLocks = [];
        this.columns = 0;
        this.content = [];
        this.username = "";
        this.checkpoint_num = 0;
        this.url = "";

        //记录编辑

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

        getSheet(HTTP_URL+'getsheet', get_data, () => {
        })

        testWS(this.fid, this.connectWS);
    }

    handleLog = () => {
        console.log("log");
    }

    connectWS = (data) => {
        const token = JSON.parse(localStorage.getItem("token"));
        const username = JSON.parse(localStorage.getItem("username"));
        this.username = username;

        this.url = WS_URL + 'sheetws?token=' + token + "&fid=" + this.fid;
        if (data.success === false) {
            this.url = data.data;
        }
        else{
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
                    this.cellLocks = data.body.cellLocks;
                    this.columns = data.body.columns;
                    this.content = data.body.content;
                    this.name = data.body.name;
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
                        showinfobar: true,
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
                    if (locking_col === col && locking_row === row) {
                        if (this.username === username) {
                            console.log("acquired");
                            locked_row = row;
                            locked_col = col;
                        } else {
                            console.log("not acquired");
                            message.error(username + " 正在输入，请稍等再点击");
                            locked_row = -1;
                            locked_col = -1;
                            luckysheet.setCellValue(row, col, username + " 正在输入 ");
                        }
                    } else {
                        luckysheet.setCellValue(row, col, username + " 正在输入 ");
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
                    let row = data.body.row;
                    let col = data.body.col;
                    if (locked_col === col && locked_row === row) {
                        console.log("release success");
                        locked_col = -1;
                        locked_row = -1;
                    } else {
                        console.log("others release", row, col);
                        luckysheet.clearCell(row, col);
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
            height: '100%',
            left: '0px',
            top: '0px',
        }
        return (
            <div>
                <div
                    id="luckysheet"
                    style={luckyCss}
                />
            </div>
        )
    }
}

export default withRouter(SheetView);
