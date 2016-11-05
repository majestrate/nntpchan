import subprocess
import os
img_ext = []
vid_ext = []

def generate(fname, tname, placeholder):
    """
    generate thumbnail
    """
    ext = fname.split('.')[-1]
    cmd = None
    if ext in img_ext:
        cmd = ['convert', '-thumbnail', '200', fname, tname]
    elif ext in vid_ext:
        cmd = ['ffmpeg', '-i', fname, '-vf', 'scale=300:200', '-vframes', '1', tname]

    if cmd is None:
        os.link(placeholder, tname)
    else:
        subprocess.call(cmd)
        
    
