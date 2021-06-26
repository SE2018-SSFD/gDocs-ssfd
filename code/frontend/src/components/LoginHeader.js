import React from "react";
import {Button, Col, Divider, Image, Input, Layout, Row, Space} from "antd";
import logo from '../assets/logo.png'
import {MenuOutlined} from "@ant-design/icons";
import '../css/login.css'
import {Link} from "react-router-dom";

const {Header} = Layout

export class LoginHeader extends React.Component {
    render() {
        return (
            <Header style={{paddingTop: 20}} className={"login-header"}>
                <Row>
                    <Col span={1} offset={0.5}>
                        <Image preview={false} src={logo} alt={'logo'} height={50} width={50}/>
                    </Col>
                    <Col span={1}>
                        <strong><h1>SSFDoc</h1></strong>
                    </Col>
                    <Col span={1} offset={17}>
                        <Space>
                            <Button type="text">下载</Button>
                            <Link to={'/'}><Button type="text">进入官网</Button></Link>
                        </Space>
                    </Col>

                    <Col span={1} offset={2}>
                        <Divider type={'vertical'}/>
                        <MenuOutlined/>
                    </Col>
                </Row>
            </Header>
        )
    }
}
