import asyncio
from typing import List
from nonebot.plugin import PluginMetadata
from nonebot import on_command
from nonebot.adapters.onebot.v11 import Bot, Message, MessageEvent
from nonebot.params import CommandArg
from nonebot.permission import SUPERUSER

from services.log import logger
from utils.manager import group_manager
from utils.message_builder import image


plugin_usage = """
usage：
    指令：
        广播- ?[消息] ?[图片]
        示例：广播- 你们好！
""".strip()

__plugin_meta__ = PluginMetadata(
    name="广播 [Superuser]",
    description="hook",
    usage= plugin_usage
)


broadcast = on_command("广播-", priority=1, permission=SUPERUSER, block=True)


@broadcast.handle()
async def _(
    bot: Bot,
    event: MessageEvent,
    arg: Message = CommandArg(),
):
    msg = arg.extract_plain_text().strip()
    rst = ""
    gl = [
        g["group_id"]
        for g in await bot.get_group_list()
        if group_manager.check_group_task_status(str(g["group_id"]), "broadcast")
    ]
    g_cnt = len(gl)
    cnt = 0
    error = ""
    x = 0.25
    for g in gl:
        cnt += 1
        if cnt / g_cnt > x:
            await broadcast.send(f"已播报至 {int(cnt / g_cnt * 100)}% 的群聊")
            x += 0.25
        try:
            await bot.send_group_msg(group_id=g, message=msg + rst)
            logger.info(f"投递广播成功", "广播", group_id=g)
        except Exception as e:
            logger.error(f"投递广播失败", "广播", group_id=g, e=e)
            error += f"GROUP {g} 投递广播失败：{type(e)}\n"
        await asyncio.sleep(0.5)
    await broadcast.send(f"已播报至 100% 的群聊")
    if error:
        await broadcast.send(f"播报时错误：{error}")
