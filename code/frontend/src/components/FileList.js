import React from 'react';
import { Table, Image} from 'antd';
import {Link} from 'react-router-dom'
import sheet from '../assets/google_doc_sheet.png'

const columns = [
    {
        title: '名称',
        dataIndex: 'name',
        render: (text,record) =>
            <Link to={{
                pathname: '/doc',
                search: '?id=' + record.key}}
                  target="_blank"
            >
                <Image src={sheet} height={20} width={20} preview={false}/>
            {text}
            </Link>
    },
    {
        title: '来自',
        dataIndex: 'from',
    },
    {
        title: '最近查看',
        dataIndex: 'recentlyOpen',
    },
];
// const data = [
//     {
//         key: '1',
//         name: '表格1',
//         from: '我',
//         recentlyOpen:'今天1：00'
//     },
//     {
//         key: '2',
//         name: '表格2',
//         from: '我',
//         recentlyOpen:'今天2：00'
//     },
//     {
//         key: '3',
//         name: '表格3',
//         from: '我',
//         recentlyOpen:'今天3：00'
//     },
//     {
//         key: '4',
//         name: '表格4',
//         from: '我',
//         recentlyOpen:'今天4：00'
//     },
//     {
//         key: '5',
//         name: '表格5',
//         from: '我',
//         recentlyOpen:'今天5：00'
//     },
// ];



export class FileList extends React.Component{

    render() {
        const sheets = JSON.parse(localStorage.getItem("sheets"));

        sheets.forEach((x)=>{x.key=x.fid})
        console.log(sheets);
        return (
            <div>
                <Table
                    columns={columns}
                    dataSource={sheets}
                />
            </div>
        );
    }
}
