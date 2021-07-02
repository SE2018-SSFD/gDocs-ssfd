import React from 'react';
import {Button, Popconfirm, Table} from 'antd';
import {Link} from 'react-router-dom'
import sheet from '../assets/google_doc_sheet.png'
import {deleteSheet} from "../api/sheetService";

const columns = [
    {
        title: '名称',
        dataIndex: 'name',
        render: (text, record) =>
            <div style={{display:"inline-flex"}}>
                <img src={sheet} height={20} width={20} alt={"sheet"}/>
                <Link to={{
                    pathname: '/sheet',
                    search: '?id=' + record.key
                }}
                      target="_blank"
                >
                    <p style={{marginLeft:"5px"}}>{text}</p>
                </Link>
            </div>
    },
    {
        title: '来自',
        dataIndex: 'owner',
    },
    {
        title: '最近查看',
        dataIndex: 'last_update',
    },
    {
        title: '操作',
        dataIndex: 'action',
        render: (text, record) => {
            return (
                <Popconfirm
                    title="你确定要删除这个文档吗?"
                    onConfirm={() => this.deleteConfirm(record.key)}
                    okText="Yes"
                    cancelText="No"
                >
                    <Button>删除</Button>
                </Popconfirm>
            );
        },
        deleteConfirm(fid) {
            deleteSheet(fid);
        }
    }
];

export class FileList extends React.Component {

    constructor(props) {
        super(props);
        this.state = {}
    }

    componentDidMount() {

    }

    render() {
        const sheets = this.props.content;

        sheets.forEach((x) => {
            x.key = x.fid
            x.last_update = new Date(x.UpdatedAt).toLocaleString()
        })
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
