#!/usr/bin/env python3
import tiktoken
import sys

def calculate_byte_token_rate(file_path):
    """计算文本的字节数和token数量的比值"""
    # 读取文件内容
    with open(file_path, 'r', encoding='utf-8') as f:
        text = f.read()
    
    # 获取字节数
    byte_count = len(text.encode('utf-8'))
    
    # 使用cl100k_base编码器
    encoding = tiktoken.get_encoding("cl100k_base")
    tokens = encoding.encode(text)
    token_count = len(tokens)
    
    # 计算比值
    if token_count > 0:
        rate = byte_count / token_count
    else:
        rate = 0
    
    # 输出结果
    print(f"文件: {file_path}")
    print(f"字节数: {byte_count}")
    print(f"Token数: {token_count}")
    print(f"字节/Token比值: {rate:.2f}")
    
    return rate

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("用法: python cal_byte_token_rate.py <文件路径>")
        sys.exit(1)
    
    file_path = sys.argv[1]
    calculate_byte_token_rate(file_path)

