import {postRequest} from "./ajax";
import {HTTP_URL,MSG_WORDS} from "./common";
import {history} from "../route/history";
import {message} from "antd";

export const newSheet = () =>{
    const url = HTTP_URL+'newsheet';

    const token = JSON.parse(localStorage.getItem("token"));
    const name = "新建表格";

    const post_data = {
        token: token,
        name: name,
    }

    const callback = (rec_data) => {
        let msg_word = MSG_WORDS[rec_data.msg];
        if (rec_data.success === true) {
            history.push("/sheet?id=" + rec_data.data);
            let sheets = JSON.parse(localStorage.getItem('sheets'));
            sheets.push(
                {
                    fid:rec_data.data,
                    isDeleted:false,
                    name:"新建表格",
                    checkpoints:null,
                    columns:0,
                    content:null,
                }
            )
            localStorage.setItem("sheets",JSON.stringify(sheets))
            message.success(msg_word).then(r => {
            });
        } else {
            message.error(msg_word).then(r => {
            });
        }
    }
    postRequest(url, post_data, callback);
}

// need fid and token
export const getSheet = (data,callback) =>{
    const url = HTTP_URL+'getsheet';
    postRequest(url, data, callback);
}

// need fid and token
export const deleteSheet = (data,callback) =>{
    const url = HTTP_URL+'deletesheet';
    postRequest(url, data, callback);
}


// need token fid chuck
export const getChuck = (data,callback) =>{
    const url = HTTP_URL+'getchunk';
    postRequest(url, data, callback);
}
