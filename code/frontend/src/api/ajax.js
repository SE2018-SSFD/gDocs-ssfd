import {message} from "antd";
import {CHANGE_PORT} from "./common";

let getRequest = (url, callback) => {
    let opts = {
        method: "GET",
        mode:'cors',
        headers: {
            'Content-Type': 'application/json'
        },
        credentials: "include"
    };

    fetch(url, opts)
        .then((response) => {
            return response.json()
        })
        .then((data) => {
            callback(data);
        })
        .catch((error) => {
            message.error("请求错误,已更换地址，请重试").then(() => {
            })
            console.log(error);
            CHANGE_PORT();
        });
};

let postRequest = (url, json, callback) => {
    let opts = {
        method: "POST",
        mode:'cors',
        body: JSON.stringify(json),
        headers: {
            'Content-Type': 'application/json'
        },
        credentials: "include"
    };

    fetch(url, opts)
        .then((response) => {
            return response.json()
        })
        .then((data) => {
            callback(data);
        })
        .catch((error) => {
            message.error("请求错误,已更换地址，请重试").then(() => {
            })
            console.log(error);
            CHANGE_PORT();
        });
};

export {getRequest, postRequest};
