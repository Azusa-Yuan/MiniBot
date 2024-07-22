import os
import random
from datetime import datetime
from pathlib import Path
from nonebot.plugin import PluginMetadata

import ujson as json
from nonebot import on_notice, on_request
from nonebot.adapters.onebot.v11 import (
    ActionFailed,
    Bot,
    GroupDecreaseNoticeEvent,
    GroupIncreaseNoticeEvent,
)

from configs.config import NICKNAME, config
from services.log import logger
from utils.manager import group_manager, plugins2settings_manager, requests_manager
from utils.message_builder import image


__plugin_meta__ = PluginMetadata(
    name="群事件处理 [Hidden]",
    description="群事件处理",
    usage= "无",
    extra={
        "plugin_version": 0.1,
        "plugin_author": "HibiKier"
    },
)


# 群员增加处理
group_increase_handle = on_notice(priority=1, block=False)
# 群员减少处理
group_decrease_handle = on_notice(priority=1, block=False)
# （群管理）加群同意请求
add_group = on_request(priority=1, block=False)


@group_increase_handle.handle()
async def _(bot: Bot, event: GroupIncreaseNoticeEvent):
    agree = False
    if event.user_id == int(bot.self_id):
        if str(event.operator_id) in bot.config.superusers:
            agree = True
        if config["invite_manager"]["flag"]:
            agree = True
        if not agree:
            try:
                await bot.set_group_leave(group_id=event.group_id)
                await bot.send_private_msg(
                    user_id=int(list(bot.config.superusers)[0]),
                    message=f"触发强制入群保护，已成功退出群聊 {event.group_id}...",
                )
                logger.info(f"强制拉群或未有群信息，退出群聊成功", "入群检测", group_id=event.group_id)
                requests_manager.remove_request("group", event.group_id)
            except Exception as e:
                logger.info(f"强制拉群或未有群信息，退出群聊失败", "入群检测", group_id=event.group_id, e=e)
                await bot.send_private_msg(
                    user_id=int(list(bot.config.superusers)[0]),
                    message=f"触发强制入群保护，退出群聊 {event.group_id} 失败...",
                )
        # 默认群功能开关
        elif event.group_id not in group_manager.get_data().group_manager.keys():
            data = plugins2settings_manager.get_data()
            for plugin in data.keys():
                if not data[plugin].default_status:
                    group_manager.block_plugin(plugin, str(event.group_id))
            admin_default_auth = config["admin_bot_manage"]["ADMIN_DEFAULT_AUTH"]
            # 即刻刷新权限
            for user_info in await bot.get_group_member_list(group_id=event.group_id):
                if (user_info["role"] in ["owner", "admin",] 
                    and admin_default_auth is not None):
                    logger.debug(
                        f"添加默认群管理员权限: {admin_default_auth}",
                        "入群检测",
                        user_info["user_id"],
                        user_info["group_id"],
                    )
                if str(user_info["user_id"]) in bot.config.superusers:
                    logger.debug(
                        f"添加超级用户权限: 9",
                        "入群检测",
                        user_info["user_id"],
                        user_info["group_id"],
                    )


@group_decrease_handle.handle()
async def _(bot: Bot, event: GroupDecreaseNoticeEvent):
    # 被踢出群
    if event.sub_type == "kick_me":
        group_id = event.group_id
        operator_id = event.operator_id
        coffee = int(list(bot.config.superusers)[0])
        await bot.send_private_msg(
            user_id=coffee,
            message=f"****呜..一份踢出报告****\n"
            f"我被 {operator_id}\n"
            f"踢出了 ({group_id})\n"
            f"日期：{str(datetime.now()).split('.')[0]}",
        )
        return
    if event.user_id == int(bot.self_id):
        group_manager.delete_group(event.group_id)
        return
    
    logger.info(f"名称: {event.user_id} 退出群聊{event.group_id}")
