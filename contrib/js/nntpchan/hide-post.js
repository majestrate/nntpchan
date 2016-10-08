/** hidepost.js -- hides posts from page given $things */


function get_hidden_posts() {
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
  }
}



// is a post elem an OP?
function postIsOP(elem) {
  var ds = elem.dataset;
  return ds && ds.rootmsgid == ds.msgid ;
}

function _hide_elem(elem, fade) {
  if(!fade) {
    if (elem.style) {
      elem.style.display = "none";
    } else {
      elem.style = {display: "none" };
    }
    elem.dataset.userhide = "yes";
  } else {
    $(elem).fadeOut(400, function() {
      _hide_elem(elem);
    });
  }
}

function _unhide_elem(elem) {
  $(elem).fadeIn();
  elem.dataset.userhide = "no";
}

// return true if element is hidden
function _elemIsHidden(elem) {
  return elem.dataset && elem.dataset.userhide == "yes";
}

// hide a post
function hidepost(elem, nofade) {
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
  var es = elem.getElementsByClassName("attachments");
  for (var idx = 0; idx < es.length ; idx ++ ) {
    _hide_elem(es[idx], !nofade);
  }
  es = elem.getElementsByClassName("post_body");
  for (var idx = 0; idx < es.length ; idx ++ ) {
    _hide_elem(es[idx], !nofade);
  }
  es = elem.getElementsByClassName("postheader");
  for (var idx = 0; idx < es.length ; idx ++ ) {
    _hide_elem(es[idx], !nofade);
  }
  elem.dataset.userhide = "yes";
  elem.setHideLabel("[show]");
}

// unhide a post
function unhidepost(elem) {
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
  var es = elem.getElementsByClassName("attachments");
  for (var idx = 0; idx < es.length ; idx ++ ) {
    _unhide_elem(es[idx]);
  }
  es = elem.getElementsByClassName("post_body");
  for (var idx = 0; idx < es.length ; idx ++ ) {
    _unhide_elem(es[idx]);
  }
  es = elem.getElementsByClassName("postheader");
  for (var idx = 0; idx < es.length ; idx ++ ) {
    _unhide_elem(es[idx]);
  }

  elem.dataset.userhide = "no";
  elem.setHideLabel("[hide]");
}

// hide a post given a callback that checks each post
function hideposts(check_func) {
  var es = document.getElementsByClassName("post");
  for ( var idx = 0; idx < es.length ; idx ++ ) {
    var elem = es[idx];
    if(check_func && elem && check_func(elem)) {
      hidepost(elem);
    }
  }
}

// unhide all posts given callback
// if callback is null unhide all
function unhideall(check_func) {
  var es = document.getElementsByClassName("post");
  for (var idx=0 ; idx < es.length; idx ++ ) {
    var elem = es[idx];
    if(!check_func) { unhide(elem); }
    else if(check_func(elem)) { unhide(elem); }
  }
}

// inject posthide into page
onready(function() {
  var posts = document.getElementsByClassName("post");
  for (var idx = 0 ; idx < posts.length; idx++ ) {
    console.log("inject hide: "+posts[idx].dataset.msgid);
    var inject = function (elem) {
      var hider = document.createElement("a");
      hider.setAttribute("class", "hider");
      elem.setHideLabel = function (txt) {
        var e_hider = hider;
        e_hider.innerHTML = txt;
      }
      elem.hidepost = function() {
        var e_self = elem;
        var e_hider = hider;
        hidepost(e_self);
      }
      elem.unhidepost = function() {
        var e_self = elem;
        var e_hider = hider;
        unhidepost(e_self);
      }
      elem.isHiding = function() {
        var e_self = elem;
        return _elemIsHidden(e_self);
      }
      hider.appendChild(document.createTextNode("[hide]"));
      hider.onclick = function() {
        var e_self = elem;
        if(e_self.isHiding()) {
          e_self.unhidepost();
        } else {
          e_self.hidepost();
        }
      }
      elem.appendChild(hider);
    };
    inject(posts[idx]);
  }
  // apply persiting hidden posts
  var posts = get_hidden_posts();
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
