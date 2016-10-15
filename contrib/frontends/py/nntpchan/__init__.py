
from nntpchan import store
from nntpchan import message
from nntpchan import preprocess


def addArticle(msgid, group):
    """
    add article to system
    :return True if we need to regenerate otherwise False:
    """
    msg = None
    # open message
    with store.openArticle(msgid) as f:
        # read article header
        hdr = message.readHeader(f)

        mclass = message.MultipartMessage
        if hdr.isTextOnly:
            # treat as text message instead of multipart
            mclass = message.TextMessage
        elif hdr.isSigned:
            # treat as signed message
            mclass = message.TripcodeMessage
        # create messgae
        msg = mclass(hdr, f)
        
    if msg is not None:
        # we got a message that is valid
        store.storeMessage(msg)
    else:
        # invalid message
        print("invalid message: {}".format(msgid))
    return msg is not None
        
        
def regenerate(msgid, group):
    """
    regenerate markup
    """
    pass
    
