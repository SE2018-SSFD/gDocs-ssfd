import {CHANGE_PORT, GET_PORT} from "./common";

const TIMEOUT = 3000;
let getRequest = (url, callback) => {
    let opts = {
        method: "GET",
        mode: 'cors',
        headers: {
            'Content-Type': 'application/json'
        },
        credentials: "include",
        redirect: "follow"
    };

    Promise.race([
        fetch(url, opts)
            .then((response) => {
                return response.json()
            }),
        new Promise(function (resolve, reject) {
            setTimeout(() => reject(new Error('request timeout')), TIMEOUT)
        })])
        .then((data) => {
            callback(data);
        }).catch((error) => {
        // message.error("请求错误,自动更新后端地址").then(() => {
        // })
        console.log(error)
        let port_old = GET_PORT();
        CHANGE_PORT();
        let port_new = GET_PORT();
        let new_url = url.replace(port_old, port_new);
        postRequest(new_url, callback);
    });
};

let postRequest = (url, json, callback) => {
    let opts = {
        method: "POST",
        mode: 'cors',
        body: JSON.stringify(json),
        headers: {
            'Content-Type': 'application/json'
        },
        credentials: "include",
        redirect: "follow"
    };

    Promise.race([
        fetch(url, opts)
            .then((response) => {
                return response.json()
            }),
        new Promise(function (resolve, reject) {
            setTimeout(() => reject(new Error('request timeout')), TIMEOUT)
        })])
        .then((data) => {
            callback(data);
        }).catch((error) => {
        // message.error("请求错误,自动更新后端地址").then(() => {
        // })
        console.log(error)
        let port_old = GET_PORT();
        CHANGE_PORT();
        let port_new = GET_PORT();
        let new_url = url.replace(port_old, port_new);
        postRequest(new_url, json, callback);
    });
};

let postRequestForm = (url, data, callback) => {
    let form = new FormData();

    for (let p in data){
        if(data.hasOwnProperty(p))
            form.append(p, data[p]);
    }

    let opts = {
        method: "POST",
        mode: 'cors',
        body: form,
        credentials: "include",
        redirect: "follow"
    };

    Promise.race([
        fetch(url, opts)
            .then((response) => {
                return response.json()
            }),
        new Promise(function (resolve, reject) {
            setTimeout(() => reject(new Error('request timeout')), TIMEOUT)
        })])
        .then((data) => {
            callback(data);
        }).catch((error) => {
        // message.error("请求错误,自动更新后端地址").then(() => {
        // })
        console.log(error)
        let port_old = GET_PORT();
        CHANGE_PORT();
        let port_new = GET_PORT();
        let new_url = url.replace(port_old, port_new);
        // postRequestForm(new_url, data, callback);
    });
};

export {getRequest, postRequest, postRequestForm};
