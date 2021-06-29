import {postRequest} from "../utils/ajax";
import {history} from '../utils/history';
import {message} from 'antd';
import {apiUrl} from "../utils/config";
import {MSGWORD} from "./common";

export const login = (data) => {
    const url = apiUrl + 'login';
    const callback = (data) =>{
        let msg_word = MSGWORD[data.msg];
        if (data.success === true) {
            localStorage.setItem('uid', JSON.stringify(data.data.info.uid));
            localStorage.setItem('sheets',JSON.stringify(data.data.info.sheets))
            localStorage.setItem('username',JSON.stringify(data.data.info.username));
            localStorage.setItem('token',JSON.stringify(data.data.token));
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
    const url = apiUrl + 'register';
    const callback = (data) => {
        let msg_word = MSGWORD[data.msg];
        if (data.success === true) {
            localStorage.setItem('sheets',JSON.stringify(data.data.info.sheets))
            localStorage.setItem('username',JSON.stringify(this.state.username));
            localStorage.setItem('token',JSON.stringify(data.data));
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
    //const url = apiUrl + 'logout';
    localStorage.removeItem("token");
    history.push("/login");
    message.success("登出成功！").then(() => {});
};

export const checkSession = (callback) => {
    const url = apiUrl+'getuser';
    const token = JSON.parse(localStorage.getItem('token'));
    const data = {
        token:token
    };
    postRequest(url, data, callback);
};

export const getUser = () =>{
    const url = apiUrl+'getuser';
    const token = JSON.parse(localStorage.getItem('token'));
    const data = {
        token:token
    };
    const callback = (data) =>{
        let msg_word = MSGWORD[data.msg];
        if (data.success === true) {
            localStorage.setItem('sheets',JSON.stringify(data.data.info.sheets))
            localStorage.setItem('username',JSON.stringify(this.state.username));
            localStorage.setItem('token',JSON.stringify(data.data));
        } else {
            message.error(msg_word).then(() => {
            });
        }
    }
    postRequest(url, data, callback);
}

export const modifyUser = (data) =>{
    const url = apiUrl+'modifyuser';
    const callback = (data) =>{
        let msg_word = MSGWORD[data.msg];
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

export const modifyUserAuth = (data) =>{
    const url = apiUrl+'modifyuserauth';

    const callback = (data) =>{
        let msg_word = MSGWORD[data.msg];
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



