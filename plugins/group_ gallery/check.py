from configs.path import IMAGE_PATH, DATA_PATH
import cv2
import pytesseract
import difflib
import os
import numpy as np
from .SimilarAlgorithm import aHash, cmpHash
import shutil

config = ('-l chi_sim --oem 1 --psm 3')
init_shape = 8
IMAGE_PATH = os.path.join(IMAGE_PATH, "Black_history")
DATA_PATH = os.path.join(DATA_PATH, "Black_history")
if not os.path.exists(IMAGE_PATH):
    os.makedirs(IMAGE_PATH)
if not os.path.exists(DATA_PATH):
    os.makedirs(DATA_PATH)

# 发现相似图片后的操作
def similar_next(index1, index2, imPath, data_path):
    num = len(os.listdir(data_path))
    shutil.move(imPath, os.path.join(data_path, f"{num}.jpg"))
    print(f"{data_path}\t{index1}与{index2}相似")

def check_all():
    qq_list = os.listdir(IMAGE_PATH)
    for qq in qq_list:
        file_path = os.path.join(DATA_PATH, str(qq))
        img_list_path = os.path.join(IMAGE_PATH, str(qq))

        if not os.path.exists(file_path):
            os.makedirs(file_path)
        data_path = os.path.join(file_path, "data.npy")
        data_list = []

        img_list = os.listdir(img_list_path)

        # if os.path.exists(data_path):
        #     print("已有文件")
        #     data_list = np.load(data_path, allow_pickle=True).tolist()
        # else:
        #     data_list = []

        for img_path in img_list:
            flag = 0
            hash_flag = 0
            num = int(img_path.split(".")[0])
            imPath = os.path.join(img_list_path, img_path)
            im = cv2.imread(imPath)
            text = pytesseract.image_to_string(im, config=config).replace(' ', '').replace('\n', '')
            if len(text) < 2:
                hash_flag = 1
                img = cv2.resize(im, (init_shape, init_shape), interpolation=cv2.INTER_CUBIC)
                gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)
                text = aHash(gray, init_shape)
                for i, data_flag, hash_cmp in data_list:
                    if int(data_flag) == 0:
                        continue
                    # print(hash_cmp)
                    similar_a = cmpHash(text, hash_cmp) / (init_shape * init_shape)
                    # print('均值哈希算法相似度：' + str(similar_a*100) + '%'+ImgName)
                    if similar_a > 0.9:
                        similar_next(i, num, imPath, file_path)
                        flag = 1
                        break
            else:
                for i, data_flag, hash_cmp in data_list:
                    if int(data_flag) == 1:
                        continue
                    # print(hash_cmp)
                    similar = difflib.SequenceMatcher(None, text, hash_cmp).quick_ratio()
                    # print('均值哈希算法相似度：' + str(similar_a*100) + '%'+ImgName)
                    if similar > 0.5:
                        similar_next(i, num, imPath, file_path)
                        flag = 1
                        break
            if flag == 0:
                data_list.append([num, hash_flag, text])

        # 刷新图像列表
        img_list = os.listdir(img_list_path)
        img_list_num = []
        for img_path in img_list:
            num = int(img_path.split(".")[0])
            img_list_num.append([num, img_path])
            img_list_num.sort(key=lambda x: x[0])

        for num, img_path in enumerate(img_list_num):
            imPath = os.path.join(img_list_path, img_path[1])
            os.rename(imPath, os.path.join(img_list_path, f"{num}.jpg"))

        np.save(data_path, data_list)

def delhash(num1, qq):
    file_path = os.path.join(DATA_PATH, str(qq))
    data_path = os.path.join(file_path, "data.npy")
    if os.path.exists(data_path):
        print("出错，没有hash文件")

    data_list = np.load(data_path, allow_pickle=True).tolist()
    for num, single in enumerate(data_list):
        i, data_flag, hash_cmp = single
        if int(i) == int(num1):
            data_list.pop(num)
            np.save(data_path, data_list)
            return True
    return False


def check_single(num, qq):
    hash_flag = 0
    file_path = os.path.join(DATA_PATH, str(qq))

    data_path = os.path.join(file_path, "data.npy")
    img_list_path = os.path.join(IMAGE_PATH, str(qq))
    imPath = os.path.join(img_list_path, f"{num}.jpg")

    if not os.path.exists(file_path):
        os.makedirs(file_path)
    if os.path.exists(data_path):
        data_list = np.load(data_path, allow_pickle=True).tolist()
    else:
        data_list = []

    im = cv2.imread(imPath)
    text = pytesseract.image_to_string(im, config=config).replace(' ', '').replace('\n', '')
    print(text)

    if len(text) < 2:
        hash_flag = 1
        img = cv2.resize(im, (init_shape, init_shape), interpolation=cv2.INTER_CUBIC)
        gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)
        text = aHash(gray, init_shape)
        for i, data_flag, hash_cmp in data_list:
            if int(data_flag) == 0:
                continue
            # print(hash_cmp)
            similar_a = cmpHash(text, hash_cmp) / (init_shape * init_shape)
            # print('均值哈希算法相似度：' + str(similar_a*100) + '%'+ImgName)
            if similar_a > 0.9:
                similar_next(i, num, imPath, file_path)
                return i
    else:
        for i, data_flag, hash_cmp in data_list:
            if int(data_flag) == 1:
                continue
            similar = difflib.SequenceMatcher(None, text, hash_cmp).quick_ratio()
            if similar > 0.5:
                similar_next(i, num, imPath, file_path)
                return i

    data_list.append([num, hash_flag, text])
    np.save(data_path, data_list)
    return -1






