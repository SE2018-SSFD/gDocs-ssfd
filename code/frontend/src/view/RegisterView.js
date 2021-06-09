import React from 'react';
import {withRouter} from "react-router-dom";
import {LoginHeader} from "../components/LoginHeader";
import {Col, Row} from "antd";
import LoginForm from "../components/LoginForm";
import loginPic1 from "../assets/loginpic.svg";
import loginPic2 from "../assets/loginpic2.svg";
import '../css/login.css'
import RegisterForm from "../components/RegisterForm";

class RegisterView extends React.Component{

    componentDidMount(){
        let user = localStorage.getItem("user");
        this.setState({user:user});
    }

    render(){
        return(
            <div className={"login-background"}>
                <LoginHeader />
                <Row align={"middle"} style={{paddingTop: 40}}>
                    <Col span={8} offset={4}>
                        <img src={loginPic2}/>
                        <img src={loginPic1}/>
                    </Col>
                    <Col span={8} offset={3}>
                        <RegisterForm/>
                    </Col>
                </Row>

            </div>
        );
    }
}

export default withRouter(RegisterView);
