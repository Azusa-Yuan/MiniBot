
from typing import Dict, List, Optional

from utils.decorator import Singleton
from utils.manager import group_manager, plugin_data_manager
from utils.manager.models import PluginData, PluginType



class HelpBuild:
    def __init__(self):
        self._data: Dict[str, PluginData] = plugin_data_manager.get_data()
        self._sort_data: Dict[str, List[PluginData]] = {}
        

    def sort_type(self):
        """
        说明:
            对插件按照菜单类型分类
        """
        if not self._sort_data.keys():
            for key in self._data.keys():
                plugin_data = self._data[key]
                if plugin_data.plugin_type == PluginType.NORMAL:
                    if not self._sort_data.get(plugin_data.menu_type[0]):  # type: ignore
                        self._sort_data[plugin_data.menu_type[0]] = []  # type: ignore
                    self._sort_data[plugin_data.menu_type[0]].append(self._data[key])  # type: ignore


    async def build_normal_help(self, group_id: Optional[int]) -> str:
        self.sort_type()
        msg = "功能列表:"
        for menu in self._sort_data:
            msg += f"\r\n{menu}:"
            for plugin in self._sort_data[menu]:
                if not plugin.plugin_status.status:
                    if group_id and plugin.plugin_status.block_type in ["all", "group"]:
                        continue
                    if not group_id and plugin.plugin_status.block_type in ["all","private",]:
                        continue
                if group_id and not group_manager.get_plugin_super_status(plugin.model, group_id):
                    continue
                if group_id and not group_manager.get_plugin_status(plugin.model, group_id):
                    msg += f'\r\f - {plugin.name}'
       
        return msg

