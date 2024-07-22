import os

from nonebot import on_command
from nonebot.adapters.onebot.v11 import Bot, GroupMessageEvent, Message, MessageEvent
from nonebot.params import CommandArg
from nonebot.rule import to_me

from configs.config_path import DATA_PATH, IMAGE_PATH
from services.log import logger
from nonebot.plugin import PluginMetadata
from ._data_source import get_plugin_help, get_admin_help


__plugin_meta__ = PluginMetadata(
    name="帮助",
    description="帮助",
    usage= "帮助 管理员帮助",
    extra={
        "plugin_version": 0.1,
        "plugin_author": "HibiKier"
    },
)
admin_help = on_command("管理员帮助", aliases={"管理帮助"}, priority=5, block=True)

@admin_help.handle()
async def _(bot: Bot, event: GroupMessageEvent):
    await admin_help.send(get_admin_help())



simple_help = on_command(
    "功能", rule=to_me(), aliases={"help", "帮助"}, priority=1, block=True
)


@simple_help.handle()
async def _(bot: Bot, event: MessageEvent, arg: Message = CommandArg()):
    msg = arg.extract_plain_text().strip()
    is_super = False
    if msg:
        if "-super" in msg:
            if str(event.user_id) in bot.config.superusers:
                is_super = True
            msg = msg.replace("-super", "").strip()
        msg = get_plugin_help(msg, is_super)
        if msg:
            await simple_help.send(msg)
        else:
            await simple_help.send("没有此功能的帮助信息...")
        logger.info(
            f"查看帮助详情: {msg}", "帮助", event.user_id, getattr(event, "group_id", None)
        )
    # else:
    #     if isinstance(event, GroupMessageEvent):
    #         if not _image_path.exists():
    #             await create_help_img(event.group_id)
    #         await simple_help.send(image(_image_path))
    #     else:
    #         if not simple_help_image.exists():
    #             if simple_help_image.exists():
    #                 simple_help_image.unlink()
    #             await create_help_img(None)
    #         await simple_help.finish(image("simple_help.png"))
