//
// nntpchan.js -- frontend ui niceness
//


// insert a backlink for a post given its short hash
function nntpchan_backlink(shorthash) {
  var elem = document.getElementById("postform_message");
  if ( elem )
  {
    elem.value += ">>" + shorthash.substr(0,10) + "\n";
  }
}

function get_storage() {
  var st = null;
  if (window.localStorage) {
    st = window.localStorage;
  } else if (localStorage) {
    st = localStorage;
  }
  return st;
}

function enable_theme(prefix, name) {
  if (prefix && name) {
    var theme = document.getElementById("current_theme");
    if (theme) {
      theme.href = prefix + "static/"+ name + ".css";
      var st = get_storage();
      st.nntpchan_prefix = prefix;
      st.nntpchan_theme = name;
    }
  }
}


// call an api method
// handler(json_object) on success
// handler(null) on fail
function nntpchan_apicall(url, handler, err_handler) {
  var ajax = new XMLHttpRequest();
  ajax.onreadystatechange = function() {
    if (ajax.readyState == XMLHttpRequest.DONE ) {
      var status = ajax.status;
      var j = null;
      if (status == 200) {
        // found
        try {
          j = JSON.parse(ajax.responseText);
        } catch (e) {} // ignore parse error
      } else if (status == 410) {
        if (err_handler) {err_handler("cannot fetch post: api disabled");}
        return;
      }
      handler(j);
    }
  };
  ajax.open("GET", url);
  ajax.send();
}

// build post from json
// inject into parent
// if j is null then inject "not found" post
function nntpchan_buildpost(parent, j) {
  var post = document.createElement("div");
  if (j) {
    // huehuehue
    post.innerHTML = j.PostMarkup;
  } else {
    post.setAttribute("class", "notfound post");
    post.appendChild(document.createTextNode("post not found"));
  }
  parent.appendChild(post);
}

// inject post hover behavior
function inject_hover(prefix, el, parent) {
  if (!prefix) { throw "prefix is not defined"; }
  console.log(el, parent);
  var timeout;
  var idx = -2;

  var linkhash = el.backlinkhash;
  
  var elem = document.createElement("span");
  elem.setAttribute("class", "backlink_rewritten");
  elem.appendChild(document.createTextNode(">>"+linkhash.substr(0,10)));

  parent.appendChild(elem);
  parent.removeChild(el);
  
  elem.onclick = function(ev) {
    if(elem.backlink) {
      nntpchan_apicall(prefix+"api/find?hash="+linkhash, function(j) {
        var wrapper = document.createElement("div");
        wrapper.setAttribute("class", "hover "+linkhash);
        if (j == null) {
          // not found?
          wrapper.appendChild(document.createTextNode("not found"));
        } else {
          // wrap backlink
          nntpchan_buildpost(wrapper, j);
        }
        elem.appendChild(wrapper);
        elem.backlink = false;
      }, function(msg) {
        var wrapper = document.createElement("div");
        wrapper.setAttribute("class", "hover "+linkhash);
        wrapper.appendChild(document.createTextNode(msg));
        elem.appendChild(wrapper);
        elem.backlink = false;
      });
    } else {
      var elems = document.getElementsByClassName(linkhash);
      if (!elems) throw "bad state, no backlinks open?";
      for (var idx = 0 ; idx < elems.length; idx ++ ) {
        elems[idx].parentNode.removeChild(elems[idx]);
      }
      elem.backlink = true;
    }
  };
  elem.backlink = true;
}

function livechan_got_post(widget, j) {
  // do scroll
  while (widget.children.length > 20) {
    // remove top element
    widget.removeChild(widget.children[0]);
  }
  nntpchan_buildpost(widget, j);
  // scroll to bottom
  widget.scrollTop = widget.scrollHeight;
}

// inject post form into an element
function inject_postform(prefix, parent) {
  
}

// inject livechan widget into parent
function inject_livechan_widget(prefix, parent) {
  if ( "WebSocket" in window ) {
    var url = "ws://"+document.location.host+prefix+"live";
    if ( document.location.protocol == "https:" ) {
      url = "wss://"+document.location.host+prefix+"live";
    }
    var socket = new WebSocket(url);
    var progress = function(str) {
      parent.innerHTML = "<pre>livechan: "+str+"</pre>";
    };
    progress("initialize");
    socket.onopen = function () {
      progress("streaming");
    }
    socket.onmessage = function(ev) {
      var j = null;
      try {
        j = JSON.parse(ev.data);
      } catch(e) {
        // ignore
      }
      if (j) {
        console.log(j);
        livechan_got_post(parent, j);
      }
    }
    socket.onclose = function(ev) {
      progress("connection closed");
    }
  } else {
    parent.innerHTML = "<pre>livechan mode requires websocket support</pre>";
    setTimeout(function() {
      parent.innerHTML = "";
    }, 5000);
  }
}

function ukko_livechan(prefix) {
  var ukko = document.getElementById("ukko_threads");
  if (ukko) {
    // remove children
    ukko.innerHTML = "";
    inject_livechan_widget(prefix, ukko);
  }
}

var banner_count = 3;

// inject a banner into an element
function nntpchan_inject_banners(elem, prefix) {
  var n = Math.floor(Math.random() * banner_count);
  var banner = prefix + "static/banner_"+n+".jpg";
  var e = document.createElement("img");
  e.src = banner;
  e.id = "nntpchan_banner";
  elem.appendChild(e);
}

function init(prefix) {
  // inject posthover ...
  var elems = document.getElementsByClassName("backlink");
  // ... for backlinks
  for ( var idx = 0 ; idx < elems.length ; idx ++ ) {
    // uncomment to do reply hover
    //inject_hover(prefix, elems[idx], elems[idx].parentNode);
  }
}
  
// apply themes
var st = get_storage();
enable_theme(st.nntpchan_prefix, st.nntpchan_theme);
