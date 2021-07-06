import React from 'react';
import {withRouter} from "react-router-dom";
import '../css/home.css'
import logo from '../assets/logo.png'

import {Button, Col, Image, Input, Layout, Menu, message, Row, Space} from 'antd';
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
import {MSG_WORDS} from "../api/common";
import {getUser} from "../api/userService";
import {history} from "../route/history";

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

    constructor(props) {
        super(props);
        this.state = {
            current: 'MainPage',
        }
        this.menuCallback = this.props.menuCallback;
    }

    handleClick = (e) => {
        this.menuCallback(e.key);
        this.setState({current: e.key});
    }

    handleNew() {


        const callback = (data) => {
            console.log(data);
            let msg_word = MSG_WORDS[data.msg];
            if (data.success === true) {
                history.push("/sheet?id=" + data.data);
                let sheets = JSON.parse(localStorage.getItem('sheets'));
                sheets.push(
                    {
                        fid: data.data,
                        isDeleted: false,
                        name: "新建表格",
                        checkpoints: null,
                        columns: 0,
                        content: null,
                    }
                )
                localStorage.setItem("sheets", JSON.stringify(sheets))
                message.success(msg_word).then(r => {
                });
            } else {
                message.error(msg_word).then(r => {
                });
            }
        };

        newSheet(callback)
    }

    render() {
        const {current} = this.state;
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
                        <Button size="large" type="primary" icon={<PlusOutlined/>} onClick={this.handleNew}
                                block={true}>新建</Button>
                        <Button size="large" icon={<UploadOutlined/>} block={true}>导入本地文件</Button>
                    </Space>
                </Col>
            </Row>

            <Menu mode="inline" onClick={this.handleClick} selectedKeys={[current]}>
                <Menu.Item key="MainPage" icon={<HomeOutlined/>}>
                    首页
                </Menu.Item>
                <SubMenu key="MyDoc" icon={<FileTextOutlined/>} title="我的文档">
                    <Menu.Item key="ShareMe" icon={<FolderOutlined/>}>与我共享</Menu.Item>
                    <Menu.Item key="Introduce" icon={<FolderOutlined/>}>Hi, 欢迎使用SSFDoc</Menu.Item>
                </SubMenu>
                {/*<hr size="3px" color="#E6E7E8"/>*/}
                <Menu.Item key="Template" icon={<AppstoreAddOutlined/>}>模板</Menu.Item>
                <Menu.Item key="Recycle" icon={<DeleteOutlined/>}>回收站</Menu.Item>
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
                <FileList content={this.props.content} type={"main"}/>
            </div>
        </Content>;
    }
}

class RecycleContent extends React.Component {
    render() {
        return <Content style={{margin: "24px 16px 0"}}>
            <div>
                <h1>
                    回收站
                </h1>
            </div>
            <div className="site-layout-background" style={{padding: 24, minHeight: 360}}>
                <FileList content={this.props.content} type={"recycle"}/>
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
        this.state = {
            curSection: 0,
            file: [],
            deletedFile: [],
        }
    }

    componentDidMount() {
        const callback = (data) => {
            let msg_word = MSG_WORDS[data.msg];
            if (data.success === true) {
                localStorage.setItem('sheets', JSON.stringify(data.data.sheets));
                localStorage.setItem('username', JSON.stringify(data.data.username));
            } else {
                message.error(msg_word).then(() => {
                });
            }
            let sheets = data.data.sheets;
            let file = [];
            let deletedFile = [];
            sheets.forEach((v, i) => {
                if (v.isDeleted === true) {
                    deletedFile.push(v);
                } else {
                    file.push(v);
                }
            })
            this.setState({
                file: file,
                deletedFile: deletedFile
            })

        }
        getUser(callback);
    }

    menuCallback = (key) => {
        switch (key) {
            case 'MainPage':
                this.setState({curSection: 0});
                break;
            case 'Recycle':
                this.setState({curSection: 1});
                break;
            default:
                // console.error("Not a valid key");
                this.setState({curSection: 0});
                break;
        }
    };

    render() {
        const {curSection, file, deletedFile} = this.state;
        const content =
            curSection === 0 ? (
                <HomeContent content={file}/>
            ) : curSection === 1 ? (
                <RecycleContent content={deletedFile}/>
            ) : (
                <></>
            );

        return (
            <Layout>
                <HomeSider menuCallback={this.menuCallback}/>
                <Layout>
                    <HomeHeader/>
                    {content}
                    <HomeFooter/>
                </Layout>
            </Layout>
        )
    }
}

export default withRouter(HomeView);
