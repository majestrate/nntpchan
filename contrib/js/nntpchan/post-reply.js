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
    submit.setAttribute("value", "reply");
    submit.setAttribute("class", "button");
    table_insert_row(tbody, document.createTextNode("Subject"), [elem, submit]);
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
    this.captcha_solution = elem;
    table.appendChild(tbody);
    this.elem.appendChild(table);
    document.body.appendChild(this.elem);
}

ReplyBox.prototype.result = function(msg, color) {
    var self = this;
    self._error.innerHTML = "";
    $(self._error).css("color", color);
    self._error.appendChild(document.createTextNode(msg));
    setTimeout(function() {
        $(self._error).fadeout(1000, function() {
            self._error.innerHTML = "";
        });
    }, 1000);
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
    self.captcha_img.src = "/captcha/img?" + new Date().getTime();
}

ReplyBox.prototype.show = function(info) {
    console.log("reply box show for "+info.reference);
    var self = this;
    $(self.elem).css("display", "inline-block");
    $(self.elem).css("position", "fixed");
    var off = $(info.elem).offset();
    $(self.elem).offset({
        top: off.top,
        right: off.left
    });
    self.submit.onclick = function(ev) {
        $.ajax({
            data: self.makePost(info),
            url: info.url,
            method: "POST",
            dataType: "json"
        }).success(function(data, status, xhr) {
            if(xhr.statusCode == 201) {
                self.result("posted as "+data.message_id, "green");
                self.clear();
                self.submit.onclick = function(ev) {};
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
    $(self.elem).css("display", "none");
    self.submit.onclick = function(ev) {};
}

onready(function(){
    var replyBox = new ReplyBox();
    replyBox.hide();
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
