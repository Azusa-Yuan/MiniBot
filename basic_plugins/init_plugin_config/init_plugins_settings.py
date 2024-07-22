from services.log import logger
from utils.manager import admin_manager, plugin_data_manager, plugins2settings_manager
from utils.manager.models import PluginType


def init_plugins_settings():
    """
    初始化插件设置
    """
    for key in plugin_data_manager.keys():
        plugin_data = plugin_data_manager.get(key)
        name = plugin_data.model
        try:
            if plugin_settings := plugin_data.plugin_setting:
                # 为了进行统计以及开关控制
                if name not in plugin_settings.cmd:
                    plugin_settings.cmd.append(name)
                # 管理员命令
                if plugin_data.plugin_type == PluginType.ADMIN:
                    admin_manager.add_admin_plugin_settings(
                        name,
                        plugin_settings.cmd,
                        plugin_settings.level,
                    )
                else:
                    plugins2settings_manager.add_plugin_settings(
                        name, plugin_settings
                    )
        except Exception as e:
            logger.error(
                f"{name} 初始化 plugin_settings 发生错误 {type(e)}：{e}"
            )
    plugins2settings_manager.save()
    logger.info(f"已成功加载 {len(plugins2settings_manager.get_data())} 个非限制插件.")
