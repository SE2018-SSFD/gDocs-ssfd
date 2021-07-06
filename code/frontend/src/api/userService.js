import {postRequest} from "./ajax";
import {history} from '../route/history';
import {message} from 'antd';
import {HTTP_URL} from "./common";

export const login = (data, callback) => {
    const url = HTTP_URL + 'login';
    postRequest(url, data, callback);
};

export const register = (data, callback) => {
    const url = HTTP_URL + 'register';
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
    const data = {
        token: token
    };
    postRequest(url, data, callback);
}

export const modifyUser = (data,callback) => {
    const url = HTTP_URL + 'modifyuser';
    // const callback = (data) => {
    //     let msg_word = MSG_WORDS[data.msg];
    //     if (data.success === true) {
    //         // localStorage.setItem('username',JSON.stringify(data.username));
    //         message.success(msg_word).then(() => {
    //         });
    //     } else {
    //         message.error(msg_word).then(() => {
    //         });
    //     }
    // }
    postRequest(url, data, callback);
}

export const modifyUserAuth = (data,callback) => {
    const url = HTTP_URL + 'modifyuserauth';

    // const callback = (data) => {
    //     let msg_word = MSG_WORDS[data.msg];
    //     if (data.success === true) {
    //         // localStorage.setItem('username',JSON.stringify(data.username));
    //         message.success(msg_word).then(() => {
    //         });
    //     } else {
    //         message.error(msg_word).then(() => {
    //         });
    //     }
    // }
    postRequest(url, data, callback);
}



