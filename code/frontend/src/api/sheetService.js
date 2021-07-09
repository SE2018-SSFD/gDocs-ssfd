import {getRequest, postRequest} from "./ajax";
import {HTTP_URL} from "./common";

export const newSheet = (callback) => {
    const url = HTTP_URL + 'newsheet';
    const token = JSON.parse(localStorage.getItem("token"));
    const name = "new sheet";

    const data = {
        token: token,
        name: name,
        initColumns: 60,
        intiRows: 84,
    }
    postRequest(url, data, callback);
}

export const getSheet = (fid, callback) => {
    const url = HTTP_URL + 'getsheet';
    const token = JSON.parse(localStorage.getItem("token"));
    let data = {
        fid: fid,
        token: token
    }
    postRequest(url, data, callback);
}

export const deleteSheet = (fid, callback) => {
    const url = HTTP_URL + 'deletesheet';
    const token = JSON.parse(localStorage.getItem("token"));
    let data = {
        fid: fid,
        token: token
    }
    postRequest(url, data, callback);
}

export const recoverSheet = (fid, callback) => {
    const url = HTTP_URL + 'recoversheet';
    const token = JSON.parse(localStorage.getItem("token"));
    let data = {
        fid: fid,
        token: token
    }
    postRequest(url, data, callback);
}

export const commitSheet = (fid, callback) => {
    const url = HTTP_URL + 'commitsheet';
    const token = JSON.parse(localStorage.getItem("token"));
    let data = {
        fid: fid,
        token: token
    }
    postRequest(url, data, callback);
}

// need token fid chuck
export const getChuck = (data, callback) => {
    const url = HTTP_URL + 'getchunk';
    postRequest(url, data, callback);
}

export const getSheetCkpt = (data,callback) =>{
    const url = HTTP_URL + 'getsheetchkp';
    postRequest(url, data, callback);
}

export const getSheetLog = (fid,lid,callback) =>{
    const token = JSON.parse(localStorage.getItem("token"));
    const data = {
        token: token,
        fid: fid,
        lid: lid
    }
    const url = HTTP_URL + 'getsheetlog';
    postRequest(url, data, callback);
}

export const testWS = (fid, callback) => {
    const token = JSON.parse(localStorage.getItem("token"));
    const url = HTTP_URL + 'sheetws?token=' + token + "&fid=" + fid + "&query=1";
    getRequest(url, callback)
}

export const rollbackSheet = (fid, cid, callback) => {
    const url = HTTP_URL + 'rollbacksheet';
    const token = JSON.parse(localStorage.getItem("token"));
    let data = {
        fid: fid,
        token: token,
        cid:cid,
    }
    postRequest(url, data, callback);
}
