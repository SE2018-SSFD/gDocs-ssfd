import React from 'react';
import {Button} from "antd";

// const socket = new WebSocket('ws://localhost:8080');
//
// // Connection opened
// socket.addEventListener('open', function (event) {
//     socket.send('Hello Server!');
// });
//
// // Listen for messages
// socket.addEventListener('message', function (event) {
//     console.log('Message from server ', event.data);
// });

const options = {
    container: "luckysheet",
    plugins: ['chart'],
    title: 'Hello', // 设定表格名称
    lang: 'zh', // 设定表格语言
    userInfo: 'User',
    myFolderUrl: '/',
    showinfobar: false,
    hook: {
        // 进入单元格编辑模式之前触发。
        cellEditBefore:function(range ){
            console.info(range);
        },

        //更新这个单元格后触发
        cellUpdated:function(r,c,oldValue, newValue, isRefresh){
            console.info('cellUpdated',r,c,oldValue, newValue, isRefresh);
        },
    }
}

const luckysheet = window.luckysheet;

export class LuckySheet extends React.Component {

    componentDidMount() {
        luckysheet.create(options)
    }

    HandleOutput(){
        console.log(luckysheet.getluckysheetfile())
    }

    render() {
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
                <Button onClick={this.HandleOutput}> 输出 </Button>
                <div
                    id="luckysheet"
                    style={luckyCss}
                />
            </div>
        )
    }
}
