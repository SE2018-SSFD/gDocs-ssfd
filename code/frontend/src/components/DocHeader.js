import React from "react";
import {Button, Col, Divider, Image, Layout, Row} from "antd";
import {
    FolderOutlined,
    LeftOutlined,
    MenuOutlined,
    StarOutlined,
    CheckCircleOutlined,
    EditOutlined
} from "@ant-design/icons";
import {UserAvatar} from "./UserAvatar";
import docs from "../assets/docs.png";
import {Link} from "react-router-dom";

const {Header} = Layout

export class DocHeader extends React.Component {
    render() {
        return (
            <Header className="site-layout-sub-header-background" style={{padding: 0}}>
                <Row align={"middle"} justify={"center"}>
                    <Col span={2}>
                    <Link to={{
                        pathname: '/',
                    }} target="_blank"
                    >
                        <LeftOutlined/>
                        <Image src={docs} alt={'docs'} height={50} width={50} preview={false}/>
                    </Link>
                    </Col>
                    <Col span={1}>
                        <StarOutlined/>
                    </Col>
                    <Col span={1}>
                        <FolderOutlined/>
                    </Col>
                    <Col span={1}>
                        <CheckCircleOutlined />
                    </Col>
                    <Col span={1} offset={10}>
                        <MenuOutlined/>
                    </Col>
                    <Col span={1}>
                        <EditOutlined/>
                    </Col>
                    <Col span={1}>
                        <EditOutlined/>
                    </Col>
                    <Divider type={"vertical"}/>
                    <Col span={1}>
                        <Button type={'primary'}> 分享</Button>
                    </Col>
                    <Divider type={"vertical"}/>
                    <Col span={1}>
                        <UserAvatar/>
                    </Col>

                </Row>
            </Header>
        )
    }
}
