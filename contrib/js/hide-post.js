/** hidepost.js -- hides posts from page given $things */


// is a post elem an OP?
function postIsOP(elem) {
  var ds = elem.dataset;
  return ds && ds.rootmsgid == ds.msgid ;
}

function _hide_elem(elem) {
 if (elem.style) {
    elem.style.display = "none";
  } else {
    elem.style = { display: "none" };
  }
  elem.dataset.userhide = "yes";
}

function _unhide_elem(elem) {
 if (elem.style) {
    elem.style.display = "block";
  } else {
    elem.style = { display: "block" };
  }
  elem.dataset.userhide = "no";
}

// return true if element is hidden
function _elemIsHidden(elem) {
  return elem.dataset && elem.dataset.userhide == "yes";
}

// hide a post
function hidepost(elem) {
  console.log("hidepost("+elem.dataset.msgid+")");
  
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
    _hide_elem(es[idx]);
  }
  es = elem.getElementsByClassName("post_body");
  for (var idx = 0; idx < es.length ; idx ++ ) {
    _hide_elem(es[idx]);
  }
  es = elem.getElementsByClassName("postheader");
  for (var idx = 0; idx < es.length ; idx ++ ) {
    _hide_elem(es[idx]);
  }
  elem.dataset.userhide = "yes";
  elem.setHideLabel("[show]");  
}

// unhide a post
function unhidepost(elem) {
  console.log("unhidepost("+elem.dataset.msgid+")");
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
      elem.hide = function() {
        var e_self = elem;
        var e_hider = hider;
        hidepost(e_self);
      }
      elem.unhide = function() {
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
          e_self.unhide();
        } else {
          e_self.hide();
        }
      }
      elem.appendChild(hider);
    };
    inject(posts[idx]);
  }
});
