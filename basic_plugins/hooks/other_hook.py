from nonebot.matcher import Matcher
from nonebot.plugin import PluginMetadata
from nonebot.message import run_preprocessor, IgnoredException
from nonebot.typing import T_State
from ._utils import status_message_manager
from nonebot.adapters.onebot.v11 import (
    Bot,
    MessageEvent,
    PrivateMessageEvent,
    GroupMessageEvent,
)

__plugin_meta__ = PluginMetadata(
    name="other_hook [Hidden]",
    description="hook",
    usage= '''hook'''
)


# 为什么AI会自己和自己聊天
@run_preprocessor
async def _(matcher: Matcher, bot: Bot, event: PrivateMessageEvent, state: T_State):
    if not isinstance(event, MessageEvent):
        return
    if event.user_id == int(bot.self_id):
        raise IgnoredException("为什么AI会自己和自己聊天")


# 有命令就别说话了  啥意思  没看懂
@run_preprocessor
async def _(matcher: Matcher, bot: Bot, event: MessageEvent, state: T_State):
    if not isinstance(event, MessageEvent):
        return
    if matcher.type == "message" and matcher.plugin_name == "ai":
        if (
            isinstance(event, GroupMessageEvent)
            and not status_message_manager.check(event.group_id)
        ):
            status_message_manager.delete(event.group_id)
            raise IgnoredException("有命令就别说话了")
        elif (
            isinstance(event, PrivateMessageEvent)
            and not status_message_manager.check(event.user_id)
        ):
            status_message_manager.delete(event.user_id)
            raise IgnoredException("有命令就别说话了")




