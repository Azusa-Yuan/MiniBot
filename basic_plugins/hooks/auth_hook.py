from typing import Optional
from nonebot.plugin import PluginMetadata
from nonebot.adapters.onebot.v11 import (
    Bot,
    MessageEvent,
    PrivateMessageEvent,
    Event,
)
from nonebot.matcher import Matcher
from nonebot.message import run_preprocessor, run_postprocessor, IgnoredException
from nonebot.typing import T_State

from ._utils import (
    AuthChecker,
)

__plugin_meta__ = PluginMetadata(
    name="auth_hook [Hidden]",
    description="hook",
    usage= '''hook'''
)


# 权限检测
@run_preprocessor
async def _(matcher: Matcher, bot: Bot, event: Event):
    await AuthChecker().auth(matcher, bot, event)


# 私聊处理
@run_preprocessor
async def _(matcher: Matcher, bot: Bot, event: PrivateMessageEvent, state: T_State):
    if event.sub_type != "friend":
        raise IgnoredException("只允许好友触发")