import 'antd/dist/antd.css';
import '../css/LoginForm.css'

import {FileTextOutlined} from '@ant-design/icons';
import {Button, Col, Input, Row} from 'antd';
import React from 'react';
import {Link} from 'react-router-dom';

import {register} from '../services/userService';

class RegisterForm extends React.Component {
    constructor(props) {
        super(props);
        this.state = {username: '', password: '',email:''};
    }

    usernameOnChange = (e) => {
        this.setState({username: e.target.value});
    };

    passwordOnChange = (e) => {
        this.setState({password: e.target.value});
    };

    emailOnChange = (e) => {
        this.setState({email:e.target.value});
    }

    onSubmit = () => {
        const data = {
            userName: this.state.username,
            password: this.state.password,
            email: this.state.email,
        };
        register(data);
    }

    render() {
        return (
            <div className="login-form-container">
                <Row className="login-header">
                    <Col span={24}>
                        <div style={{textAlign: 'center'}}>
                            <div>
                                <FileTextOutlined className="icon"/>
                            </div>
                            <div>
                                <h1>注册到超级文档</h1>
                            </div>
                        </div>
                    </Col>
                </Row>
                <Row className="login-form">
                    <Col span={24}>
                        <div>
                            <p className="login-hint">用户名</p>
                        </div>
                        <div>
                            <Input placeholder="请输入用户名" value={this.state.username}
                                   onChange={this.usernameOnChange} className="login-input"/>
                        </div>
                        <div>
                            <p className="login-hint">密码</p>
                        </div>
                        <div>
                            <Input placeholder="请输入密码" value={this.state.password} onChange={this.passwordOnChange}
                                   className="login-input"/>
                        </div>
                        <div>
                            <p className="login-hint">邮箱</p>
                        </div>
                        <div>
                            <Input placeholder="请输入邮箱" value={this.state.password} onChange={this.passwordOnChange}
                                   className="login-input"/>
                        </div>
                        <div style={{textAlign: 'center'}}>
                            <Button onClick={this.onSubmit} className="login-button">
                                注&nbsp;&nbsp;&nbsp;册
                            </Button>
                        </div>
                    </Col>
                </Row>
                <Row className="login-register">
                    <Col span={24} style={{padding: '15px 20px'}}>
                        <div style={{textAlign: 'center'}}>
                            <p className="register-link">
                                已有账号？<Link to={{pathname: '/login'}}>立即登录</Link>
                            </p>
                        </div>
                    </Col>
                </Row>
            </div>
        );
    }
}

export default RegisterForm;
