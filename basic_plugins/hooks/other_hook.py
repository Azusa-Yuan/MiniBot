from nonebot.matcher import Matcher
from nonebot.plugin import PluginMetadata
from nonebot.message import run_preprocessor, IgnoredException
from nonebot.typing import T_State
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
    if event.user_id == int(bot.self_id):
        raise IgnoredException("为什么AI会自己和自己聊天")





