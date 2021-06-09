import React from 'react';
import {withRouter} from "react-router-dom";
import '../css/home.css'
import docs from '../assets/docs.png'

import { Button, Col, Divider, Image, Layout, Menu, Row, Space, Typography} from 'antd';
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

const {Content, Footer, Sider} = Layout;

const {Title} = Typography;

const {SubMenu} = Menu;

class HomeView extends React.Component {

    componentDidMount() {
        let user = localStorage.getItem("user");
        this.setState({user: user});
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
                            <Image src={docs} alt={'docs'} height={50} width={50} preview={false}/>
                        </Col>
                        <Col span={18} offset={1}>
                            <Title> SSFDoc</Title>
                        </Col>
                    </Row>
                    <Row>
                        <Col span={20} offset={2}>
                            <Space direction="vertical">
                                <Button size="large" type="primary" icon={<PlusOutlined/>} block={true}>新建</Button>
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
                                最近查看
                            </Menu.Item>
                            <Menu.Item key="star">
                                星标
                            </Menu.Item>
                        </Menu>
                        <div className="site-layout-background" style={{padding: 24, minHeight: 360}}>
                            <FileList/>
                        </div>
                    </Content>

                    <Footer style={{textAlign: 'center'}}>SSF Doc ©2021 Created by SJTU Super SofTware
                        Developer
                    </Footer>
                </Layout>
            </Layout>
        )

    }
}

export default withRouter(HomeView);
