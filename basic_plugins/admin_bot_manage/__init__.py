import nonebot
from pathlib import Path
from nonebot.plugin import PluginMetadata

__plugin_meta__ = PluginMetadata(
    name="admin_bot_manage [Hidden]",
    description="无",
    usage= "无"
)



nonebot.load_plugins(str(Path(__file__).parent.resolve()))
