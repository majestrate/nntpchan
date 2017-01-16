
function getReplyTo() {
    if(!document.dynreply) { 
        document.dynreply = new DynReply(document.body);
    }
    var e = document.dynreply.elem;
    e.style.position = "fixed";
    e.style.left = document.dynreply.x+ "px";
    e.style.top = document.dynreply.y + "px";
    e.setAttribute("class", "shadow shadow-box");
    return document.dynreply;
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
function DynReply(rootElem) {
    
    var elem = document.createElement("div");
    this.elem = elem;
    // reference
    elem = document.createElement("input");
    elem.name = "reference";
    elem.type = "hidden";
    this.elem.appendChild(elem);
    
    var table = document.createElement("table");
    table.setAttribute("class", "postform");
    var tbody = document.createElement("tbody");

    var span = document.createElement("span");
    // name 
    elem = document.createElement("input");
    elem.setAttribute("name", "name");
    elem.setAttribute("value", "Anonymous");
    elem.setAttribute("id", "postform_name");
    span.appendChild(elem);
    // error message
    var err_elem = document.createElement("span");
    err_elem.setAttribute("id", "postform_msg");
    span.appendChild(err_elem);
    this._error = err_elem;
    table_insert_row(tbody, document.createTextNode("Name"), [span]);
    
    // subject
    elem = document.createElement("input");
    elem.setAttribute("value", "");
    this.subject = elem;
    // submit
    var submit = document.createElement("input");
    submit.setAttribute("value", "reply");
    submit.setAttribute("class", "button");
    table_insert_row(tbody, document.createTextNode("Subject"), [elem, submit]);
    this.submit = submit;
    
    // Comment
    elem = document.createElement("textarea");
    elem.setAttribute("id", "postform_message");
    elem.setAttribute("cols", "40");
    elem.setAttribute("rows", "5");
    table_insert_row(tbody, document.createTextNode("Comment"), [elem]);
    this.message = elem;
    
    // file
    elem = document.createElement("input");
    elem.setAttribute("class", "postform_attachment");
    elem.setAttribute("type", "file");
    elem.setAttribute("multiple", "multiple");
    this.files = elem;
    table_insert_row(tbody, document.createTextNode("Files"), [elem]);

    // captcha
    elem = document.createElement("img");
    elem.alt = "captcha";
    table_insert_row(tbody, document.createTextNode("Captcha"), [elem]);
    this.captcha_img = elem;

    // captcha solution
    elem = document.createElement("input");
    elem.name = "captcha";
    elem.autocomplete = "off";
    table_insert_row(tbody, document.createTextNode("Solution"), [elem]);
    this.captcha_solution = elem;
    table.appendChild(tbody);
    this.elem.appendChild(table);
    this.board = null;
    this.rootmsg = null;
    this.prefix = null;
    this.url = null;
    this.x = 1;
    this.y = 1;
    rootElem.appendChild(this.elem);
}

DynReply.prototype.show = function() {
    console.log("show dynreply");
    this.updateCaptcha();
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
    // reset value
    this.captcha_solution.value = "";
}

// clear postform elements
DynReply.prototype.clearPostbox = function() {
    this.subject.value = "";
    this.message.value = "";
    this.files.value = null;
}

DynReply.prototype.post = function(cb, err_cb) {
    if (this.url) {
        var data = new FormData();
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
        if (!this.captcha_img) {
            this.captcha_img = document.createElement("img");
            this.captcha_img.src = this.prefix + "captcha/img";
        }
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

DynReply.prototype.setRoot = function(rootmsg) {
    if (rootmsg) {
        this.rootmsg = rootmsg;
    }
}

DynReply.prototype.showError = function(msg) {
    console.log("error in dynreply: "+msg);
    this._error.setAttribute("class", "error message");
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
    var elem = document.getElementById("postform_message");
    if ( elem )
    {
        elem.value += ">>" + shorthash.substr(0,10) + "\n";
    }
}


function init(prefix, enabled) {
    if(!enabled) return;
    var rpl  = getReplyTo();
    rpl.setPrefix(prefix);
    rpl.updateCaptcha();
    
    // position replyto widget
    var e = rpl.elem;
    var mouseDownX, mouseDownY;
    
    var $dragging = null;

    $(rpl.elem).on("mousemove", function(ev) {
        if ($dragging) {
            var x = ev.pageX - $(this).width() / 2,
                y = ev.pageY - $(this).height() / 2;

            $(e).offset({
                top: y,
                left: x
            });
        }
    });


    $(rpl.elem).on("mousedown", e, function (ev) {
        if (ev.button == 0 ) {
            $dragging = $(rpl.elem);
        }
    });

    $(rpl.elem).on("mouseup", function (e) {
        $dragging = null;
    });
    
    // add replyto post handlers
    var postit = function() {
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
}

