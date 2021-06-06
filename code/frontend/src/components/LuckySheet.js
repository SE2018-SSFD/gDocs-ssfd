import React from 'react';

const options = {
    container: "luckysheet",
    plugins:['chart'],
    title: 'Hello', // 设定表格名称
    lang: 'zh', // 设定表格语言
    userInfo:'User',
    myFolderUrl:'/',
    showinfobar:false
}

export class LuckySheet extends React.Component {

    componentDidMount() {
        const luckysheet = window.luckysheet;
        luckysheet.create(options)
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
            <div
                id="luckysheet"
                style={luckyCss}
            />
        )
    }
}
