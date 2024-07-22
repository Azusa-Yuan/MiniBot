import asyncio
import re
import time
from datetime import datetime
from nonebot.plugin import PluginMetadata


from nonebot import on_message, on_request
from nonebot.adapters.onebot.v11 import (
    ActionFailed,
    Bot,
    FriendRequestEvent,
    GroupRequestEvent,
    MessageEvent,
)

from configs.config import NICKNAME
# from models.friend_user import FriendUser
# from models.group_info import GroupInfo
from services.log import logger
from utils.manager import requests_manager
from utils.utils import scheduler

from .utils import time_manager

AUTO_ADD_FRIEND = False

__plugin_meta__ = PluginMetadata(
    name="好友群聊处理请求 [Hidden]",
    description="无",
    usage= "无",
    extra={
        "plugin_version": 0.1,
        "plugin_author": "HibiKier",
        
    },
)


friend_req = on_request(priority=5, block=True)
group_req = on_request(priority=5, block=True)
x = on_message(priority=999, block=False, rule=lambda: False)


@friend_req.handle()
async def _(bot: Bot, event: FriendRequestEvent):
    if time_manager.add_user_request(event.user_id):
        logger.debug(f"收录好友请求...", "好友请求", target=event.user_id)
        user = await bot.get_stranger_info(user_id=event.user_id)
        nickname = user["nickname"]
        sex = user["sex"]
        age = str(user["age"])
        comment = event.comment
        await bot.send_private_msg(
            user_id=int(list(bot.config.superusers)[0]),
            message=f"*****一份好友申请*****\n"
            f"昵称：{nickname}({event.user_id})\n"
            f"自动同意：{'√' if AUTO_ADD_FRIEND else '×'}\n"
            f"日期：{str(datetime.now()).split('.')[0]}\n"
            f"备注：{event.comment}",
        )
        if AUTO_ADD_FRIEND:
            logger.debug(f"已开启好友请求自动同意，成功通过该请求", "好友请求", target=event.user_id)
            await bot.set_friend_add_request(flag=event.flag, approve=True)
        else:
            requests_manager.add_request(
                str(bot.self_id),
                event.user_id,
                "private",
                event.flag,
                nickname=nickname,
                sex=sex,
                age=age,
                comment=comment,
            )
    else:
        logger.debug(f"好友请求五分钟内重复, 已忽略", "好友请求", target=event.user_id)


@group_req.handle()
async def _(bot: Bot, event: GroupRequestEvent):
    # 邀请
    if event.sub_type == "invite":
        if str(event.user_id) in bot.config.superusers:
            try:
                logger.debug(
                    f"超级用户自动同意加入群聊", "群聊请求", event.user_id, target=event.group_id
                )
                await bot.set_group_add_request(
                    flag=event.flag, sub_type="invite", approve=True
                )
            except ActionFailed as e:
                logger.error(
                    "超级用户自动同意加入群聊发生错误",
                    "群聊请求",
                    event.user_id,
                    target=event.group_id,
                    e=e,
                )
        else:
            if time_manager.add_group_request(event.user_id, event.group_id):
                logger.debug(
                    f"收录 用户[{event.user_id}] 群聊[{event.group_id}] 群聊请求", "群聊请求"
                )
                user = await bot.get_stranger_info(user_id=event.user_id)
                nickname = user["nickname"]
                sex = user["sex"]
                age = str(user["age"])
                await bot.send_private_msg(
                    user_id=int(list(bot.config.superusers)[0]),
                    message=f"*****一份入群申请*****\n"
                    f"邀请人：nickname({event.user_id})\n"
                    f"群聊：{event.group_id}\n"
                    f"邀请日期：{datetime.now().replace(microsecond=0)}",
                )
                await bot.send_private_msg(
                    user_id=event.user_id,
                    message=f"想要邀请我偷偷入群嘛~已经提醒{NICKNAME}的管理员大人了\n"
                    "请确保已经群主或群管理沟通过！\n"
                    "等待管理员处理吧！",
                )
                requests_manager.add_request(
                    str(bot.self_id),
                    event.user_id,
                    "group",
                    event.flag,
                    nickname=nickname,
                    invite_group=event.group_id,
                    sex=sex,
                    age=age,
                )
            else:
                logger.debug(
                    f"群聊请求五分钟内重复, 已忽略",
                    "群聊请求",
                    target=f"{event.user_id}:{event.group_id}",
                )


@x.handle()
async def _(event: MessageEvent):
    await asyncio.sleep(0.1)
    r = re.search(r'groupcode="(.*?)"', str(event.get_message()))
    if r:
        group_id = int(r.group(1))
    else:
        return
    r = re.search(r'groupname="(.*?)"', str(event.get_message()))
    if r:
        group_name = r.group(1)
    else:
        group_name = "None"
    requests_manager.set_group_name(group_name, group_id)


@scheduler.scheduled_job(
    "interval",
    minutes=5,
)
async def _():
    time_manager.clear()
