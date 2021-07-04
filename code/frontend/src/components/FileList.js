import React, {useCallback} from 'react';
import {Button, message, Popconfirm, Table} from 'antd';
import {Link} from 'react-router-dom'
import sheet from '../assets/google_doc_sheet.png'
import {deleteSheet} from "../api/sheetService";
import {MSG_WORDS} from "../api/common";

export class FileList extends React.Component {
    deleteHandler=(text,record)=>{
        const callback = (data)=>{
            console.log(data);
            let msg_word = MSG_WORDS[data.msg];
            if(data.success===true){
                message.success(msg_word);
            }
            else{
                message.error(msg_word);
            }
            let sheets = this.state.sheets;
            let newSheets = [];
            for (let i = 0; i< sheets.length;i++)
            {
                if (sheets[i].fid !== record.fid)
                {
                    newSheets.push(sheets[i]);
                }
            }
            this.setState({
                sheets:newSheets,
                first:false,
            })
        }
        deleteSheet(record.fid,callback)
    };

    constructor(props) {
        super(props);
        this.state = {
            sheets:[],
            first:true,
        }
    }

    componentDidMount() {
        let sheets = this.props.content;
        sheets.forEach((x) => {
            x.key = x.fid;
            x.last_update = new Date(x.UpdatedAt).toLocaleString();
        })

        this.setState({
            sheets:sheets
        })
    }

    render() {
        const columns = [
            {
                title: '名称',
                dataIndex: 'name',
                render: (text, record) =>
                    <div style={{display: "inline-flex"}}>
                        <img src={sheet} height={20} width={20} alt={"sheet"}/>
                        <Link to={{
                            pathname: '/sheet',
                            search: '?id=' + record.key
                        }}
                              target="_blank"
                        >
                            <p style={{marginLeft: "5px"}}>{text}</p>
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
                            onConfirm={() => this.deleteHandler(text,record)}
                            okText="Yes"
                            cancelText="No"
                        >
                            <Button>删除</Button>
                        </Popconfirm>
                    );
                },
            }
        ];

        let sheets = this.props.content;
        sheets.forEach((x) => {
            x.key = x.fid;
            x.last_update = new Date(x.UpdatedAt).toLocaleString();
        })

        if(this.state.first!==true){
            sheets = this.state.sheets;
        }

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
