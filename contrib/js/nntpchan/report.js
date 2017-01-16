/** censor tools */


function show_censortools() {
    var e = document.getElementById("censor-tools");
    if(e) e.checked = true;
}

function nntpchan_report_thread(posthash) {
    var thread = document.getElementById("thread_"+posthash);
    if (!thread) return;
    var posts = thread.getElementsByClassName("post");
    for (var idx = 0; idx < posts.length; idx ++ ) {
        var post = posts[idx];
        nntpchan_report(post.dataset.msgid);
    }
}

function nntpchan_report(msgid, msgid_hash, refid, refid_hash) {
    var e = document.getElementById("nntpchan_censor_actions");
    if (!e) return;
    if(refid == msgid) {
        nntpchan_report_thread(refid_hash);
    } else {
        e.value += "delete "+msgid+"\n";
    }
    show_censortools();
}

function nntpchan_submit_censor(form, regular_url) {

    var result = document.getElementById("nntpchan_censor_result");

    var show_result = function(msg) {
        while(result.children.length > 0) {
            result.children[0].remove();
        }
        result.appendChild(document.createTextNode(msg));
    };

    var handle_result = function(j) {
        var err = j.error;
        if(err) {
            show_result("error: "+err);
            return;
        }
        var msgid = j.message_id;
        if(msgid) {
            show_result("submitted report as "+msgid);
        } else {
            show_result("post failed, bad captcha?");
        }
    };

    // build url to ctl
    var parts = regular_url.split('/');
    parts[parts.length-1] = 'ctl';
    var url = parts.join('/');
    url += '/json';
    console.log(url);
    var captcha = form.captcha.value;
    if(!captcha) {
        show_result("no captcha solution provided");
        return;
    }
    var secret = document.getElementById("nntp_censor_secret").value;
    if(!secret) {
        show_result("no mod key provided");
        return;
    }
    var actions = document.getElementById("nntpchan_censor_actions").value;
    if(!actions) {
        show_result("no mod actions provided");
    }
    var msg = "";
    var lines = actions.split("\n");
    for (var idx = 0; idx < lines.length; idx ++ ) {
        var line = lines[idx].trim();
        if(!line) continue;
        msg += line + "\n";
    }
    if(!msg) {
        show_result("no mod actions given");
        return;
    }
    msg = msg.trim() + "\n\n";
    var formdata = new FormData();
    formdata.append("name", "mod#"+secret);
    formdata.append("subject", "censor");
    formdata.append("message", msg);
    formdata.append("captcha", captcha);
    formdata.append("reference", "");
    nntpchan_apicall(url, handle_result, null, "POST", formdata);
}

