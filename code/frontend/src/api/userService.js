import {postRequest} from "./ajax";
import {history} from '../route/history';
import {message} from 'antd';
import {HTTP_URL, MSG_WORDS} from "./common";

export const login = (data) => {
    const url = HTTP_URL + 'login';
    const callback = (data) => {
        let msg_word = MSG_WORDS[data.msg];
        if (data.success === true) {
            localStorage.setItem('uid', JSON.stringify(data.data.info.uid));
            localStorage.setItem('sheets', JSON.stringify(data.data.info.sheets))
            localStorage.setItem('username', JSON.stringify(data.data.info.username));
            localStorage.setItem('token', JSON.stringify(data.data.token));
            history.push("/");
            message.success(msg_word).then(() => {
            });
        } else {
            message.error(msg_word).then(() => {
            });
        }
    }
    postRequest(url, data, callback);
};

export const register = (data) => {
    const url = HTTP_URL + 'register';
    const callback = (data) => {
        let msg_word = MSG_WORDS[data.msg];
        if (data.success === true) {
            localStorage.setItem('sheets', JSON.stringify(data.data.info.sheets))
            localStorage.setItem('username', JSON.stringify(this.state.username));
            localStorage.setItem('token', JSON.stringify(data.data));
            history.push("/");
            message.success(msg_word).then(() => {
            });
        } else {
            message.error(msg_word).then(() => {
            });
        }
    }
    postRequest(url, data, callback);
};

export const logout = () => {
    localStorage.removeItem("token");
    history.push("/login");
    message.success("登出成功！").then(() => {
    });
};

export const checkSession = (callback) => {
    const url = HTTP_URL + 'getuser';
    const token = JSON.parse(localStorage.getItem("token"));
    if (token === null) {
        const data = {
            success: false,
        }
        callback(data)
    } else {
        const data = {
            token: token
        };
        postRequest(url, data, callback);
    }
};

export const getUser = (callback) => {
    const url = HTTP_URL + 'getuser';
    const token = JSON.parse(localStorage.getItem('token'));
    const post_data = {
        token: token
    };
    postRequest(url, post_data, callback);
}

export const modifyUser = (data) => {
    const url = HTTP_URL + 'modifyuser';
    const callback = (data) => {
        let msg_word = MSG_WORDS[data.msg];
        if (data.success === true) {
            // localStorage.setItem('username',JSON.stringify(data.username));
            message.success(msg_word).then(() => {
            });
        } else {
            message.error(msg_word).then(() => {
            });
        }
    }
    postRequest(url, data, callback);
}

export const modifyUserAuth = (data) => {
    const url = HTTP_URL + 'modifyuserauth';

    const callback = (data) => {
        let msg_word = MSG_WORDS[data.msg];
        if (data.success === true) {
            // localStorage.setItem('username',JSON.stringify(data.username));
            message.success(msg_word).then(() => {
            });
        } else {
            message.error(msg_word).then(() => {
            });
        }
    }
    postRequest(url, data, callback);
}



