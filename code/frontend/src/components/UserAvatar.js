import React from 'react';
import { Avatar, Dropdown, Menu} from 'antd';
import * as userService from '../api/userService'

export class UserAvatar extends React.Component {

    componentDidMount() {
    }

    render() {
        const username = JSON.parse(localStorage.getItem('username')).toUpperCase()
        const menu = (
            <Menu>
                <Menu.Item>
                    <div>
                        详细信息
                    </div>
                </Menu.Item>
                <Menu.Item>
                    <div onClick={userService.logout}>
                        Log Out
                    </div>
                </Menu.Item>
            </Menu>
        );

        return(
            <div id="avatar">
                <Dropdown overlay={menu} placement="bottomRight">
                    <Avatar shape="square" style={{cursor:"pointer"}}>
                        {username}
                    </Avatar>
                </Dropdown>
            </div>
        );
    }
}
