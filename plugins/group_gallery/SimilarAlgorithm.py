import numpy as np
#import cv2
# 均值哈希算法
def aHash(gray, shape):

    # s为像素和初值为0，hash_str为hash值初值为''
    s = 0
    hash_str = ''
    # 遍历累加求像素和
    for i in range(shape):
        for j in range(shape):
            s = s + gray[i, j]
    # 求平均灰度
    avg = s / (shape * shape)
    # 灰度大于平均值为1相反为0生成图片的hash值
    for i in range(shape):
        for j in range(shape):
            if gray[i, j] > avg:
                hash_str = hash_str + '1'
            else:
                hash_str = hash_str + '0'
    return hash_str


# 差值感知算法
def dHash(gray):
    hash_str = ''
    shape = gray.shape[1]
    # 每行前一个像素大于后一个像素为1，相反为0，生成哈希
    for i in range(shape):
        for j in range(shape - 1):
            if gray[i, j] > gray[i, j+1]:
                hash_str = hash_str+'1'
            else:
                hash_str = hash_str+'0'
    return hash_str


# Hash值对比
def cmpHash(hash_a, hash_b):
    n = 0
    # hash长度不同则返回-1代表传参出错
    if len(hash_a) != len(hash_b):
        return -1
    # 遍历判断
    for i in range(len(hash_a)):
        # 相等则n计数+1，n最终为相似度
        if hash_a[i] == hash_b[i]:
            n += 1
        # print(n)
    return n


# 通过得到RGB每个通道的直方图来计算相似度
def classify_hist_with_split(image1, image2):
    # 分离为RGB三个通道，再计算每个通道的相似值
    sub_image1 = cv2.split(image1)
    sub_image2 = cv2.split(image2)
    sub_data = 0
    for im1, im2 in zip(sub_image1, sub_image2):
        sub_data += calculate(im1, im2)
    sub_data = sub_data / 3
    return sub_data


# 计算单通道的直方图的相似值
def calculate(image1, image2):
    hist1 = cv2.calcHist([image1], [0], None, [256], [0.0, 255.0])
    hist2 = cv2.calcHist([image2], [0], None, [256], [0.0, 255.0])
    # 计算直方图的重合度
    degree = 0
    for i in range(len(hist1)):
        if hist1[i] != hist2[i]:
            degree = degree + (1 - abs(hist1[i] - hist2[i]) / max(hist1[i], hist2[i]))
        else:
            degree = degree + 1
    degree = degree / len(hist1)
    return degree


# 感知哈希算法(pHash)
# def pHash(gray):
#     # 将灰度图转为浮点型，再进行dct变换
#     dct = cv2.dct(np.float32(gray))
#     # opencv实现的掩码操作
#     dct_roi = dct[0:10, 0:10]
#     hash = []
#     avreage = np.mean(dct_roi)
#     for i in range(dct_roi.shape[0]):
#         for j in range(dct_roi.shape[1]):
#             if dct_roi[i, j] > avreage:
#                 hash.append(1)
#             else:
#                 hash.append(0)
#     return hash