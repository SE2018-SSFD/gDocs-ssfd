import React from "react";
import {Col, Input, Layout, Row} from "antd";
import {BellOutlined, MenuOutlined, TabletOutlined} from "@ant-design/icons";
import {UserAvatar} from "./UserAvatar";

const {Header} = Layout
const {Search} = Input
export class HomeHeader extends React.Component{
    render() {
        return(
            <Header className="site-layout-sub-header-background" style={{padding: 0}}>
                <Row align={"middle"} justify={"center"}>
                    <Col span={9} offset={1} >
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
                    <Col>
                        <UserAvatar/>
                    </Col>
                </Row>

            </Header>
        )
    }
}
