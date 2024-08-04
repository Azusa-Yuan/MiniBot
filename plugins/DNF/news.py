from bs4 import BeautifulSoup
import requests
import difflib

# 配置
url_list = ["https://bbs.colg.cn/home.php?mod=space&uid=4120473",
            "https://bbs.colg.cn/home.php?mod=space&uid=80727"]
author_list =["白猫之惩戒", "魔法少女QB"]
key_words = ["韩服", "爆料", "国服", "前瞻", "韩测"]
limit = 6


async def colg(url):
    global limit
    base_url = "https://bbs.colg.cn/"

    # BeautifulSoup html解析器
    soup = BeautifulSoup(requests.get(url=url).text, 'html.parser')
    lis = soup.find("div", {"id": "thread_content"}).find_all("li")
    context = ""
    for li in lis[: limit]:
        context += li.text + '\r\n'
        context += base_url + li.find("a").get("href") + '\r\n'
    return context, lis[0].text


async def colg_news():
    context = ""
    for url in url_list:
        tmp_context, head = await colg(url)
        context += tmp_context
    return context

pre_head = []


async def new_scheduled_job():
    global pre_head
    if len(pre_head) == 0:
        for url in url_list:
            _, tmp_head = await colg(url)
            pre_head.append(tmp_head)

    new_list = []
    for order, url in enumerate(url_list):
        context, head = await colg(url)
        if head != pre_head[order]:
            if (any(key_word in head for key_word in key_words)
                    and difflib.SequenceMatcher(None, head, pre_head[order]).quick_ratio() < 0.8):
                context = "colg资讯已更新:\r\n" + str(context) + author_list[order]
                new_list.append(context)

            pre_head[order] = head

    return new_list

