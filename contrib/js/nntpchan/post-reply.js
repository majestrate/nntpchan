/**
 post reply box
 */


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
 reply box
 */
function ReplyBox() {
    var elem = document.createElement("div");
    this.elem = elem;
    elem.setAttribute("class", "shadow shadow-box");
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
    this.name = elem;
    span.appendChild(elem);
    // error message
    var err_elem = document.createElement("span");
    span.appendChild(err_elem);
    this._error = err_elem;
    table_insert_row(tbody, document.createTextNode("Name"), [span]);
    
    // subject
    elem = document.createElement("input");
    elem.setAttribute("value", "");
    this.subject = elem;
    // submit
    var submit = document.createElement("button");
    submit.innerHTML = "reply";
    table_insert_row(tbody, document.createTextNode("Subject"), [elem]);
    this.submit = submit;
    
    // Comment
    elem = document.createElement("textarea");
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
    table_insert_row(tbody, document.createTextNode("Post"), [submit]);
    this.captcha_solution = elem;
    table.appendChild(tbody);
    this.elem.appendChild(table);
    document.body.appendChild(this.elem);
    $(this.elem).css("position", "fixed");
    $(this.elem).css("display", "none");
    this._open = false;
}

ReplyBox.prototype.result = function(msg, color) {
    var self = this;
    self._error.innerHTML = "";
    $(self._error).css("color", color);
    self._error.appendChild(document.createTextNode(msg));
    setTimeout(function() {
        $(self._error).fadeOut(1000, function() {
            self._error.innerHTML = "";
        });
    }, 1000);
}

ReplyBox.prototype.visible = function() {
    return this._open != false;
}


ReplyBox.prototype.makePost = function(info) {
    var self = this;
    var data = new FormData();
    data.append("name", self.name.value);
    data.append("subject", self.subject.value);
    data.append("captcha", self.captcha_solution.value);
    data.append("message", self.message.value);
    $(self.files.files).each(function(_, f) {
        data.append("attachment_uploaded", f);
    });
    data.append("reference", info.reference);
    return data;
}

ReplyBox.prototype.clear = function() {
    var self = this;
    self.name.value = "";
    self.subject.value = "";
    self.message.value = "";
    self.captcha_solution.value = "";
}

ReplyBox.prototype.reload = function() {
    var self = this;
    self.captcha_img.src = "/captcha/img?" + new Date().getTime();
}

ReplyBox.prototype.show = function(info) {
    var self = this;
    self.reload();
    self._open = true;
    console.log("reply box show for "+info.reference);
    $(self.elem).css("display", "inline-block");
    var off = $(info.elem).offset();
    $(self.elem).offset({
        top: off.top,
        left: $(info.elem).width() + off.left
    });
    self.submit.onclick = function(ev) {
        var post = self.makePost(info);
        var a = $.ajax({
            data: post,
            processData: false,
            contentType: false,
            url: info.url,
            method: "POST",
            dataType: "json"
        }).success(function(data, status, xhr) {
            if(data.message_id) {
                self.result("posted as "+data.message_id, "green");
                self.clear();
                setTimeout(function() {
                    self.hide();
                    if(data.url)
                        window.location = data.url;
                }, 1000);
            } else {
                self.result("error: " + data.error, "red");
            }
        }).fail(function() {
            self.result("request failed", "red");
        }).done(function() {
            self.reload();
        });
    };
}

ReplyBox.prototype.hide = function() {
    var self = this;
    if(!self.visible()) return;
    self._open = false;
    $(self.elem).fadeOut(400, function() {
        self.submit.onclick = function(ev) {};
    });
}

onready(function(){
    var replyBox = new ReplyBox();
    replyBox.hide();
    document.reply = replyBox;
    $(".post").each(function(_, elem) {
        var replyInfo = {
            show: false,
            url: (prefix || "/") + "post/" + elem.dataset.newsgroup + "/json",
            reference: elem.dataset.rootmsgid,
            board: elem.dataset.newsgroup,
            elem: elem
        };
        var elems = elem.getElementsByClassName("postreply");
        if(elems && elems[0]) {
            var e = elems[0].children[1];
            console.log("inject reply box into "+e);
            e.onclick = function(ev) {
                if(!replyInfo.show) {
                    // is hidden, show reply
                    replyBox.show(replyInfo);
                } else {
                    // is open, hide reply
                    replyBox.hide();
                }
                replyInfo.show = ! replyInfo.show;
                ev.preventDefault();
            }
        }
    });
});
