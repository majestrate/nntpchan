
// reply box function
function nntpchan_reply(parent, shorthash) {
    if (parent && document.dynreply) {
        var boardname = parent.dataset.newsgroup;
        var rootmsg = parent.dataset.rootmsgid;
        var replyto = getReplyTo();
        // set target
        replyto.setBoard(boardname);
        replyto.setRoot(rootmsg);
        // show it
        replyto.show();
    }
    
    var elem;
    if(document.reply && document.reply.visible()) {
        elem = document.reply.message;
    } else {
        elem = document.getElementById("postform_message");
    }
    if ( elem )
    {
        elem.value += ">>" + shorthash.substr(0,10) + "\n";
    }
}

function init(prefix) {
    
}
