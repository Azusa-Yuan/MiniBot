from utils.manager import (
    plugins2cd_manager,
    plugins2block_manager,
    plugins2count_manager,
    plugin_data_manager,
)
from utils.utils import get_matchers
from configs.config_path import DATA_PATH


def init_plugins_cd_limit():
    """
    加载 cd 限制
    """
    plugins2cd_file = DATA_PATH / "configs" / "plugins2cd.yaml"
    plugins2cd_file.parent.mkdir(exist_ok=True, parents=True)
    for key in plugin_data_manager.keys():
        plugin_data = plugin_data_manager.get(key)
        if plugin_data.plugin_cd:
            plugins2cd_manager.add_cd_limit(
                plugin_data.model, plugin_data.plugin_cd
            )
    plugins2cd_manager.save()
    # add_cd_limit只是添加数据，没有真正创建limiter
    plugins2cd_manager.reload_cd_limit()


def init_plugins_block_limit():
    """
    加载阻塞限制
    """
    for key in  plugin_data_manager.keys():
        plugin_data = plugin_data_manager.get(key)
        if plugin_data.plugin_block:
            plugins2block_manager.add_block_limit(
                plugin_data.model, plugin_data.plugin_block
            )
    
    plugins2block_manager.save()
    plugins2block_manager.reload_block_limit()


def init_plugins_count_limit():
    """
    加载次数限制
    """
    for key in plugin_data_manager.keys():
        plugin_data = plugin_data_manager.get(key)
        if plugin_data.plugin_count:
            plugins2count_manager.add_count_limit(
                plugin_data.model, plugin_data.plugin_count
            )

    plugins2count_manager.save()
    plugins2count_manager.reload_count_limit()
