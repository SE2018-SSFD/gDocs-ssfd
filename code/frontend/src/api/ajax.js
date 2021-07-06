import {message} from "antd";

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
            // return data;
            callback(data);
        })
        .catch((error) => {
            message.error("请求错误").then(() => {
            })
            console.log(error);
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
            message.error("请求错误").then(() => {
            })
            console.log(error);
        });
};

export {getRequest, postRequest};
