import React from 'react';
import {withRouter} from "react-router-dom";
import {getSheet} from "../api/sheetService";
import {WS_URL} from "../api/common";

class SheetView extends React.Component{

    constructor(props) {
        super(props);
        this.state = {
            fid: 0,
            name:"",
            checkpoint:[],
            columns:60,
            content:"",
        }
    }

    componentDidMount() {
        console.log("Mount");

        const query = this.props.location.search;
        const arr = query.split('&');
        const fid = parseInt(arr[0].substr(4));

        const luckysheet = window.luckysheet;
        const token = JSON.parse(localStorage.getItem("token"));
        const username = JSON.parse(localStorage.getItem("username"));


        const post_data = {
            token:token,
            fid:fid
        }
        console.log("post_data",post_data);

        const getData = (data) =>{
            console.log("getData",data)
            if(data.success===true)
            {
                this.setState({
                    fid:data.fid,
                    name:data.name,
                    checkpoints:data.checkpoints,
                    columns:data.columns,
                    content:this.arrTrans(data.columns,data.content),
                })
            }
        }
        getSheet(post_data,getData);

        let wsURL = WS_URL + "sheetws";
        wsURL += "?token=" + token;
        wsURL += "&fid=" + fid;

        let socket = new WebSocket(wsURL);
        socket.addEventListener('open', function (event) {
            console.log('WebSocket open: ', event);
        });
        socket.addEventListener('message', function (event) {
            console.log('WebSocket message: ', event);
            let data = JSON.parse(event.data);
            if(data.locked===true)
            {
                let row = data.row;
                let col = data.col;
                // let uid = data.col;
                let username = data.username;
                luckysheet.setCellValue(row, col, username+"is writing");
            }
            else{
                //
            }

        });
        socket.addEventListener('error', function (event) {
            console.log('WebSocket error: ', event);
        });
        luckysheet.create({
                container: "luckysheet",
                title: 'A new table', // 设定表格名称
                lang: 'zh', // 设定表格语言
                gridKey: fid,
                // loadUrl: HTTP_URL + "load",
                // loadSheetUrl: HTTP_URL + "loadSheet",
                // allowUpdate: true,
                // updateUrl: wsURL,
                // updateImageUrl: wsURL,
                //data
                plugins: ["chart"],
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
                // TODO: functionBuyyion:
                //showConfigWindowResize:true
                //forceCalculation:false,cellRightClickConfig:{},sheetRightClickConfig:{}
                //rowHeaderWidth:46,columnHeaderHeight:20,sheetFormulaBar:true,
                //defaultFontSize:11,limitSheetNameLength:true,defaultSheetNameMaxLength:31,
                //pager:null
            }
        )
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
