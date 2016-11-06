import subprocess
import os
img_ext = ['png', 'jpg', 'jpeg', 'gif', 'webp', 'ico']
vid_ext = ['mp4', 'webm', 'm4v', 'ogv', 'avi']
txt_ext = ['txt', 'pdf', 'ps']

def generate(fname, tname, placeholder):
    """
    generate thumbnail
    """
    ext = fname.split('.')[-1]
    cmd = None
    if ext in img_ext:
        cmd = ['/usr/bin/convert', '-thumbnail', '200', fname, tname]
    elif ext in vid_ext or ext in txt_ext:
        cmd = ['/usr/bin/ffmpeg', '-i', fname, '-vf', 'scale=300:200', '-vframes', '1', tname]

    if cmd is None:
        os.link(placeholder, tname)
    else:
        subprocess.run(cmd, check=True)
        
    
