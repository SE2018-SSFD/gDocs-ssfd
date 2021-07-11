export function arrTrans(num, arr) {
    console.log(num,arr)
    const iconsArr = []; // 声明数组
    arr.forEach((item, index) => {
        const page = Math.floor(index / num); // 计算该元素为第几个素组内
        if (!iconsArr[page]) { // 判断是否存在
            iconsArr[page] = [];
        }
        iconsArr[page].push(item);
    });
    return iconsArr;
}

export function RowMap(row){
    return row + 1;
}
export function ColMap(col){
    let first_idx = Math.floor(col/26);
    let second_idx = col % 26;
    if(first_idx === 0)
    {
        return String.fromCharCode(65+ second_idx);
    }else{
        return String.fromCharCode(65+ first_idx) + String.fromCharCode(65+ second_idx);
    }
}
