import {postRequest} from "../utils/ajax";
import {apiUrl} from "../utils/config";
import {MSGWORD} from "./common";
import {history} from "../utils/history";
import {message} from "antd";

export const newSheet = (data) =>{
    const url = apiUrl+'newsheet';

    const callback = (data) => {
        let msg_word = MSGWORD[data.msg];

        if (data.success === true) {
            history.push("/doc?id=" + data.data);
            message.success(msg_word).then(r => {
            });
        } else {
            message.error(msg_word).then(r => {
            });
        }
    }
    postRequest(url, data, callback);
}

export const getSheet = (data,callback) =>{
    const url = apiUrl+'getsheet';
    postRequest(url, data, callback);
}

export const modifySheet = (data,callback) =>{
    const url = apiUrl+'modifysheet';
    postRequest(url, data, callback);
}

export const deleteSheet = (data,callback) =>{
    const url = apiUrl+'deletesheet';
    postRequest(url, data, callback);
}

export const commitSheet = (data,callback) =>{
    const url = apiUrl+'commitsheet';
    postRequest(url, data, callback);
}

export const getChuck = (data,callback) =>{
    const url = apiUrl+'getchunk';
    postRequest(url, data, callback);
}
