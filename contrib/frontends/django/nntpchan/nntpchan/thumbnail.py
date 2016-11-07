from django.conf import settings
import subprocess
import os
img_ext = ['png', 'jpg', 'jpeg', 'gif', 'webp', 'ico']
vid_ext = ['mp4', 'webm', 'm4v', 'ogv', 'avi', 'txt']

def generate(fname, tname, placeholder):
    """
    generate thumbnail
    """
    ext = fname.split('.')[-1].lower()
    cmd = None
    if ext in img_ext:
        cmd = [settings.CONVERT_PATH, '-thumbnail', '200', fname, tname]
    elif ext in vid_ext:
        cmd = [settings.FFMPEG_PATH, '-i', fname, '-vf', 'scale=300:200', '-vframes', '1', tname]

    if cmd is None:
        os.link(placeholder, tname)
    else:
        subprocess.run(cmd, check=True)
        
    
