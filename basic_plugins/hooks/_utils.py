import time

from nonebot.adapters.onebot.v11 import (
    Bot,
    Event,
    GroupMessageEvent,
    Message,
    MessageEvent,
    PokeNotifyEvent,
    PrivateMessageEvent,
)
from nonebot.exception import ActionFailed, IgnoredException
from nonebot.internal.matcher import Matcher
from configs.config import config
from services.log import logger
from utils.decorator import Singleton
from utils.manager import (
    admin_manager,
    group_manager,
    plugins2cd_manager,
    plugins2count_manager,
    plugins2settings_manager,
    plugins_manager,
)
from utils.manager.models import PluginType
from utils.message_builder import at
from utils.utils import FreqLimiter

other_limit_plugins = []

async def send_msg(msg: str, bot: Bot, event: MessageEvent):
    """
    说明:
        发送信息
    参数:
        :param msg: pass
        :param bot: pass
        :param event: pass
    """
    if "[uname]" in msg:
        uname = event.sender.card or event.sender.nickname or ""
        msg = msg.replace("[uname]", uname)
    if "[at]" in msg and isinstance(event, GroupMessageEvent):
        msg = msg.replace("[at]", str(at(event.user_id)))
    try:
        if isinstance(event, GroupMessageEvent):
            await bot.send_group_msg(group_id=event.group_id, message=Message(msg))
        else:
            await bot.send_private_msg(user_id=event.user_id, message=Message(msg))
    except ActionFailed:
        pass

class IsSuperuserException(Exception):
    pass


@Singleton
class AuthChecker:
    """
    权限检查
    """

    def __init__(self):
        # 暂时用不到  
        check_notice_info_cd = config["hook"]["CHECK_NOTICE_INFO_CD"]
        if check_notice_info_cd is None or check_notice_info_cd < 0:
            raise ValueError("模块: [hook], 配置项: [CHECK_NOTICE_INFO_CD] 为空或小于0")
        self._flmt = FreqLimiter(check_notice_info_cd)
        self._flmt_g = FreqLimiter(check_notice_info_cd)
        self._flmt_s = FreqLimiter(check_notice_info_cd)
        self._flmt_c = FreqLimiter(check_notice_info_cd)

    async def auth(self, matcher: Matcher, bot: Bot, event: Event):
        """
        说明:
            权限检查
        参数:
            :param matcher: matcher
            :param bot: bot
            :param event: event
        """
        user_id = getattr(event, "user_id", None)
        group_id = getattr(event, "group_id", None)
        
        
        try:
            if plugin_name := matcher.plugin_name:
                # self.auth_hidden(matcher, plugin_name)
                # if user_id and str(user_id) not in bot.config.superusers:
                await self.auth_basic(plugin_name, bot, event)
                self.auth_group(plugin_name, bot, event)
                await self.auth_admin(plugin_name, matcher, bot, event)
                await self.auth_plugin(plugin_name, matcher, bot, event)
                await self.auth_limit(plugin_name, bot, event)
                
        except IsSuperuserException:
            logger.debug(f"超级用户或被ban跳过权限检测...", "HOOK", user_id, group_id)

    # def auth_hidden(self, matcher: Matcher):
    #     if plugin_data := plugin_data_manager.get(matcher.plugin_name):         # type: ignore

    async def auth_limit(self, plugin_name: str, bot: Bot, event: Event):
        """
        说明:
            插件限制
        参数:
            :param plugin_name: 模块名
            :param bot: bot
            :param event: event
        """
        user_id = getattr(event, "user_id", None)
        if not user_id:
            return
        group_id = getattr(event, "group_id", None)
        
        # 检测cd
        if plugins2cd_manager.check_plugin_cd_status(plugin_name):
            if (plugin_cd_data := plugins2cd_manager.get_plugin_cd_data(plugin_name)
            ) and (plugin_data := plugins2cd_manager.get_plugin_data(plugin_name)):
                check_type = plugin_cd_data.check_type
                limit_type = plugin_cd_data.limit_type
                if (
                    (isinstance(event, PrivateMessageEvent) and check_type == "private")
                    or (isinstance(event, GroupMessageEvent) and check_type == "group")
                    or plugin_data.check_type == "all"
                ):
                    cd_type_ = user_id
                    if limit_type == "group" and isinstance(event, GroupMessageEvent):
                        cd_type_ = event.group_id
                    if not plugins2cd_manager.check(plugin_name, cd_type_):
                        await send_msg(f"{plugin_name} 正在cd中...", bot, event)  # type: ignore
                        logger.debug(
                            f"{plugin_name} 正在cd中...", "HOOK", user_id, group_id
                        )
                        raise IgnoredException(f"{plugin_name} 正在cd中...")
                    else:
                        plugins2cd_manager.start_cd(plugin_name, cd_type_)
        
        # Count
        if plugins2count_manager.check_plugin_count_status(plugin_name) \
            and user_id not in bot.config.superusers:
            if plugin_count_data := plugins2count_manager.get_plugin_count_data(plugin_name):
                limit_type = plugin_count_data.limit_type
                count_type_ = user_id
                if limit_type == "group" and isinstance(event, GroupMessageEvent):
                    count_type_ = event.group_id
                if not plugins2count_manager.check(plugin_name, count_type_):
                    await send_msg(msg, bot, event)  # type: ignore
                    logger.debug(
                        f"{plugin_name} count次数限制...", "HOOK", user_id, group_id
                    )
                    raise IgnoredException(f"{plugin_name} count次数限制...")
                else:
                    plugins2count_manager.increase(plugin_name, count_type_)

    async def auth_plugin(
        self, plugin_name: str, matcher: Matcher, bot: Bot, event: Event
    ):
        """
        说明:
            插件状态
        参数:
            :param plugin_name: 模块名
            :param matcher: matcher
            :param bot: bot
            :param event: event
        """
        if plugin_name in plugins2settings_manager.keys():
            user_id = getattr(event, "user_id", None)
            if not user_id:
                return
            group_id = getattr(event, "group_id", None)
            # 戳一戳单独判断
            if group_id:
                # 判断权限等级
                if plugins2settings_manager[plugin_name].level \
                    > group_manager.get_group_level(group_id):
                    raise IgnoredException("群权限不足")
                
                # 判断群管理是否关闭
                if group_manager.is_close(plugin_name, group_id):
                    logger.debug(f"{plugin_name} 未开启此功能...", "HOOK", user_id, group_id)
                    raise IgnoredException("未开启此功能...")
                
                # 判断插件是否群聊禁用
                if not plugins_manager.get_plugin_status(
                    plugin_name, block_type="group"
                ):
                    logger.debug(
                        f"{plugin_name} 该插件在群聊中已被禁用...", "HOOK", user_id, group_id
                    )
                    raise IgnoredException("该插件在群聊中已被禁用...")
            else:
                # 私聊没有分配权限等级
                # 私聊禁用
                if plugins_manager.get_plugin_status(
                    plugin_name, block_type="private"
                ):
                    # 判断插件是否群聊禁用
                    logger.debug(
                        f"{plugin_name} 该插件在私聊中已被禁用...", "HOOK", user_id, group_id
                    )
                    raise IgnoredException("该插件在私聊中已被禁用...")
                
            # 维护
            if not plugins_manager.get_plugin_status(plugin_name, block_type="all"):
                # 白名单群号，测试用
                if group_id and group_manager.check_group_is_white(event.group_id):
                    raise IsSuperuserException()
                logger.debug(f"{plugin_name} 此功能正在维护...", "HOOK", user_id, group_id)
                raise IgnoredException("此功能正在维护...")

    async def auth_admin(
        self, plugin_name: str, matcher: Matcher, bot: Bot, event: Event
    ):
        """
        说明:
            管理员命令 个人权限
        参数:
            :param plugin_name: 模块名
            :param matcher: matcher
            :param bot: bot
            :param event: event
        """
        user_id = getattr(event, "user_id", None)
        if not user_id:
            return
        group_id = getattr(event, "group_id", None)
        if plugin_name in admin_manager.keys():
            # 私聊默认是自己的管理？这里  很奇怪  需要重构
            if isinstance(event, GroupMessageEvent):
                # 个人权限
                try:
                    # 因为event.sender.role不是协议里规定必须提供的，所以要用try
                    if event.sender.role == "member":             
                        logger.debug(f"{plugin_name} 管理员权限不足...", "HOOK", user_id, group_id)
                        raise IgnoredException("管理员权限不足")
                except:
                    logger.debug(f"{plugin_name} 管理员权限不足...", "HOOK", user_id, group_id)
                    raise IgnoredException("管理员权限不足")

    def auth_group(self, plugin_name: str, bot: Bot, event: Event):
        """
        说明:
            群总开关检测
        参数:
            :param plugin_name: 模块名
            :param bot: bot
            :param event: event
        """
        user_id = getattr(event, "user_id", None)
        group_id = getattr(event, "group_id", None)
        if not group_id:
            return
        
        if not group_manager.check_group_bot_status(group_id):
            try:
                if str(event.get_message()) != "醒来":
                    logger.debug(
                        f"{plugin_name} 功能总开关关闭状态...", "HOOK", user_id, group_id
                    )
                    raise IgnoredException("功能总开关关闭状态")
            except ValueError:
                logger.debug(f"{plugin_name} 功能总开关关闭状态...", "HOOK", user_id, group_id)
                raise IgnoredException("功能总开关关闭状态")

    async def auth_basic(self, plugin_name: str, bot: Bot, event: Event):
        """
        说明:
            检测是否满足超级用户权限，是否被ban等
        参数:
            :param plugin_name: 模块名
            :param bot: bot
            :param event: event
        """
        user_id = getattr(event, "user_id", None)
        if not user_id:
            return
        plugin_setting = plugins2settings_manager.get_plugin_data(plugin_name)
        if str(user_id) in bot.config.superusers:
            if not plugin_setting:
                raise IsSuperuserException()
            else:
                if not plugin_setting.limit_superuser:
                    raise IsSuperuserException()
                    
    