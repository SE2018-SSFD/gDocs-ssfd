import React from 'react';
import {withRouter} from "react-router-dom";
import '../css/home.css'
import logo from '../assets/logo.png'

import {Button, Col, Divider, Image, Input, Layout, Menu, Row, Space} from 'antd';
import {
    AppstoreAddOutlined,
    BellOutlined,
    DeleteOutlined,
    FileTextOutlined,
    FolderOutlined,
    HomeOutlined,
    MenuOutlined,
    PlusOutlined,
    TabletOutlined,
    UploadOutlined
} from '@ant-design/icons';
import {FileList} from "../components/FileList";
import {newSheet} from "../api/sheetService";
import {UserAvatar} from "../components/UserAvatar";
import {getUser} from "../api/userService";

const {Header, Content, Footer, Sider} = Layout;

const {SubMenu} = Menu;
const {Search} = Input;

class HomeHeader extends React.Component {
    render() {
        return (
            <Header className="site-layout-sub-header-background" style={{padding: 0}}>
                <Row align={"middle"} justify={"center"}>
                    <Col span={8} offset={1} style={{marginTop: "18px"}}>
                        <Search placeholder="搜索"/>
                    </Col>
                    <Col span={1} offset={10}>
                        <BellOutlined/>
                    </Col>
                    <Col span={1}>
                        <TabletOutlined/>
                    </Col>
                    <Col span={1}>
                        <MenuOutlined/>
                    </Col>
                    <Col span={1} style={{marginLeft: "5px", marginBottom: "5px"}}>
                        <UserAvatar/>
                    </Col>
                </Row>
            </Header>
        )
    }
}

class HomeSider extends React.Component {

    newSheet() {
        const token = localStorage.getItem("token");
        const data = {
            token: token,
            name: 'a new sheet',
            intiRows: 84,
            initColumns: 60,
        };
        newSheet(data)
    }

    render() {
        return <Sider className="sider" width={264} style={{
            background: "#fafbfc",
        }
        }>
            <Row align={"middle"}>
                <Col span={4} offset={1}>
                    <Image src={logo} alt={"docs"} height={100} width={200} preview={false}/>
                </Col>
                <Col span={18} offset={1}>

                </Col>
            </Row>
            <Row>
                <Col span={20} offset={2}>
                    <Space direction="vertical">
                        <Button size="large" type="primary" icon={<PlusOutlined/>} onClick={this.newSheet}
                                block={true}>新建</Button>
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
                <Menu.Item icon={<AppstoreAddOutlined/>}>模板</Menu.Item>
                <Menu.Item icon={<DeleteOutlined/>}>回收站</Menu.Item>
            </Menu>
        </Sider>;
    }
}

class HomeContent extends React.Component {
    render() {
        return <Content style={{margin: "24px 16px 0"}}>
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
        </Content>;
    }
}

class HomeFooter extends React.Component {
    render() {
        return <Footer style={{textAlign: "center"}}>SSF Doc ©2021 Created by SJTU Super SoFtware
            Developer
        </Footer>;
    }
}

class HomeView extends React.Component {

    constructor(props) {
        super(props);

    }
    componentDidMount() {
        getUser();
    }

    render() {
        return (
            <Layout>
                <HomeSider/>
                <Layout>
                    <HomeHeader/>
                    <HomeContent/>
                    <HomeFooter/>
                </Layout>
            </Layout>
        )
    }
}

export default withRouter(HomeView);
