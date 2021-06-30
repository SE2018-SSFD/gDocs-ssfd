const BACKEND_ADDRESS = '123.57.65.161'
const PORTS = ['10086', '10087', '10088']


export const HTTP_URL = 'http://' + BACKEND_ADDRESS + ':' + PORTS[1]+'/';
export const WS_URL = 'ws://' + BACKEND_ADDRESS + ':' + PORTS[1]+'/';

export const MSG_WORDS = [
    "非法格式", "非法口令", "用户未注册", "密码错误", "登陆成功",
    "注册成功", "用户名已存在", "用户信息修改成功", "密码修改成功", "用户名已存在",
    "用户信息获取成功", "新建表格成功", "获取表格成功", "表格无权限访问", "不存在此表格",
    "修改表格成功", "删除表格成功",
]
