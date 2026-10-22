import sys, os, math, random, platform, subprocess, threading
from unittest import ut
import pathlib
from backend.io.io import io, thread
from backend.connection import wifi

#path setup
current_dir = pathlib.Path(__file__).parent.resolve()
parent_dir = current_dir.parent

sys.path.append(str(parent_dir))

def BAUDRATE():
    return 9600


class thread:
    
    def create(name, target, args=()):
        return threading.Thread(target=target, args=args, name=name)

    def start(thread):
        thread.start()
    
    def join(thread):
        thread.join()


def send_wifi_file(ip, file_path):
    wifi.send(ip, file_path)
