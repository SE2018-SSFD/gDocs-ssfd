import React from 'react';
import {withRouter} from "react-router-dom";
import {getSheet, testWS} from "../api/sheetService";
import {WS_URL} from "../api/common";

const luckysheet = window.luckysheet;
let socket;
let fid;

class SheetView extends React.Component {


    constructor(props) {
        super(props);
        this.state={
            columns:0,
            content:[],
        }
    }

    connectWS(data) {
        console.log("connectWS", data);
        const token = JSON.parse(localStorage.getItem("token"));
        let url = WS_URL + 'sheetws?token=' + token + "&fid=" + fid;
        if (data.success === false) {
            url = data.data;
        }
        socket = new WebSocket(url);
        socket.addEventListener('open', function (event) {
            console.log('WebSocket open: ', event);
        });
        socket.addEventListener('message', function (event) {
            console.log('WebSocket message: ', event);
            let data = JSON.parse(event.data);
            console.log(data);
            switch (data.msgType) {
                case "acquire": {
                    let row = data.body.row;
                    let col = data.body.col;
                    let username = data.body.username;
                    luckysheet.setCellValue(row, col, username + " is writing");
                    break;
                }
                case "modify": {
                    let row = data.body.row;
                    let col = data.body.col;
                    let content = data.body.content;
                    luckysheet.setCellValue(row, col, content);
                    break;
                }
                case "release": {
                    let row = data.body.row;
                    let col = data.body.col;
                    if (luckysheet.getCellValue(row, col).indexOf("is writing") !== -1) {
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

    componentDidMount() {
        console.log("Mount");

        const query = this.props.location.search;
        const arr = query.split('&');
        fid = parseInt(arr[0].substr(4));

        const token = JSON.parse(localStorage.getItem("token"));
        const username = JSON.parse(localStorage.getItem("username"));

        const post_data = {
            token: token,
            fid: fid
        }
        console.log("post_data", post_data);

        const getData = (data) => {
            console.log("getData", data);
            if (data.success === true) {
                let name = data.data.name;
                let columns = data.data.columns;
                let content = data.data.content;
                let j = 0, k = 0;
                let celldata= [];
                for (let i = 0; i< content.length;i++)
                {
                    j = Math.floor(i / columns);
                    k = i % columns;
                    if(content[i]!=="") {
                        celldata.push({
                                "r": j,
                                "c": k,
                                "v": content[i],
                            }
                        )
                    }
                }
                console.log(celldata);
                luckysheet.create({
                        container: "luckysheet",
                        title: name, // 设定表格名称
                        lang: 'zh', // 设定表格语言
                        gridKey: fid,
                        // loadUrl: HTTP_URL + "load",
                        // loadSheetUrl: HTTP_URL + "loadSheet",
                        // allowUpdate: true,
                        // updateUrl: wsURL,
                        // updateImageUrl: wsURL,
                        data:[{ "name": "Sheet1", color: "", "status": "1", "order": "0", "celldata": celldata, "config": {}, "index":0 }],
                        // plugins: ["chart"],
                        //column:60,row:84,autoFormatw:false,accuracy:undefined,allowCopy:true
                        //showtoolbar:true,showtoolbarConfig,showinfobar:true,showsheetbar:true
                        //showsheetbarconfig:{},showstaticBar:true,showstaticBarConfig:{},
                        //enableAddRow,enableAddBackTop:true,
                        userInfo: username,
                        userMenuItem: [
                            {url: "/", "icon": '<i class="fa fa-folder" aria-hidden="true"></i>', "name": "我的表格"},
                        ],
                        myFolderUrl: "/",
                        //devicePixelRatio:window.devicePixelRatio
                        // TODO: functionButton:
                        //showConfigWindowResize:true
                        //forceCalculation:false,cellRightClickConfig:{},sheetRightClickConfig:{}
                        //rowHeaderWidth:46,columnHeaderHeight:20,sheetFormulaBar:true,
                        //defaultFontSize:11,limitSheetNameLength:true,defaultSheetNameMaxLength:31,
                        //pager:null
                        hook: {
                            // 进入单元格编辑模式之前触发。
                            cellEditBefore: function (range) {
                                console.info('cellEditBefore', range[0]);
                                const data = {
                                    msgType: "acquire",
                                    body: {
                                        row: range[0].row_focus,
                                        col: range[0].column_focus
                                    }
                                }
                                socket.send(JSON.stringify(data))
                            },
                            // cellEdit:function(range ){
                            //     console.info('cellEdit',range);
                            //     socket.send("cellEdit" +JSON.stringify(range))
                            // },

                            // cellUpdateBefore:function(r,c,value,isRefresh){
                            //     console.info('cellUpdateBefore',r,c,value,isRefresh)
                            //     socket.send("cellUpdateBefore"+ r+ c+ value+ isRefresh)
                            // },

                            //更新这个单元格后触发
                            cellUpdated: function (r, c, oldValue, newValue, isRefresh) {
                                console.info('cellUpdated', r, c, oldValue, newValue, isRefresh);
                                let content = "";
                                console.log(newValue);
                                if (newValue.ct.t === "inlineStr") {
                                    content = newValue.ct.s[0].v;
                                } else if (newValue.ct.t === "n") {
                                    content = newValue.v.toString();
                                } else if (newValue.ct.t === "g") {
                                    content = newValue.v;
                                }
                                const data1 = {
                                    msgType: "modify",
                                    body: {
                                        row: r,
                                        col: c,
                                        content: content
                                    }
                                }
                                console.log(data1)
                                socket.send(JSON.stringify(data1))
                                const data2 = {
                                    msgType: "release",
                                    body: {
                                        row: r,
                                        col: c,
                                    }
                                }
                                socket.send(JSON.stringify(data2))
                            },
                        }
                    }
                )

            }
        }

        getSheet(post_data, getData);
        testWS(fid, this.connectWS)
    }

    render() {
        // const {columns,content} = this.state;
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
