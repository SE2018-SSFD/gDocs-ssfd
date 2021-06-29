import React from 'react';
import {withRouter} from "react-router-dom";
import {Layout} from "antd";
import {LuckySheet} from "../components/LuckySheet";
import {DocHeader} from "../components/DocHeader";
import {getSheet} from "../services/sheetService";

const {Content,Footer} = Layout

class DocView extends React.Component{

    constructor(props) {
        super(props);
        this.state={
            data:""
        }
    }

    componentDidMount(){
        const token = JSON.parse(localStorage.getItem('token'));
        const query = this.props.location.search;
        //分离
        const arr = query.split('&');
        //？id后面开始
        const fid = parseInt(arr[0].substr(4));

        const data = {
            token:token,
            fid:fid,
        }
        const callback = (data) =>{
            this.setState({
                data: data.data
            })
        }
        getSheet(data,callback);
    }

    render(){
        const {data} = this.state;
        return(
         <Layout>
               <DocHeader data={data}/>
                <Content style={{margin: '24px 16px 0'}}>
                    <LuckySheet data={data} fid/>
                </Content>

                <Footer style={{textAlign: 'center'}}>SSF Doc ©2021 Created by SJTU Super SoFtware
                    Developer
                </Footer>
            </Layout>
        );
    }
}

export default withRouter(DocView);
