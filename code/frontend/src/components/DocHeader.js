// import React from "react";
// import {Button, Row,Col, Divider, Image, Layout } from "antd";
// import {
//     FolderOutlined,
//     LeftOutlined,
//     MenuOutlined,
//     StarOutlined,
//     CheckCircleOutlined,
//     EditOutlined
// } from "@ant-design/icons";
// import {UserAvatar} from "./UserAvatar";
// import docs from "../assets/tencent_doc_word.png";
// import {Link} from "react-router-dom";
//
// const {Header} = Layout
//
// export class DocHeader extends React.Component {
//
//     render() {
//         return (
//             <Header className="site-layout-sub-header-background" style={{padding: 0}}>
//                 <Row justify={"center"}>
//                     <Col span={2} offset={1}>
//                     <Link to={{
//                         pathname: '/',
//                     }}
//                     >
//                         <LeftOutlined/>
//                         <Image src={docs} alt={'docs'} height={50} width={50} preview={false}/>
//                     </Link>
//                     </Col>
//                     <Col span={1}>
//                         <StarOutlined/>
//                     </Col>
//                     <Col span={1}>
//                         <FolderOutlined/>
//                     </Col>
//                     <Col span={1}>
//                         <CheckCircleOutlined />
//                     </Col>
//                     <Col span={4}>
//                         <h1>{this.props.data.name}</h1>
//                     </Col>
//                     <Col span={1} offset={13}>
//                         <MenuOutlined/>
//                     </Col>
//                     <Col span={1}>
//                         <EditOutlined/>
//                     </Col>
//                     <Divider type={"vertical"}/>
//                     <Col span={1}>
//                         <Button type={'primary'}> 分享</Button>
//                     </Col>
//                     <Divider type={"vertical"}/>
//                     <Col span={1}>
//                         <UserAvatar/>
//                     </Col>
//                 </Row>
//             </Header>
//         )
//     }
// }
