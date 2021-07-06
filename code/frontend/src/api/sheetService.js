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

export const getSheet = (url, data, callback) => {
    postRequest(url, data, callback);
}

// need fid and token
export const deleteSheet = (fid, callback) => {
    const urlCallback = (url) => {
        let myurl = url + "deletesheet";
        const token = JSON.parse(localStorage.getItem("token"));
        let data = {
            fid: fid,
            token: token
        }
        postRequest(myurl, data, callback);
    }
    getURL(fid, urlCallback);
}


// need token fid chuck
export const getChuck = (data, callback) => {
    const url = HTTP_URL + 'getchunk';
    postRequest(url, data, callback);
}

export const testWS = (fid, callback) => {
    const token = JSON.parse(localStorage.getItem("token"));
    const url = HTTP_URL + 'sheetws?token=' + token + "&fid=" + fid + "&query=1";
    getRequest(url, callback)
}

export const getURL = (fid, callback1) => {
    const callback = (data) => {
        if (data.success === false) {
            callback1("http" + data.data.slice(2, 25));
        } else {
            callback1(HTTP_URL);
        }
    }
    testWS(fid, callback)
}
