import {postRequest} from "../utils/ajax";
import {history} from '../utils/history';
import {message} from 'antd';
import {apiUrl} from "../utils/config";

const login_register_callback = (data) => {
    if (data.status >= 0) {
        localStorage.setItem('token', JSON.stringify(data.data));
        history.push("/");
        message.success(data.msg).then(r => {
        });
    } else {
        message.error(data.msg).then(r => {
        });
    }
};

export const login = (data) => {
    const url = apiUrl + 'login';
    postRequest(url, data, login_register_callback(data));
};

export const register = (data) => {
    const url = apiUrl + 'register';
    postRequest(url, data, login_register_callback(data));
};

export const logout = () => {
    const url = apiUrl + 'logout';

    const callback = (data) => {
        if (data.status >= 0) {
            localStorage.removeItem("user");
            history.push("/login");
            message.success(data.msg).then(r => {});
        } else {
            message.error(data.msg).then(r => {});
        }
    };
    postRequest(url, {}, callback);
};

export const checkSession = (callback) => {
    const url = apiUrl+'checkSession';
    postRequest(url, {}, callback);
};

export const getUser = (data,callback) =>{
    const url = apiUrl+'getuser';
    postRequest(url, data, callback);
}

export const modifyUser = (data,callback) =>{
    const url = apiUrl+'modifyuser';
    postRequest(url, data, callback);
}

export const modifyUserAuth = (data,callback) =>{
    const url = apiUrl+'modifyuserauth';
    postRequest(url, data, callback);
}



