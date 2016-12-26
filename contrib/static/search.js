



/** inject search widget */
function inject_search(elem) {
    var inner = document.createElement("div");
    inner.setAttribute("class", "search-navbar");
    var hash = document.createElement("input");
    var button = document.createElement("button");
    var input = document.createElement("input");
    var newsgroup = document.createElement("input");
    var status = document.createElement("span");
    var output = document.createElement("div");
    output.setAttribute("class", "searchresult");

    button.innerHTML = "search";
    
    function inject_search_result(r) {
        if(!r)return;
        var e = document.createElement("div");
        var a = document.createElement("a");
        a.href = r.URL;
        a.target = "_blank";
        e.appendChild(a);
        var txt = document.createTextNode(r.Message_id);
        a.appendChild(txt);
        output.appendChild(document.createElement("hr"));
        output.appendChild(e);
    }

    var search = function() {

        var h = hash.value;
        var text = input.value;
        input.value = "";
        var group = newsgroup.value;
        while(output.children.length > 0) 
            output.children[0].remove();
        var ajax = new XMLHttpRequest();
        ajax.onreadystatechange = function() {
            if (ajax.readyState == XMLHttpRequest.DONE) {
                // done
                if(ajax.status == 200) {
                    // good
                    var result = JSON.parse(ajax.responseText);
                    var num = result.length - 1;
                    if (num <= 0) {
                        status.innerHTML = "no results";
                    } else {
                        status.innerHTML = "found "+num+" results";
                        for (var idx = 0 ; idx < result.length; idx++ ) {
                            inject_search_result(result[idx]);
                        }
                    }
                } else {
                    if (ajax.status == 404) {
                        status.innerHTML = "no results";
                    } else {
                        status.innerHTML = "HTTP "+ajax.status;
                    }
                }
            } else {
                status.innerHTML = "searching... ";
            }
        };
        if (!h) {
            ajax.open("GET", "/api/find?text="+text+"&group="+group);
        } else {
            ajax.open("GET", "/api/find?hash="+h);
        }
        ajax.send();
    };

    button.onclick = function(ev) {
        search();
    };

    input.onkeydown = function(ev) {
        if (ev.keyCode == 13)
            search();
    };

    newsgroup.onkeydown = function(ev) {
        if (ev.keyCode == 13)
            search();
    };

    inner.appendChild(document.createTextNode("text: "));
    inner.appendChild(input);
    inner.appendChild(document.createTextNode("board: "));
    inner.appendChild(newsgroup);
    inner.appendChild(document.createTextNode("posthash: "))
    inner.appendChild(hash);
    inner.appendChild(button);
    inner.appendChild(status);
    
    elem.appendChild(inner);
    elem.appendChild(output);
}
