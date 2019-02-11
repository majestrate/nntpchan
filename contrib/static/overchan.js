
var _onreadyfuncs = [];

var onready = function(f) {
  _onreadyfuncs.push(function() {f();});
};

var ready = function() {
  for(var idx = 0; idx < _onreadyfuncs.length; idx++) _onreadyfuncs[idx]();
};

var nntpchan_mod_mark_spam = function(longhash) {
  var elem = document.getElementById(longhash);
  if(!elem) return;
  elem.dataset.spam = "yes";
  elem.innerText = "spam";
};

var nntpchan_mod_decode_ipban = function(longhash) {
  var elem = document.getElementById("post_body_" + longhash);
  if(!elem) return;
  var lines = elem.innerText.split("\n");
  console.log(lines);
  for(var i = 0; i < lines.length; ++i)
  {
    if(!lines[i])
      continue;
    if (!lines[i].startsWith("overchan-inet-ban"))
      continue;
    console.log(lines[i]);
    var parts = lines[i].split(" ");
    if(parts.length < 2) continue;
    parts = parts[1].split(":");
    if(parts.legngth < 2) continue;
    var a1 = atob(parts[0]);
    var a2 = atob(parts[1]);
    var txt = "";
    for(var idx = 0; idx < a1.length; ++idx)
    {
      txt += String.fromCharCode(a1.charCodeAt(idx) ^ a2.charCodeAt(idx));
    }
    elem.appendChild(document.createTextNode(txt + "\n"));
  }
};

var nntpchan_mod_action = function(mod_action, elem) {

  var csrf_ajax = new XMLHttpRequest();
  csrf_ajax.onreadystatechange = function() {
    if (csrf_ajax.readyState == XMLHttpRequest.DONE) {
      // get csrf token
      var csrf = csrf_ajax.getResponseHeader("X-CSRF-Token");
      // fire off ajax
      var ajax = new XMLHttpRequest();
      ajax.onreadystatechange = function() {
        if (ajax.readyState == XMLHttpRequest.DONE) {
          var status = ajax.status;
          // we gud?
          if (status == 200) {
            // yah
            var txt = ajax.responseText;
            var j = JSON.parse(txt);
            if (j.error) {
              var e = document.createTextNode(j.error);
              elem.appendChild(e);
            } else {
              if (mod_action.handle) {
                var result = mod_action.handle(j);
                if (result) {
                  elem.appendChild(result);
                }
              }
            }
          } else if (status) {
            // nah
            // http error
            elem.innerHTML = "error: HTTP "+status;
          }
          // clear input
          if (input) {
            input.value = "";
          }
        }
      }
      if (mod_action.name) {
        var url = "/mod/" + mod_action.name;
        ajax.open(mod_action.method || "GET", url);
        ajax.setRequestHeader("X-CSRF-Token", csrf);
        var data = mod_action.data;
        if (data) {
          ajax.setRequestHeader("Content-type","text/json");
          ajax.send(JSON.stringify(data));
        } else {
          ajax.send();
        }
      } else {
        alert("mod action has no name");
      }
    }
  }
  csrf_ajax.open("GET", "/mod/");
  csrf_ajax.send();
};


var nntpchan_do_admin = function(method, param, result_elem) {
  nntpchan_mod_action({
    name:"admin/"+method,
    method: ( param && "POST" ) || "GET",
    data: param
  }, result_elem);
};

var nntpchan_mod_trust_mod = function(pubkey, elem) {
  nntpchan_do_admin("pubkey.add", {pubkey: pubkey}, elem);
};

var nntpchan_mod_untrust_mod = function(pubkey, elem) {
  nntpchan_do_admin("pubkey.del", {pubkey: pubkey}, elem);
};

var nntpchan_mod_commit_spam = function(elem) {
  var formdata = new FormData();
  var posts = document.getElementsByClassName("post");
  var spams = [];
  for (var idx = 0; idx < posts.length; idx ++)
  {
    if(posts[idx].dataset.spam == "yes")
    {
      spams.push(posts[idx].dataset.msgid);
    }
  }
  formdata.set("spam", spams.join(","));
  var jax = new XMLHttpRequest();
  jax.onreadystatechange = function() {
    if(jax.readyState == 4)
    {
      if(jax.status == 200)
      {
        
        var ajax = new XMLHttpRequest();
        ajax.onreadystatechange = function() {
        if(ajax.readyState == 4)
          {
            if(ajax.status == 200)
            {
              // success (?)
              var j = JSON.parse(ajax.responseText);
              if(j.error)
              {
                elem.innerText = "could not mark as spam: " + j.error;
              }
              else
              {
                elem.innerText = "OK: marked as spam";
              }
            }
            else 
            {
              elem.innerText = "post not marked as spam on server: "+ ajax.statusText;
            }
          }
        };
        ajax.open("POST", "/mod/spam")
        ajax.setRequestHeader("X-CSRF-Token", jax.getResponseHeader("X-CSRF-Token"));
        ajax.send(formdata);
      }
      else
      {
        elem.innerText = "failed to moderate, not logged in";
      }
    }
  };
  jax.open("GET", "/mod/");
  jax.send();
};

var nntpchan_mod_delete = function(longhash) {
  var elem = document.getElementById(longhash);
  var ajax = new XMLHttpRequest();
  ajax.onreadystatechange = function() {
    if(ajax.readyState == 4)
    {
      if(ajax.status == 200)
      {
        // success (?)
        var j = JSON.parse(ajax.responseText);
        if(j.deleted.length > 0)
        {
          elem.appendChild(document.createTextNode("deleted: " + j.deleted.join(",")));
        }
      }
      else 
      {
        elem.innerHTML = "post not deleted from server: "+ ajax.statusText;
      }
    }
  };
  ajax.open("GET", "/mod/del/"+longhash);
  ajax.send();
  elem.innerHTML = "";
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

var reloadThreadJSON = function(ourPost) {
  var url = window.location.pathname + "/json";
  var ajax = new XMLHttpRequest();
  ajax.onreadystatechange = function() {
    if(ajax.readyState == 4) {
      if(ajax.status == 404) {
        console.log("thread gone");
      } else if (ajax.status == 200) {
        var rootelem = document.getElementById("thread_"+window.location.pathname.split("/")[2]);
        var posts = JSON.parse(ajax.responseText);
        for(var idx = 0; idx < posts.length; idx ++ )
        {
          var id = posts[idx].HashLong;
          var e = document.getElementById(id);
          if(!e) {
            e = document.createElement("div");
            e.innerHTML = posts[idx].PostMarkup;
            rootelem.appendChild(e.childNodes[0]);
            e.remove();
            if(ourPost && posts[idx].Message_id == ourPost) {
              // focus on our post
              window.location.hash = id;
            }
          }
        }
      }
    }
  };
  ajax.open("GET", url);
  ajax.send();
}

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
          form.reset();
          cb(null, j);
        } else if (ajax.status == 200) {
          cb(err, j);
        } else {
          cb("http "+ajax.status, j);
        }
      } else {
        elem.value += ".";
      }
    };
    var data = new FormData();
    data.append("message", document.getElementById("comment").value);
    var inputs = form.getElementsByTagName("input");
    for(var idx = 0; idx < inputs.length; idx++)
    {
      var input = inputs[idx];
      console.log(input);
      if(input.files)
      {
        for (var i =0 ; i < input.files.length; i++)
        {
          var file = input.files[i];
          data.append(input.name, file, file.name);
        }
      }
      else if(input.name)
        data.append(input.name, input.value);
    }
    console.log("posting...");
    ajax.open("POST", form.action+"/json");
    ajax.send(data);
  };
  var elems = document.getElementsByClassName("postbutton");
  if(elems && elems.length > 0 && elems[0]) {
    var e = elems[0];
    var parent = e.parentElement;
    var origText = e.value;
    e.remove();
    e = document.createElement("input");
    e.type = "button";
    parent.appendChild(e);
    e.value = origText;
    e.onclick = function(ev) {
      console.log("clicked post");
      e.disabled = true;
      e.value = "posting ";
      submitPost(document.forms[0], e, function(err, j) {
        if(err) {
          var captcha = document.getElementById("captcha_solution");
          if(captcha) {
            captcha.value = "";
          }
        }
        var msg = err || "posted";
        console.log(msg, j.url);
        e.value = msg;
        if(window.location.pathname === j.url) {
          reloadThreadJSON(j.message_id);
        } else if (j && j.url) {
          // do redirect
          window.location.pathname = j.url;
          return;
        }
        var img = document.getElementById("captcha_img");
        if (img) {
          reloadImg(img);
        }
        setTimeout(function() {
          e.disabled = false;
          e.value = origText;
        }, 1000);
        
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

// inline reply expand
onready(function() {
  return;
  var fetchpost = function(url, cb)
  {
    var parts = url.split("#");
    var base = parts[0];
    var h = parts[1];
    var ajax = new XMLHttpRequest();
    ajax.onreadystatechange = function()
    {
      if(ajax.readyState == 4)
      {
        if(ajax.status == 200)
        {
          var j = JSON.parse(ajax.responseText);
          for(var idx =0 ; idx < j.length; idx ++)
          {
            if (j[idx].HashLong == h)  {
              cb(j[idx]);
              return;
            }
          }
          cb(null);
        }
        else
        {
          cb(null);
        }
      }
    };
    ajax.open("GET", base +"json");
    ajax.send();
  };
  
  var elems = document.getElementsByClassName("backlink");
  var inj = function(elem)
  {
    var showhover = function(parent, url, id)
    {
      fetchpost(url, function(post) {
        var wrapper = document.createElement("div");
        var e = document.createElement("div");
        e.innerHTML = post.PostMarkup || "post not found"
        wrapper.setAttribute('id', id);
        wrapper.setAttribute("class", "hover");
        wrapper.appendChild(e);
        var cl = document.createElement("div");
        cl.innerHTML = "[X]";
        cl.onclick = function() {
          wrapper.remove();
        };
        wrapper.appendChild(cl);
        parent.appendChild(wrapper);
      });
    };

    var hidehover = function(parent, id)
    {
      var hover = document.getElementById(id);
        if(hover) hover.remove();
    };

   
    var parent = elem.parentNode.parentNode.id;
    var wrapper = document.createElement("div");
    elem.parentNode.insertBefore(wrapper, elem);
    var el = elem.cloneNode(true);
    elem.remove();
    var parts = el.href.split("#");
    wrapper.appendChild(el);

    var id = "hover_"+parts[1]+"_"+parent;
    console.log(id);
    el.onpointerenter = function() {
      showhover(wrapper, el.href, id); 
    };

    el.onpointerleave = function()
    {
      hidehover(wrapper, id);
    };
  }
  for (var idx = 0 ; idx < elems.length ; idx++) inj(elems[idx]);
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
  
  var posts = document.getElementsByClassName("post");
  for (var idx = 0 ; idx < posts.length; idx++ ) {
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
