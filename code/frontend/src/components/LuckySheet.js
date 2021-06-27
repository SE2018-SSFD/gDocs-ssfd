import React from 'react';
import {websocketUrl} from "../utils/config";

const luckysheet = window.luckysheet;

export class LuckySheet extends React.Component {

    constructor(props) {
        super(props);
        this.state={
            data: "",
            connected:""
        }
    }

    componentDidMount() {

    }

    render() {
        console.log(this.props.data);
        if(this.props.data!=="")
        {
            let fid = this.props.data.fid;
            let wsURL = websocketUrl + "sheetws";

            const enableJWT = true;
            if (enableJWT) {
                // This is just a signature and a payload of an example content,
                // please replace this with your logic.
                //
                // Add a random letter in front of the token to make it
                // invalid and see that this client is not allowed to dial the websocket server.
                const token = JSON.parse(localStorage.getItem("token"));
                wsURL += "?token=" + token;
            }
            wsURL +="&fid="+fid;
            console.log(wsURL)
            let socket = new WebSocket(wsURL);
            socket.addEventListener('open', function (event) {
                console.log('WebSocket open: ', event);
            });
            socket.addEventListener('message', function (event) {
                console.log('WebSocket message: ', event);
            });
            socket.addEventListener('error', function (event) {
                console.log('WebSocket error: ', event);
            });
            luckysheet.create({
                container: "luckysheet",
                plugins: ['chart'],
                title: 'Hello', // 设定表格名称
                lang: 'zh', // 设定表格语言
                userInfo: 'User',
                myFolderUrl: '/',
                showinfobar: false,
                allowUpdate: true,
                updateUrl: 'localhost:8080',
                // loadUrl:'localhost:8080',
                hook: {
                    // 进入单元格编辑模式之前触发。
                    cellEditBefore: function (range) {
                        console.info('cellEditBefore', range);
                        socket.send("cellEditBefore" + JSON.stringify(range))
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
                        socket.send("cellUpdate" + r + c + oldValue + newValue + isRefresh)
                    },
                }
            });
        }

        const luckyCss = {
            margin: '0px',
            padding: '0px',
            position: 'absolute',
            width: '100%',
            height: '90%',
            left: '0px',
            top: '60px',
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
