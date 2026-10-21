import io
import platform
import subprocess
import random
import sys
import threading
from main import hash
import pathlib
from src.main import io, thread
import pyserial



#path setup
current_dir = pathlib.Path(__file__).parent.resolve()
parent_dir = current_dir.parent

sys.path.append(str(parent_dir))

thread.create("input", target=io.input, args=())

def BAUDRATE():
    return 9600

class io:
    def input():
        
        
