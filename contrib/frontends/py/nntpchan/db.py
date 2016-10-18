from nntpchan import config

import sqlalchemy

def allowsMessage(msgid):
    return True

def allowsNewsgroup(group):
    return True



def init():
    """
    initialize db backend
    """
