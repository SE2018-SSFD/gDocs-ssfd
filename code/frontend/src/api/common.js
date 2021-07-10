const BACKEND_ADDRESS = '123.57.65.161'
const PORTS = ['10086', '10087', '10088']

export function GET_PORT() {
    let PORT_COUNT = localStorage.getItem("PORT_COUNT");
    if (PORT_COUNT === null) {
        let PORT_COUNT = 0;
        localStorage.setItem("PORT_COUNT", JSON.stringify(PORT_COUNT));
        return PORTS[0];
    } else {
        PORT_COUNT = JSON.parse(PORT_COUNT);

        return PORTS[PORT_COUNT];
    }

}

export function CHANGE_PORT() {
    let PORT_COUNT = JSON.parse(localStorage.getItem("PORT_COUNT"));
    PORT_COUNT = (PORT_COUNT + 1) % PORTS.length;
    localStorage.setItem("PORT_COUNT", JSON.stringify(PORT_COUNT));
}

export function GET_HTTP_URL() {
    const PORT = GET_PORT();
    return 'http://' + BACKEND_ADDRESS + ':' + PORT + '/';
}

export function GET_WS_URL() {
    const PORT = GET_PORT();
    return 'ws://' + BACKEND_ADDRESS + ':' + PORT + '/';
}

export const MSG_WORDS = [
    "非法请求格式", "口令过期，请重新登陆", "用户未注册", "密码错误", "登陆成功", "注册成功",
    "用户名已存在", "用户信息修改成功", "密码修改成功", "用户名已存在", "用户信息获取成功",
    "新建表格成功", "获取表格成功", "表格无权限访问", "不存在此表格", "修改表格成功",
    "删除表格成功", "表格重复连接", "表格已删除", "WS重定向", "WS地址获取成功",
    "恢复存档成功", "检查点不存在", "获取编辑记录成功", "编辑记录不存在", "恢复文件成功",
    "文件已恢复", "未打开sheet", "保存成功", "你还没有修改文档", "文件回滚成功",
    "不能获取文件","错误表格数据","上传文件成功","获取文件成功"
]
