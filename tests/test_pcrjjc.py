from datetime import datetime

import pytest
from nonebug import App
from nonebot.adapters.onebot.v11 import Bot, MessageEvent, Message
from nonebot.adapters.onebot.v11.event import Sender
@pytest.mark.asyncio
async def test_pcrjjc(app: App):
    from plugins.pcrjjc3 import jjcObserver
    
    sender = Sender(user_id=123456)

    event = MessageEvent(
        post_type="message" ,
        sub_type="other",
        user_id=12346,
        message_type="message_type",
        message_id=121415,
        message=Message("竞技场关注 2 748597743"),
        original_message=Message("竞技场关注 2 748597743"),
        raw_message="竞技场关注 2 748597743",
        font=32,
        sender=sender,
        to_me=False,
        time=int(datetime.now().timestamp()),
        self_id=123456,
    )
    
    async with app.test_matcher(jjcObserver) as ctx:
        bot = ctx.create_bot()
        ctx.receive_event(bot, event)
        ctx.should_call_send(event, "竞技场关注", result="请输入正确的uid")
        ctx.should_finished(jjcObserver)