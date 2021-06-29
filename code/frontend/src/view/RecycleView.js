import React from 'react';
import {withRouter} from "react-router-dom";
import '../css/home.css'
import logo from '../assets/logo.png'

import {Button, Col, Divider, Image, Layout, Menu, Row, Space} from 'antd';
import {
    AppstoreAddOutlined,
    DeleteOutlined,
    FileTextOutlined,
    FolderOutlined,
    HomeOutlined,
    PlusOutlined,
    UploadOutlined
} from '@ant-design/icons';
import {FileList} from "../components/FileList";
import {HomeHeader} from "../components/HomeHeader";
import {newSheet} from "../services/sheetService";

const {Content, Footer, Sider} = Layout;

const {SubMenu} = Menu;

class RecycleView extends React.Component {

    componentDidMount() {
    }


    newSheet(){
        const token = JSON.parse(localStorage.getItem("token"));
        const data = {
            token:token,
            name:'a new sheet',
            intiRows: 84,
            initColumns:60,
        };

        newSheet(data)
    }

    render() {
        return (
            <Layout>
                <Sider className='sider' width={264} style={{
                    background: '#fafbfc',
                }
                }>
                    <Row align={"middle"}>
                        <Col span={4} offset={1}>
                            <Image src={logo} alt={'docs'} height={100} width={200} preview={false}/>
                        </Col>
                        <Col span={18} offset={1}>

                        </Col>
                    </Row>
                    <Row>
                        <Col span={20} offset={2}>
                            <Space direction="vertical">
                                <Button size="large" type="primary" icon={<PlusOutlined/> } onClick={this.newSheet} block={true}>新建</Button>
                                <Button size="large" icon={<UploadOutlined/>} block={true}>导入本地文件</Button>
                            </Space>
                        </Col>
                    </Row>

                    <Menu mode="inline">
                        <Menu.Item key="1" icon={<HomeOutlined/>}>
                            首页
                        </Menu.Item>
                        <SubMenu key="2" icon={<FileTextOutlined/>} title="我的文档">
                            <Menu.Item key="3" icon={<FolderOutlined/>}>与我共享</Menu.Item>
                            <Menu.Item key="4" icon={<FolderOutlined/>}>Hi, 欢迎使用SSFDoc</Menu.Item>
                        </SubMenu>
                        <Divider/>
                        <Menu.Item key={"5"} icon={<AppstoreAddOutlined/>}>模板</Menu.Item>
                        <Menu.Item key={"6"} icon={<DeleteOutlined/>}>回收站</Menu.Item>
                    </Menu>
                </Sider>
                <Layout>
                    <HomeHeader/>
                    <Content style={{margin: '24px 16px 0'}}>
                        <Menu mode="horizontal">
                            <Menu.Item key="nearlyLook">
                                回收站
                            </Menu.Item>
                        </Menu>
                        <div className="site-layout-background" style={{padding: 24, minHeight: 360}}>
                            <FileList data={}/>
                        </div>
                    </Content>

                    <Footer style={{textAlign: 'center'}}>SSF Doc ©2021 Created by SJTU Super SoFtware Developer
                    </Footer>
                </Layout>
            </Layout>
        )
    }
}

export default withRouter(HomeView);
