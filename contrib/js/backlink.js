
var dynreply;

function getReplyTo() {
  if(!dynreply) {
    var e = document.getElementById("postform_container");
    if (e) {
      // use existing postform
      dynreply = new DynReply(e);
    } else {
      // build a new postform
      dynreply = new DynReply();
    }
  }
  return dynreply;
}

function table_insert_row(table, header, items) {
  var tr = document.createElement("tr");
  // insert header element
  var th = document.createElement("th");
  th.appendChild(header);
  tr.appendChild(th);
  // insert the rest of the elements
  for (var idx = 0; idx < items.length; idx ++ ) {
    var elem = document.createElement("td");
    elem.appendChild(items[idx]);
    tr.appendChild(elem);
  }
  table.appendChild(tr);
}

/**
   build dynamic reply box
*/
function DynReply(existingElem) {
  if (existingElem) {
    // wrap existing post form
    this.elem = existingElem;
    this.form = this.elem.querySelector("form");
    this._error = document.getElementById("postform_msg");
    this.url = this.form.action + "?t=json";
    return;
  }

  // build new post form
  
  var elem = document.createElement("div");
  elem.setAttribute("id", "postform_container");
  this.elem = elem;
  // build post form
  this.form = document.createElement("form");
  this.form.enctype = "multipart/form-data";
  this.form.name = "post";
  this.form.method = "post";
  // reference
  elem = document.createElement("input");
  elem.setAttribute("id", "postform_reference");
  elem.name = "reference";
  elem.type = "hidden";
  this.form.appendChild(elem);
  
  var table = document.createElement("table");
  table.setAttribute("class", "postform");
  var tbody = document.createElement("tbody");

  // name 
  elem = document.createElement("input");
  elem.setAttribute("name", "name");
  elem.setAttribute("value", "Anonymous");
  elem.setAttribute("id", "postform_name");
  var err_elem = document.createElement("span");
  err_elem.setAttribute("id", "postform_msg");
  this._error = err_elem;
  table_insert_row(tbody, document.createTextNode("Name"), [elem, err_elem])
  
  // subject
  elem = document.createElement("input");
  elem.setAttribute("name", "subject");
  elem.setAttribute("value", "");
  elem.setAttribute("id", "postform_subject");
  // submit
  var submit = document.createElement("input");
  submit.setAttribute("value", "reply");
  submit.setAttribute("class", "button");
  submit.setAttribute("type", "submit");
  submit.setAttribute("id", "postform_submit");
  table_insert_row(tbody, document.createTextNode("Subject"), [elem, submit]);

  
  // Comment
  elem = document.createElement("textarea");
  elem.setAttribute("id", "postform_message");
  elem.setAttribute("name", "message");
  elem.setAttribute("cols", "40");
  elem.setAttribute("rows", "5");
  table_insert_row(tbody, document.createTextNode("Comment"), [elem]);
  
  // file
  elem = document.createElement("input");
  elem.setAttribute("class", "postform_attachment");
  elem.setAttribute("id", "postform_attachments");
  elem.setAttribute("type", "file");
  elem.setAttribute("name", "attachment_uploaded");
  elem.setAttribute("multiple", "multiple");
  table_insert_row(tbody, document.createTextNode("Files"), [elem]);

  // dubs
  elem = document.createElement("input");
  elem.setAttribute("type", "checkbox");
  elem.setAttribute("name", "dubs");
  table_insert_row(tbody, document.createTextNode("Get Dubs"), [elem]);

  // captcha
  elem = document.createElement("img");
  elem.setAttribute("id", "captcha_img");
  elem.alt = "captcha";
  table_insert_row(tbody, document.createTextNode("Captcha"), [elem]);

  // captcha solution
  elem = document.createElement("input");
  elem.name = "captcha";
  elem.autocomplete = "off";
  elem.setAttribute("id", "captcha_solution");
  table_insert_row(tbody, document.createTextNode("Solution"), [elem])
    
  table.appendChild(tbody);
  this.form.appendChild(table);
  this.elem.appendChild(this.form);
  document.body.appendChild(this.elem);
  this.board = null;
  this.roothash = null;
  this.prefix = null;
  this.url = null;
}

DynReply.prototype.moveTo = function(x,y) {
  x = window.screenX - x ;
  this.elem.setAttribute("style", "top: "+y+"px; right: "+x+"px;");
}

DynReply.prototype.update = function() {
  if (this.prefix) {
    // update captcha
    this.updateCaptcha();
    if (this.board && this.roothash) {
      // update post form
      var ref = document.getElementById("postform_reference");
      ref.setAttribute("value", this.roothash);
      this.url = this.prefix + "post/" + this.board + "?t=json";
    }
  }
}

DynReply.prototype.show = function() {
  console.log("show dynreply");
  this.update();
  this.elem.style.display = 'inline';
}

DynReply.prototype.hide = function() {
  console.log("hide dynreply");
  this.elem.style.display = "none";
}

// clear all fields
DynReply.prototype.clear = function() {
  this.clearSolution();
  this.clearPostbox();
}


// clear captcha solution
DynReply.prototype.clearSolution = function() {
  var e = document.getElementById("captcha_solution");
  // reset value
  e.value = "";
}

// clear postform elements
DynReply.prototype.clearPostbox = function() {
  var e = document.getElementById("postform_subject");
  e.value = "";
  e = document.getElementById("postform_message");
  e.value = "";
  e = document.getElementById("postform_attachments");
  e.value = null;
}

DynReply.prototype.post = function(cb, err_cb) {
  if (this.url && this.form) {
    var data = new FormData(this.form);
    var ajax = new XMLHttpRequest();
    ajax.onreadystatechange = function(ev) {
      if (ajax.readyState == XMLHttpRequest.DONE) {
        var j = null;
        try {
          j = JSON.parse(ajax.responseText);
          cb(j);
        } catch (e) {
          if(err_cb) {
            err_cb(e);
          }
        }
      }
    }
    ajax.open("POST", this.url);
    ajax.send(data);
  }
}

DynReply.prototype.updateCaptcha = function() {
  if (this.prefix) {
    var captcha_img = document.getElementById("captcha_img");
    captcha_img.src = this.prefix + "captcha/img";
  }
  this.clearSolution();
}

DynReply.prototype.setPrefix = function(prefix) {
  this.prefix = prefix;
}

DynReply.prototype.hide = function() {
  this.elem.style.display = 'none';
}


DynReply.prototype.setBoard = function(boardname) {
  if (boardname) {
    this.board = boardname;
  }
}

DynReply.prototype.setRoot = function(roothash) {
  if (roothash) {
    this.roothash = roothash;
  }
}

DynReply.prototype.showError = function(msg) {
  console.log("error in dynreply: "+msg);
  this._error.setAttribute("class", "error");
  this._error.appendChild(document.createTextNode(msg));
  this.updateCaptcha();
}

DynReply.prototype.showMessage = function(msg) {
  this._error.setAttribute("class", "message");
  this._error.innerHTML = "";
  this._error.appendChild(document.createTextNode(msg));
  var e = this._error;
  setTimeout(function() {
    // clear it
    e.innerHTML = "";
  }, 2000);
}


// reply box function
function nntpchan_reply(parent, shorthash) {
  if (parent) {
    var boardname = parent.getAttribute("boardname");
    var roothash = parent.getAttribute("root");
    var replyto = getReplyTo();
    // set target
    replyto.setBoard(boardname);
    replyto.setRoot(roothash);
    // show it
    replyto.show();
  }
  var elem = document.getElementById("postform_message");
  if ( elem )
  {
      elem.value += ">>" + shorthash.substr(0,10) + "\n";
  }
}

// inject post hover behavior
function inject_hover(prefix, el, parent) {
  if (!prefix) { throw "prefix is not defined"; }
  var linkhash = el.getAttribute("backlinkhash");
  if (!linkhash) { throw "linkhash undefined"; }
  console.log("rewrite linkhash "+linkhash);

  var elem = document.createElement("span");
  elem.setAttribute("class", "backlink_rewritten");
  elem.appendChild(document.createTextNode(">>"+linkhash.substr(0,10)));
  if (!parent) {
    parent = el.parentNode;
  }
  parent.removeChild(el);
  parent.appendChild(elem);
  
  elem.onclick = function(ev) {
    if(parent.backlink) {
      nntpchan_apicall(prefix+"api/find?hash="+linkhash, function(j) {
        var wrapper = document.createElement("div");
        wrapper.setAttribute("class", "hover "+linkhash);
        if (j == null) {
          // not found?
          wrapper.setAttribute("class", "hover notfound-hover "+linkhash);
          wrapper.appendChild(document.createTextNode("not found"));
        } else {
          // wrap backlink
          nntpchan_buildpost(wrapper, j);
        }
        parent.appendChild(wrapper);
        parent.backlink = false;
      }, function(msg) {
        var wrapper = document.createElement("div");
        wrapper.setAttribute("class", "hover "+linkhash);
        wrapper.appendChild(document.createTextNode(msg));
        parent.appendChild(wrapper);
        parent.backlink = false;
      });
    } else {
      var elems = document.getElementsByClassName(linkhash);
      if (!elems) throw "bad state, no backlinks open?";
      for (var idx = 0 ; idx < elems.length; idx ++ ) {
        elems[idx].parentNode.removeChild(elems[idx]);
      }
      parent.backlink = true;
    }
  };
  parent.backlink = true;
}

// inject post hover for all backlinks in an element
function inject_hover_for_element(elem) {  
  var elems = elem.getElementsByClassName("backlink");
  var ls = [];
  var l = elems.length;
  for ( var idx = 0 ; idx < l ; idx ++ ) {
    var e = elems[idx];
    ls.push(e);
  }
  for( var elem in ls ) {
      inject_hover(prefix, ls[elem]);
  }
}

function init(prefix) {
  // inject posthover ...
  inject_hover_for_element(document);
  // initialize replyto widget
  var rpl = getReplyTo();
  rpl.setPrefix(prefix);

  // position replyto widget
  var e = rpl.elem;
  e.setAttribute("draggable", "true");
  var mouseDownX, mouseDownY;

  var originalX = window.screenX - 150;
  var originalY = 10;
  rpl.moveTo(originalX, originalY);

  e.addEventListener("dragstart", function(ev) {
    mouseDownX = ev.clientX;
    mouseDownY = ev.clientY;
    if (!ev.ctrlKey) {
      ev.preventDefault();
    }
  }, false);
  
  e.addEventListener("dragend", function(ev) {
    var x = originalX + ev.clientX - mouseDownX;
    var y = originalY + ev.clientY - mouseDownY;
    x -= window.screenLeft;
    y -= window.screenTop;
    rpl.moveTo(x, y);
    originalX = x;
    originalY = y;
  }, false);

  // add replyto post handlers
  e = document.getElementById("postform_submit");
  var postit = function() {
    var f = document.querySelector("form");
    // do ajax request to post data
    var r = getReplyTo();
    r.showMessage("posting... ");
    r.post(function(j) {
      if(j.error) {
        // an error happened
        r.showError(j.error);
      } else {
        // we're good
        r.showMessage("posted :^)");
        r.updateCaptcha();
        r.clear();
      }
    }, function(err) {
      r.showError(err);
      r.clearSolution();
    });
  }
  var f = document.querySelector("form");
  f.onsubmit = function() {
    postit();
    return false;
  }
  
}

