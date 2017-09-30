
var _onreadyfuncs = [];

var onready = function(f) {
    _onreadyfuncs.push(function() {f();});
};

var ready = function() {
    for(var idx = 0; idx < _onreadyfuncs.length; idx++) _onreadyfuncs[idx]();
};


var quickreply = function(shorthash, longhash, url) {
    if (!window.location.pathname.startsWith("/t/"))
    {
        window.location.href = url;
        return;
    }
    var elem = document.getElementById("comment");
    if(!elem) return;
    elem.value += ">>" + shorthash + "\n";
};

var get_storage = function() {
    var st = null;
    if (window.localStorage) {
        st = window.localStorage;
    } else if (localStorage) {
        st = localStorage;
    }
    return st;
};

var setSrcQuery = function(e, q) {
	var src  = e.src;
	var p = src.indexOf('?');
	if (p >= 0) {
		src = src.substr(0, p);
	}
	e.src = src + "?" + q
};


var reloadImg = function(el) {
	setSrcQuery(el, "reload=" + (new Date()).getTime());
	return false;
};

// form resubmit
onready(function() {
  var submitPost = function(form, elem, cb) {
    var ajax = new XMLHttpRequest();
    ajax.onreadystatechange = function() {
      if (ajax.readyState == 4) {
        var err = "unknown error";
        var j = null;
        try {
          j = JSON.parse(ajax.responseText);
          err = j.error || err;
        } catch (ex) {
          err = "error parsing reply: "+ ex;
        }
        if(ajax.status == 201) {
          // success
          cb(null, j);
        } else if (ajax.status == 200) {
          cb(err, j);
        } else {
          cb("http "+ajax.status, j);
        }
      } else {
        elem.innerHTML += ".";
      }
    };
    ajax.open(form.method, form.action+"/json");
    ajax.send(new FormData(form));
  };
  var elems = document.getElementsByClassName("postbutton");
  if(elems && elems.length > 0 && elems[0]) {
    var e = elems[0];
    var parent = e.parentElement;
    var origText = e.value;
    e.remove();
    e = document.createElement("button");
    parent.appendChild(e);
    e.innerHTML = origText;
    e.onclick = function() {
      e.disabled = true;
      e.innerHTML = "posting ";
      submitPost(document.forms[0], e, function(err, j) {
        var msg = err || "posted";
        console.log(msg, j);
        e.innerHTML = msg;
        setTimeout(function() {
          e.disabled = false;
          e.innerHTML = origText;
        }, 1000);
        var img = document.getElementById("captcha_img");
        if (img) {
          reloadImg(img);
        }
      });
    }
  }
     
});

// captcha reload
onready(function(){

  var e = document.getElementById("captcha_img");
  if (e) {
    e.onclick = function() {
		  reloadImg(e);
	  };
  }
});

// rewrite all images to add inline expand
onready(function() {

    // is the filename matching an image?
    var filenameIsImage = function(fname) {
        return /\.(gif|jpeg|jpg|png|webp)/.test(fname.toLowerCase());
    };

    // setup image inlining for 1 element
    var setupInlineImage = function(thumb, url) {
        if(thumb.inlineIsSetUp) return;
        thumb.inlineIsSetUp = true;
        var img = thumb.querySelector("img.image");
        var expanded = false;
        var oldurl = img.src;
        thumb.onclick = function() {
            if (expanded) {
                img.setAttribute("class", "image");
                img.src = oldurl;
                expanded = false;
            } else {
                img.setAttribute("class", "expanded-thumbnail");
                img.src = url;
                expanded = true;
            }
            return false;
        };
    };

    // set up image inlining for all applicable children in an element
    var setupInlineImageIn = function(element) {
        var thumbs = element.querySelectorAll("a.image_link");
        for ( var i = 0 ; i < thumbs.length ; i++ ) {
            var url = thumbs[i].href;
            if (filenameIsImage(url)) {
                // match
                setupInlineImage(thumbs[i], url);
            }
        }
    };
    // Setup Javascript events for document
    setupInlineImageIn(document);

    // Setup Javascript events via updater
    if (window.MutationObserver) {
        var observer = new MutationObserver(function(mutations) {
            for (var i = 0; i < mutations.length; i++) {
                var additions = mutations[i].addedNodes;
                if (additions == null) continue;
                for (var j = 0; j < additions.length; j++) {
                    var node = additions[j];
                    if (node.nodeType == 1) {
                        setupInlineImageIn(node);
                    }
                }
            }
        });
        observer.observe(document.body, {childList: true, subtree: true});
    }
});

// set up post hider
onready(function() {

    var get_hidden_posts = function() {
        var st = get_storage();
        var prefix = "nntpchan_hide_post_";
        return  {
            all : function() {
                var msgids = [];
                for ( var k in st) {
                    if (k.indexOf(prefix) == 0) {
                        var m = k.substring(prefix.length);
                        msgids.push(m);
                    }
                }
                return msgids;
            },

            add : function (msg) {
                st[prefix+msg] = "post";
            },

            del : function (msg) {
                st.removeItem(prefix+msg);
            }
        };
    };


    // is a post elem an OP?
    var postIsOP = function(elem) {
        var ds = elem.dataset;
        return ds && ds.rootmsgid == ds.msgid ;
    };

    var hide_elem = function(elem, fade) {
        if (elem.style) {
            elem.style.display = "none";
        } else {
            elem.style = {display: "none" };
        }
        elem.dataset.userhide = "yes";
    };

    var unhide_elem = function(elem) {
        elem.style = "";
        elem.dataset.userhide = "no";
    };

    // return true if element is hidden
    var elemIsHidden = function(elem) {
        return elem.dataset && elem.dataset.userhide == "yes";
    };

    // hide a post
    var hidepost = function(elem, nofade) {
        console.log("hidepost("+elem.dataset.msgid+")");
        var posts = get_hidden_posts();
        if (posts) {
            // add to persitant hide
            posts.add(elem.dataset.msgidhash);
        }
        if(postIsOP(elem)) {
            // hide thread it's an OP
            var thread = document.getElementById("thread_"+elem.dataset.rootmsgidhash);
            if (thread) {
                var e = thread.getElementsByClassName("post");
                for ( var idx = 0; idx < e.length ; idx ++ ) {
                    if (e[idx].dataset.msgid == elem.dataset.msgid) continue; // don't apply
                    hidepost(e[idx]);
                }
            }
        }
        // hide attachments and post body
        var es = elem.getElementsByClassName("image");
        for (var idx = 0; idx < es.length ; idx ++ ) {
            hide_elem(es[idx]);
        }
        es = elem.getElementsByClassName("message_span");
        for (var idx = 0; idx < es.length ; idx ++ ) {
            hide_elem(es[idx]);
        }
        es = elem.getElementsByClassName("topicline");
        for (var idx = 0; idx < es.length ; idx ++ ) {
            hide_elem(es[idx]);
        }
        elem.dataset.userhide = "yes";
        elem.setHideLabel("[show]");
    };

    // unhide a post
    var unhidepost = function(elem) {
        console.log("unhidepost("+elem.dataset.msgid+")");
        var posts = get_hidden_posts();
        if (posts) {
            // remove from persiting hide
            posts.del(elem.dataset.msgidhash);
        }
        if(postIsOP(elem)) {
            var thread = document.getElementById("thread_"+elem.dataset.rootmsgidhash);
            if(thread) {
                var e = thread.getElementsByClassName("post");
                for ( var idx = 0; idx < e.length ; idx ++ ) {
                    if(e[idx].dataset.msgid == elem.dataset.msgid) continue;
                    unhidepost(e[idx]);
                }
            }
        }
        // unhide attachments and post body
        var es = elem.getElementsByClassName("image");
        for (var idx = 0; idx < es.length ; idx ++ ) {
            unhide_elem(es[idx]);
        }
        es = elem.getElementsByClassName("message_span");
        for (var idx = 0; idx < es.length ; idx ++ ) {
            unhide_elem(es[idx]);
        }
        es = elem.getElementsByClassName("topicline");
        for (var idx = 0; idx < es.length ; idx ++ ) {
            unhide_elem(es[idx]);
        }
        elem.dataset.userhide = "no";
        elem.setHideLabel("[hide]");
    };

    // hide a post given a callback that checks each post
    var hideposts = function(check_func) {
        var es = document.getElementsByClassName("post");
        for ( var idx = 0; idx < es.length ; idx ++ ) {
            var elem = es[idx];
            if(check_func && elem && check_func(elem)) {
                hidepost(elem);
            }
        }
    };

    // unhide all posts given callback
    // if callback is null unhide all
    var unhideall = function(check_func) {
        var es = document.getElementsByClassName("post");
        for (var idx=0 ; idx < es.length; idx ++ ) {
            var elem = es[idx];
            if(!check_func) { unhide(elem); }
            else if(check_func(elem)) { unhide(elem); }
        }
    };

    // inject posthide into page

    var posts = document.getElementsByClassName("post");
    for (var idx = 0 ; idx < posts.length; idx++ ) {
        var inject = function (elem) {
            var hider = document.createElement("a");
            hider.setAttribute("class", "hider");
            elem.setHideLabel = function (txt) {
                var e_hider = hider;
                e_hider.innerHTML = txt;
            };
            elem.hidepost = function() {
                var e_self = elem;
                var e_hider = hider;
                hidepost(e_self);
            };
            elem.unhidepost = function() {
                var e_self = elem;
                var e_hider = hider;
                unhidepost(e_self);
            };
            elem.isHiding = function() {
                var e_self = elem;
                return elemIsHidden(e_self);
            };
            hider.appendChild(document.createTextNode("[hide]"));
            hider.onclick = function() {
                var e_self = elem;
                if(e_self.isHiding()) {
                    e_self.unhidepost();
                } else {
                    e_self.hidepost();
                }
            };
            elem.appendChild(hider);
        };
        inject(posts[idx]);
    }
    // apply persiting hidden posts
    posts = get_hidden_posts();
    if(posts) {
        var all = posts.all();
        for ( var idx = 0 ; idx < all.length; idx ++ ) {
            var id = all[idx];
            var elem = document.getElementById(id);
            if(elem)
                hidepost(elem, true);
        }
    }
});
