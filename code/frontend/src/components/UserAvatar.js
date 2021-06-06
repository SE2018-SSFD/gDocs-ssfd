import React from 'react';
import { Avatar, Dropdown, Menu} from 'antd';
import * as userService from '../services/userService'

export class UserAvatar extends React.Component {

    render() {
        const menu = (
            <Menu>
                <Menu.Item>
                    <a target="_blank" rel="noopener noreferrer" href="http://www.alipay.com/">
                        详细信息
                    </a>
                </Menu.Item>
                <Menu.Item>
                    <a href="#" onClick={userService.logout}>
                        Log Out
                    </a>
                </Menu.Item>
            </Menu>
        );

        const {user} = this.props;

        return(
            <div id="avatar">
                <Dropdown overlay={menu} placement="bottomRight">
                    <Avatar shape="square" style={{cursor:"pointer"}}>
                        LHN
                    </Avatar>
                </Dropdown>
            </div>
        );
    }
}
