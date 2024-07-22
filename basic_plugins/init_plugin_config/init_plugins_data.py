
from utils.manager import plugins_manager, plugin_data_manager

def init_plugins_data():
    """
    初始化插件数据信息
    """
    for key in plugin_data_manager.keys():
        plugin_data = plugin_data_manager.get(key)

        plugin_version = plugin_data.plugin_status.version
        plugin_name = plugin_data.name
        plugin_author = plugin_data.plugin_status.author

        if plugin_data.model not in plugins_manager.keys():
            plugins_manager.add_plugin_data(
                plugin_data.model,
                plugin_name=plugin_name,
                author=plugin_author,
                version=plugin_version,
            )

            
    plugins_manager.save()
