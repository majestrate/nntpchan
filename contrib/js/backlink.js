
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
  table_insert_row(tbody, document.createTextNode("Name"), [elem])
  
  // subject
  elem = document.createElement("input");
  elem.setAttribute("name", "subject");
  elem.setAttribute("value", "");
  // submit
  var submit = document.createElement("input");
  submit.setAttribute("type", "submit");
  submit.setAttribute("value", "reply");
  submit.setAttribute("class", "button");
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
  table_insert_row(tbody, document.createTextNode("Name"), [elem])
    
  table.appendChild(tbody);
  this.form.appendChild(table);
  this.elem.appendChild(this.form);
  document.body.appendChild(this.elem);
  this.board = null;
  this.roothash = null;
  this.prefix = null;
}

DynReply.prototype.update = function() {
  if (this.prefix) {
    // update captcha
    this.updateCaptcha();
    if (this.board && this.roothash) {
      // update post form
      var ref = document.getElementById("postform_reference");
      ref.setAttribute("value", this.roothash);
      this.form.action = this.prefix + "post/" + this.board;
    }
  }
}

DynReply.prototype.show = function() {
  console.log("show dynreply");
  this.update();
  this.elem.style.display = 'inline';
}

DynReply.prototype.updateCaptcha = function() {
  if (this.prefix) {
    var captcha_img = document.getElementById("captcha_img");
    captcha_img.src = this.prefix + "captcha/img";
  }
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

// reply box function
function nntpchan_reply(prefix, parent, shorthash) {
  if (prefix && parent) {
    var boardname = parent.getAttribute("boardname");
    var roothash = parent.getAttribute("root");
    var replyto = getReplyTo();
    // set target
    replyto.setBoard(boardname);
    replyto.setRoot(roothash);
    replyto.setPrefix(prefix);
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
  // dynamic post reply draggable
  var rpl = getReplyTo();
  rpl.setPrefix(prefix);
  var e = rpl.elem;
  e.setAttribute("draggable", "true");
  e.ondragend = function(ev) {
    ev.preventDefault();
    var el = document.getElementById("postform_container");
    el.setAttribute("style", "top: "+ev.y+"; left: "+ev.x+ "; position: fixed; ");
  }
}

