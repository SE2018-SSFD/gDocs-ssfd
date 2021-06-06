import {postRequest} from "../utils/ajax";
import {apiUrl} from "../config";

export const newSheet = (data,callback) =>{
    const url = apiUrl+'newsheet';
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
