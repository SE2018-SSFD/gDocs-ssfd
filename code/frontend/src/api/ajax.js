import {message} from "antd";

let postRequest = (url, json, callback) => {
    let opts = {
        method: "POST",
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
            message.error("请检查您的网络连接").then(() => {})
            console.log(error);
        });
};

let getRequest = (url,callback) => {
    let opts = {
        method: "GET",
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
            message.error("请检查您的网络连接").then(() => {})
            console.log(error);
        });
};


export {getRequest,postRequest};
