import pytest
import nonebot
# 导入适配器
from nonebot.adapters.onebot.v11 import Adapter as ConsoleAdapter
from nonebug import NONEBOT_INIT_KWARGS
import os

os.environ["ENVIRONMENT"] = "test"

@pytest.fixture(scope="session", autouse=True)
def load_bot():
    # 加载适配器
    driver = nonebot.get_driver()
    driver.register_adapter(ConsoleAdapter)

    # 加载插件
    nonebot.load_plugin("nonebot_plugin_apscheduler")
    nonebot.load_plugins("basic_plugins")
    nonebot.load_plugins("plugins")