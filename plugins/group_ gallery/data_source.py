from configs.config import NICKNAME
from typing import List
from configs.path import IMAGE_PATH, DATA_PATH
from services.log import logger
from utils.utils import cn2py
from utils.http_utils import AsyncHttpx
import os
from .check import check_single

# 处理群友信息获得群友Q号列表
def get_member_list(all_list):
    id_list = []
    for member_list in all_list:
        id_list.append(str(member_list['user_id']))
    return id_list

# 判断是否为QQ
def is_qq(msg: str):
    return msg.isdigit() and 11 >= len(msg) >= 5

async def upload_image_to_local(
    img_list: List[str], path: str, name: str, qq
) -> str:
    img_id = len(os.listdir(path))
    failed_list = []
    success_id = ""
    check_info = ""
    for img_url in img_list:
        if await AsyncHttpx.download_file(img_url, os.path.join(path, f"{img_id}.jpg")):
            check_num = int(check_single(img_id, qq))
            if check_num > -1:
                check_info += f"\n存在与图片：{check_num}相似的图片"
                failed_list.append(img_url)
            else:
                success_id += str(img_id) + "，"
                img_id += 1
        else:
            failed_list.append(img_url)
    failed_result = ""

    for img in failed_list:
        failed_result += str(img) + "\n"

    
    logger.info(
        f" 上传{name}黑历史共 {len(img_list)} 张，失败 {len(failed_list)} 张，id={success_id[:-1]}"
    )
    if failed_list:
        return (
            f"上传{name}黑历史共 {len(img_list) - len(failed_list)} 张\n"
            f"依次的Id为：{success_id[:-1]}\n上传失败：{failed_result[:-1]}"+check_info
        )
    else:
        return (
            f"上传{name}黑历史共 {len(img_list)} 张\nID依次为："
            f"{success_id[:-1]}"
            + check_info
        )
