import React from 'react';
import {Button, message, Popconfirm, Table} from 'antd';
import {Link} from 'react-router-dom'
import sheet from '../assets/google_doc_sheet.png'
import {deleteSheet, recoverSheet} from "../api/sheetService";
import {MSG_WORDS} from "../api/common";
import {getUser} from "../api/userService";

export class FileList extends React.Component {

    constructor(props) {
        super(props);
        this.state = {
            sheets: [],
            deleteSheet: [],
        }
    }


    componentDidMount() {
        const callback = (data) => {
            let msg_word = MSG_WORDS[data.msg];
            if (data.success === true) {
                localStorage.setItem('sheets', JSON.stringify(data.data.sheets));
                localStorage.setItem('username', JSON.stringify(data.data.username));
            } else {
                message.error(msg_word).then(() => {
                });
            }
            let sheets = data.data.sheets;
            let file = [];
            let deletedFile = [];



            sheets.sort(function (a,b){
                let date1 = new Date(a.UpdatedAt),date2 =new Date(b.UpdatedAt);
                return date2-date1;
            })

            sheets.forEach((v, i) => {
                v.key = v.fid;
                v.last_update = new Date(v.UpdatedAt).toLocaleString();
                if (v.isDeleted === true) {
                    deletedFile.push(v);
                } else {
                    file.push(v);
                }
            })

            this.setState({
                sheets:file,
                deletedSheets:deletedFile
            })
        }
        getUser(callback);
    }

    handleDelete = (fid) => {
        const callback = (data) => {
            let msg_word = MSG_WORDS[data.msg];
            if (data.success === true) {
                message.success(msg_word).then(r => {});
            } else {
                message.error(msg_word).then(r => {});
            }
            let {sheets,deletedSheets} = this.state;
            let newSheets = [];
            for (let i = 0; i < sheets.length; i++) {
                if (sheets[i].fid !== fid) {
                    newSheets.push(sheets[i]);
                }else{
                    deletedSheets.push(sheets[i]);
                }
            }
            this.setState({
                sheets: newSheets,
                deletedSheets:deletedSheets
            })
        }
        deleteSheet(fid, callback)
    };

    handleRecover = (fid) => {
        const callback = (data) => {
            let msg_word = MSG_WORDS[data.msg];
            if (data.success === true) {
                message.success(msg_word).then(r => {});
            } else {
                message.error(msg_word).then(r => {});
            }
            let {sheets,deletedSheets} = this.state;
            let newSheets = [];
            for (let i = 0; i < deletedSheets.length; i++) {
                if (deletedSheets[i].fid !== fid) {
                    newSheets.push(deletedSheets[i]);
                }else{
                    sheets.push(deletedSheets[i]);
                }
            }
            this.setState({
                sheets: sheets,
                deletedSheets:newSheets
            })
        }
        recoverSheet(fid, callback)
    }


    render() {
        const type = this.props.type;
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

        ];

        const main_op =  {
            title: '操作',
            dataIndex: 'action',
            render: (_, record) => {
                return (
                    <Popconfirm
                        title="你确定要删除这个文档吗?"
                        onConfirm={() => this.handleDelete(record.key)}
                        okText="Yes"
                        cancelText="No"
                    >
                        <Button>删除</Button>
                    </Popconfirm>
                );
            },
        }

        const recycle_op = {
            title: '操作',
            dataIndex: 'action',
            render: (_, record) => {
                return (
                        <Button onClick={()=>this.handleRecover(record.key)}>还原</Button>
                );
            },
        }

        let sheets = []

        if(type==="main"){
            columns.push(main_op);
            sheets = this.state.sheets;
        }else{
            columns.push(recycle_op);
            sheets = this.state.deletedSheets;
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
