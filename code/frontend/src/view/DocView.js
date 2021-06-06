import React from 'react';
import {withRouter} from "react-router-dom";
import {Layout} from "antd";
import {LuckySheet} from "../components/LuckySheet";
import {DocHeader} from "../components/DocHeader";

const {Content,Footer} = Layout

class DocView extends React.Component{


    componentDidMount(){
        let user = localStorage.getItem("user");
        this.setState({user:user});
    }

    render(){
        return(
         <Layout>
               <DocHeader/>
                <Content style={{margin: '24px 16px 0'}}>
                    <LuckySheet/>
                </Content>

                <Footer style={{textAlign: 'center'}}>SSF Doc ©2021 Created by SJTU Super SofTware
                    Developer
                </Footer>
            </Layout>
        );
    }
}

export default withRouter(DocView);
