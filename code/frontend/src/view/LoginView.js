import React from 'react';
import {withRouter} from "react-router-dom";
import {LoginHeader} from "../components/LoginHeader";
import {Col, Row} from "antd";
import LoginForm from "../components/LoginForm";
import loginPic1 from "../assets/loginpic.svg";
import loginPic2 from "../assets/loginpic2.svg";
import '../css/login.css'

class LoginView extends React.Component{

    componentDidMount(){
    }

    render(){
        return(
            <div className={"login-background"}>
                <LoginHeader />
                <Row align={"middle"} style={{paddingTop: 40}}>
                    <Col span={8} offset={4}>
                        <img src={loginPic2} alt={"show_pic"}/>
                        <img src={loginPic1} alt={"show_pic"}/>
                    </Col>
                    <Col span={8} offset={3}>
                        <LoginForm/>
                    </Col>
                </Row>

            </div>
        );
    }
}

export default withRouter(LoginView);
