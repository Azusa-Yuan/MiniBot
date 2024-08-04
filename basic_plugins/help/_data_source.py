from typing import Optional


from utils.manager.models import PluginType
from utils.manager import admin_manager, plugin_data_manager, plugins2settings_manager


def get_admin_help():
    """
    创建管理员帮助
    """
    
    plugin_data_ = plugin_data_manager.get_data()
    name_list = []
    for plugin_data in [plugin_data_[x] for x in plugin_data_]:
        if plugin_data.plugin_type == PluginType.ADMIN and plugin_data.usage:
            name_list.append(plugin_data.name)
        
    msg += "群管理员插件"    
    for name in name_list:
        msg += "\r\n" + name
    return msg



def get_plugin_help(msg: str, is_super: bool = False) -> Optional[str]:
    """
    说明:
        获取功能的帮助信息
    参数:
        :param msg: 功能cmd
        :param is_super: 是否为超级用户
    """
    module = plugins2settings_manager.get_plugin_module(msg) or admin_manager.get_plugin_module(msg)
    if module and (plugin_data := plugin_data_manager.get(module)):
        plugin_data.superuser_usage
        if is_super:
            result = plugin_data.superuser_usage
        else:
            result = plugin_data.usage
        if result:
            return result
    return None
