import React from 'react';
import {Table} from 'antd';
import {Link} from 'react-router-dom'
import sheet from '../assets/google_doc_sheet.png'

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
        title: '创建者',
        dataIndex: 'owner',
    },
    {
        title: '最近修改',
        dataIndex: 'last_update',
    },
];

export class FileList extends React.Component {

    constructor(props) {
        super(props);
        this.state = {}
    }

    componentDidMount() {

    }

    render() {
        const sheets = JSON.parse(localStorage.getItem("sheets"));

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
