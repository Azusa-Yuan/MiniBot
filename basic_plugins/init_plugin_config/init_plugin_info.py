from types import ModuleType
from typing import Any, Dict

from services import logger
from utils.manager import (
    plugin_data_manager,
    plugins2block_manager,
    plugins2cd_manager,
    plugins2count_manager,
    plugins2settings_manager,
    plugins_manager,
)
from utils.manager.models import (
    Plugin,
    PluginBlock,
    PluginCd,
    PluginCount,
    PluginData,
    PluginSetting,
    PluginType,
)
import nonebot


def get_attr(module: ModuleType, name: str, default: Any = None) -> Any:
    """
    说明:
        获取属性
    参数:
        :param module: module
        :param name: name
        :param default: default
    """
    return getattr(module, name, None) or default


def init_plugin_info():        
    
    for plugin in nonebot.get_loaded_plugins():
        try:
            if plugin.name:
                metadata = plugin.metadata
                extra = metadata.extra if metadata else {}
                if hasattr(plugin, "module"):
                    plugin_model = plugin.name
                    plugin_name = metadata.name
                    if not plugin_name:
                        logger.error(f"配置文件 模块：{plugin_model} 获取 plugin_name 失败...")
                        continue
                    # 提前为插件设置权限
                    if "[Admin]" in plugin_name:
                        plugin_type = PluginType.ADMIN
                        plugin_name = plugin_name.replace("[Admin]", "").strip()
                    elif "[Hidden]" in plugin_name:
                        plugin_type = PluginType.HIDDEN
                        plugin_name = plugin_name.replace("[Hidden]", "").strip()
                    elif "[Superuser]" in plugin_name:
                        plugin_type = PluginType.SUPERUSER
                        plugin_name = plugin_name.replace("[Superuser]", "").strip()
                    else:
                        plugin_type = PluginType.NORMAL
                    plugin_usage = metadata.usage
                    plugin_des = metadata.description
                    menu_type = extra.get("plugin_type") or ("normal",)
                    plugin_setting = extra.get("plugin_settings")
                    if plugin_setting:
                        plugin_setting = PluginSetting(**plugin_setting)
                        plugin_setting.plugin_type = menu_type
                    plugin_superuser_usage = extra.get("superuser_usage")

                    plugin_cd = extra.get("plugin_cd_limit")
                    if plugin_cd:
                        plugin_cd = PluginCd(**plugin_cd)
                    plugin_block = extra.get("plugin_block_limit")
                    if plugin_block:
                        plugin_block = PluginBlock(**plugin_block)
                    plugin_count = extra.get("plugin_count_limit")
                    if plugin_count:
                        plugin_count = PluginCount(**plugin_count)
                    plugin_resources = extra.get("plugin_resources")
                    plugin_configs = extra.get("plugin_configs")
                    # 配置文件优先级比插件元数据的优先级高 如果有会覆盖
                    if settings := plugins2settings_manager.get(plugin_model):
                        plugin_setting = settings
                    if plugin_cd_limit := plugins2cd_manager.get(plugin_model):
                        plugin_cd = plugin_cd_limit
                    if plugin_block_limit := plugins2block_manager.get(plugin_model):
                        plugin_block = plugin_block_limit
                    if plugin_count_limit := plugins2count_manager.get(plugin_model):
                        plugin_count = plugin_count_limit
                    # 统一插件配置信息管理
                   
                    plugin_status = plugins_manager.get(plugin_model)
                    if not plugin_status:
                        plugin_status = Plugin(plugin_name=plugin_model)
                    plugin_data = PluginData(
                        model=plugin_model,
                        name=plugin_name.strip(),
                        plugin_type=plugin_type,
                        usage=plugin_usage,
                        superuser_usage=plugin_superuser_usage,
                        des=plugin_des,
                        menu_type=menu_type,
                        plugin_setting=plugin_setting,
                        plugin_cd=plugin_cd,
                        plugin_block=plugin_block,
                        plugin_count=plugin_count,
                        plugin_resources=plugin_resources,
                        plugin_configs=plugin_configs,  # type: ignore
                        plugin_status=plugin_status,
                    )
                    plugin_data_manager.add_plugin_info(plugin_data)
        except Exception as e:
            logger.error(f"构造插件数据失败 {plugin.name}", e=e)
