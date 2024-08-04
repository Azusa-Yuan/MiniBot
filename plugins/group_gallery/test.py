import cv2
import pytesseract
import difflib
import os
import numpy as np
config = ('-l chi_sim --oem 1 --psm 3')
imPath = "superuser_help.png"
im = cv2.imread(imPath)
text = pytesseract.image_to_string(im, config=config).replace(' ', '').replace('\n', '')
print(text)