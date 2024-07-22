import nonebot
from nonebot import Driver
from nonebot.plugin import PluginMetadata

from services.log import logger

from .init_plugin_info import init_plugin_info
from .init_plugins_data import init_plugins_data
from .init_plugins_limit import (
    init_plugins_block_limit,
    init_plugins_cd_limit,
    init_plugins_count_limit,
)
from .init_plugins_resources import init_plugins_resources
from .init_plugins_settings import init_plugins_settings

__plugin_meta__ = PluginMetadata(
    name="初始化插件数据 [Hidden]",
    description="插件管理",
    usage= "无",
    extra={
        "plugin_version": 0.1,
        "plugin_author": "HibiKier"
    },
)


driver: Driver = nonebot.get_driver()


@driver.on_startup
async def _():
    """
    初始化数据
    """
    init_plugin_info()
    init_plugins_settings()
    init_plugins_cd_limit()
    init_plugins_block_limit()
    init_plugins_count_limit()
    init_plugins_data()
    # 未来可能去除
    init_plugins_resources()
    # 删除
    # init_none_plugin_count_manager()
    logger.info("初始化数据完成...")


