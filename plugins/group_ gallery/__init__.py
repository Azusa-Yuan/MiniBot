from nonebot import on_command
from nonebot.rule import to_me
from nonebot.typing import T_State
from nonebot.plugin import PluginMetadata
from nonebot.adapters.onebot.v11 import Bot, MessageEvent, GroupMessageEvent, Message, MessageSegment
from utils.utils import get_message_img

from .check import *
from .data_source import *

from nonebot.params import CommandArg, Arg, ArgStr
from typing import List
from nonebot.permission import SUPERUSER

from nonebot.adapters.onebot.v11.permission import (
    GROUP_ADMIN,
    GROUP_OWNER,
)
from pil_utils import BuildImage, Text2Image
from utils.message_builder import image
from configs.path import IMAGE_PATH, DATA_PATH
import shlex
import os
import random
IMAGE_PATH = os.path.join(IMAGE_PATH, "Black_history")

__plugin_usage__ = """
usage：
    上传图库（仅管理员可用）
    指令：
        上传图库 @群友 [图片]
    * 支持一次性上传多张图片 *
    发送随机图库图片或者指定群友的随机图库图片或特定图库图片
    指令：
        随机图库 
        图库 @群友
        图库 @群友 图片ID
    图库图片数量统计
    指令：
        图库统计
""".strip()


__plugin_meta__ = PluginMetadata(
    name="图库",
    description="无",
    usage= __plugin_usage__
)


PERM_EDIT = GROUP_ADMIN | GROUP_OWNER | SUPERUSER
PERM_GLOBAL = SUPERUSER

upload_img = on_command("上传图库", aliases={"上传黑历史"}, priority=5, block=True, permission=PERM_EDIT)
show_list = on_command("图库统计", aliases={"黑历史统计"}, priority=5, block=True)
send_img = on_command("随机图库", aliases={"黑历史"},priority=5, block=True)
delete_img = on_command("删除图库", aliases={"删除黑历史"}, priority=5, block=True, permission=SUPERUSER)
check_img = on_command("黑历史查重", aliases={"图库查重"}, priority=5, block=True, permission=SUPERUSER)


def get_member_list(all_list):
    id_list = []
    for member_list in all_list:
        id_list.append(str(member_list['user_id']))
    return id_list


@upload_img.handle()
async def _(event: MessageEvent, state: T_State, arg: Message = CommandArg()):
    msg: Message = state["_prefix"]["command_arg"]
    # 获取黑历史的QQ号
    for msg_seg in msg:
        if msg_seg.type == "at":
            state["qq"] = str(msg_seg.data['qq'])
            break
        elif msg_seg.type == "text":
            raw_text = str(msg_seg.data['text'])
            try:
                texts = shlex.split(raw_text)
            except:
                texts = raw_text.split()
            for text in texts:
                if is_qq(text):
                    state["qq"] = text

    img_list = get_message_img(event.json())
    if img_list:
        state["img_list"] = arg


@upload_img.got(
    "qq",
    prompt=f"请选择要上传的对象\n- ",
)
@upload_img.got("img_list", prompt="图呢图呢图呢图呢！GKD！")
async def _(
    bot: Bot,
    event: MessageEvent,
    state: T_State,
    qq: str = ArgStr("qq"),
    img_list: Message = Arg("img_list"),
):
    if not get_message_img(img_list):
        await upload_img.reject_arg("img_list", "图呢图呢图呢图呢！GKD！")

    save_path = os.path.join(IMAGE_PATH, qq)
    if not os.path.exists(save_path):
        os.makedirs(save_path)

    img_list = get_message_img(img_list)
    
    # 获取群昵称或者QQ昵称
    if isinstance(event, GroupMessageEvent):
        group_id = event.group_id
        info = await bot.get_group_member_info(
            group_id=int(group_id), user_id=int(qq)
        )
        name = info.get("card", "") or info.get("nickname", "")
    else:
        info = await bot.get_stranger_info(user_id=int(qq))
        name = info.get("nickname", "")

    await upload_img.send(
        await upload_image_to_local(img_list, save_path, name, qq)
    )


@delete_img.handle()
async def _(event: MessageEvent, state: T_State, arg: Message = CommandArg()):
    qq = False
    ID = False
    msg: Message = state["_prefix"]["command_arg"]
    # 获取黑历史的QQ号
    for msg_seg in msg:
        if msg_seg.type == "at":
            qq = str(msg_seg.data['qq'])
            state["qq"] = qq

        elif msg_seg.type == "text":
            raw_text = str(msg_seg)
            try:
                texts = shlex.split(raw_text)
            except:
                texts = raw_text.split()
            for text in texts:
                if is_qq(text):
                    qq = text
                    state["qq"] = qq
                elif qq and text.isdigit():
                    state["ID"] = text
@delete_img.got(
    "qq",
    prompt=f"请选择要删除的对象\n ",
)
@delete_img.got("ID", prompt="ID呢！GKD！")
async def _(
        bot: Bot,
        event: MessageEvent,
        state: T_State,
        qq: str = ArgStr("qq"),
        ID = Arg("ID"),
):
    if delhash(ID, qq):
        await delete_img.send("删除hash成功")
    file_path = os.path.join(IMAGE_PATH, qq)
    ID_max = len(os.listdir(file_path)) - 1
    if int(ID) > ID_max:
        await delete_img.finish("ID超出上限")
    os.remove(os.path.join(file_path, ID+".jpg"))
    if int(ID_max) != int(ID):
        os.rename(os.path.join(file_path, str(ID_max)+".jpg"), os.path.join(file_path, ID+".jpg"))

    await delete_img.finish("删除成功")

@send_img.handle()
async def _(bot: Bot, event: MessageEvent, state: T_State, arg: Message = CommandArg()):
    # 获取黑历史的QQ号
    qq = False
    ID = False

    msg: Message = state["_prefix"]["command_arg"]
    for msg_seg in msg:
        if msg_seg.type == "at":
            qq = str(msg_seg.data['qq'])

        elif msg_seg.type == "text":
            raw_text = str(msg_seg)
            try:
                texts = shlex.split(raw_text)
            except:
                texts = raw_text.split()
            for text in texts:
                if is_qq(text):
                    print(bot.config.superusers)
                    if str(event.user_id) in bot.config.superusers:
                        qq = text
                elif qq and text.isdigit():
                    ID = text

    if not qq:
        if isinstance(event, GroupMessageEvent):
            group_id = event.group_id
            # 获取群友q号
            all_list = await bot.get_group_member_list(group_id=group_id)
            id_list = get_member_list(all_list)
            # 获取文件夹名字列表并求交集
            file_list = os.listdir(IMAGE_PATH)
            qq_set = set(id_list).intersection(file_list)
            qq = random.choice(list(qq_set))
        else:
            await send_img.finish("只能看群友的黑历史哦")

    # 获取群昵称或者QQ昵称
    if isinstance(event, GroupMessageEvent):
        try:
            group_id = event.group_id
            info = await bot.get_group_member_info(
            group_id=int(group_id), user_id=int(qq) )
            # print(info)
            name = info.get("card", "") or info.get("nickname", "")
        except Exception as e:
            info = await bot.get_stranger_info(user_id=int(qq))
            name = info.get("nickname", "")
    else:
        info = await bot.get_stranger_info(user_id=int(qq))
        name = info.get("nickname", "")

    file_path = os.path.join(IMAGE_PATH, qq)
    if not os.path.exists(file_path):
        await send_img.finish("没有该群友的黑历史")
    
    ID_max = len(os.listdir(file_path)) - 1
    if not ID:
        ID = str(random.randint(0, ID_max))
    
    if int(ID) > ID_max:
        await send_img.finish("ID超出上限")

    file_path = os.path.join("Black_history", qq, ID+".jpg")
    await send_img.finish(image(file_path)+f"\n{name}的黑历史 图片ID:{int(ID)}")

@show_list.handle()
async def _(bot: Bot, event: MessageEvent, state: T_State, arg: Message = CommandArg()):
    file_list = os.listdir(IMAGE_PATH)
    if isinstance(event, GroupMessageEvent):
        group_id = event.group_id
        # 获取群友q号
        all_list = await bot.get_group_member_list(group_id=group_id)
        id_list = get_member_list(all_list)
        # 获取文件夹名字列表并求交集
        qq_set = set(id_list).intersection(file_list)
        qq_list = list(qq_set)
    else:
        if str(event.user_id) in bot.config.superusers:
            qq_list = file_list
        else:
            return

    black_list = []
    for qq in qq_list:
        if isinstance(event, GroupMessageEvent):
            group_id = event.group_id
            info = await bot.get_group_member_info(
                group_id=int(group_id), user_id=int(qq)
            )
            name = info.get("card", "") or info.get("nickname", "")
        else:
            info = await bot.get_stranger_info(user_id=int(qq))
            name = info.get("nickname", "")

        file_path = os.path.join(IMAGE_PATH, qq)
        black_list.append([name, len(os.listdir(file_path))])

    imgs = []
    names = []
    nums = []
    black_list = sorted(black_list, key=lambda x:x[1], reverse=True)

    head_text = "黑历史统计"
    head = Text2Image.from_text(head_text, 30, weight="bold").to_image(padding=(20, 10))
    for i, j in black_list:
        names.append(i)
        nums.append(str(j))

    names = "\n".join(names)
    nums = "\n".join(nums)
    imgs.append(Text2Image.from_bbcode_text(names, 30).to_image(padding=(20, 10)))
    imgs.append(Text2Image.from_bbcode_text(nums, 30).to_image(padding=(20, 10)))
    w = sum((img.width for img in imgs))
    h = head.height + max((img.height for img in imgs))
    frame = BuildImage.new("RGBA", (w, h), "white")
    frame.paste(head, alpha=True)
    current_w = 0
    for img in imgs:
        frame.paste(img, (current_w, head.height), alpha=True)
        current_w += img.width

    await show_list.finish(MessageSegment.image(frame.save_jpg()))

@check_img.handle()
async def _(bot: Bot, event: MessageEvent, state: T_State, arg: Message = CommandArg()):
    check_all()
