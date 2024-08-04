from nonebot.adapters.onebot.v11 import Bot, GroupMessageEvent, Message, MessageEvent
from nonebot import require, on_command, on_message, on_fullmatch, get_bot
from nonebot.params import CommandArg
from utils.message_builder import image
from nonebot.plugin import PluginMetadata
require("nonebot_plugin_apscheduler")

from nonebot_plugin_apscheduler import scheduler
from json import load, dump
from os.path import dirname, join, exists
from .exchange import DNFExRate_, DNFExRateTrend_, tmpDNFExRateTrendPath, tmpDNFExRatePath
from .news import colg_news, new_scheduled_job
from nonebot.permission import SUPERUSER

from nonebot.adapters.onebot.v11.permission import (
    GROUP_ADMIN,
    GROUP_OWNER,
)


__plugin_usage__ = """
usage:
    比例趋势 跨2
""".strip()

__plugin_meta__ = PluginMetadata(
    name="DNF",
    description="无",
    usage= __plugin_usage__
)



PERM_EDIT = GROUP_ADMIN | GROUP_OWNER | SUPERUSER

DNFExRateTrend = on_command("比例趋势", priority=5, block=True)
DNFExRate = on_command("DNF比例", aliases={"比例", "游戏币比例", "游戏币", "金币"}, priority=5, block=True)
DNFMaodun = on_command("矛盾", priority=5, block=True)
colgNews = on_command(("colg资讯"), priority=5, block=True)
bookColgNews = on_command(("订阅colg资讯"), priority=1, block=True, permission=PERM_EDIT)
cancelColgNews = on_command("取消colg资讯", priority=1, block=True, permission=PERM_EDIT)


@DNFExRateTrend.handle()
async def _(event: MessageEvent, arg: Message = CommandArg()):
    server = arg.extract_plain_text().strip()
    if not server:
        return
    if DNFExRateTrend_(server, tmpDNFExRateTrendPath):
        await DNFExRateTrend.finish(image("DNFExRateTrend.png"))
    return


@DNFExRate.handle()
async def _(event: MessageEvent, arg: Message = CommandArg()):
    server = arg.extract_plain_text().strip()
    if not server:
        return
    url, base64_data = await DNFExRate_(server, 'youxibi')
    if base64_data == None:
        return

    await DNFExRate.finish(image("base64://" + base64_data) + url)


@DNFMaodun.handle()
async def _(event: MessageEvent, arg: Message = CommandArg()):
    server = arg.extract_plain_text().strip()
    if not server:
        return
    url, base64_data = await DNFExRate_(server, 'maodundejiejingti')
    if base64_data == None:
        return
    await DNFMaodun.finish(image("base64://"+base64_data) + url)


@colgNews.handle()
async def _(event: MessageEvent, arg: Message = CommandArg()):
    await colgNews.finish(await colg_news())


curPath = dirname(__file__)
userPath = join(curPath, 'user.json')
user = {
  "group": [],
  "qq": []
}
if exists(userPath):
    with open(userPath) as fp:
        user = load(fp)


def save_binds():
    global user
    with open(userPath, 'w') as fp:
        dump(user, fp, indent=4)


@bookColgNews.handle()
async def _(event: MessageEvent, arg: Message = CommandArg()):
    global user
    save_binds()
    await bookColgNews.finish(f"订阅成功", at_sender=True, )


@cancelColgNews.handle()
async def _(event: MessageEvent, arg: Message = CommandArg()):
    global user

    save_binds()
    await cancelColgNews.finish(f"取消订阅成功")


@scheduler.scheduled_job(
    "interval",
    minutes=3,
)
async def __():
    global user
    bot = get_bot()
    news_list = await new_scheduled_job()
    for context in news_list:
        for group_id in user["group"]:
            await bot.send_group_msg(group_id=group_id, message=context)
        for user_id in user["qq"]:
            await bot.send_private_msg(user_id=user_id, message=context)