// call an api method
// handler(json_object) on success
// handler(null) on fail
function nntpchan_apicall(url, handler, err_handler, method, data) {
  var ajax = new XMLHttpRequest();
  ajax.onreadystatechange = function() {
    if (ajax.readyState == XMLHttpRequest.DONE ) {
      var status = ajax.status;
      var j = null;
      if (status == 200 || status == 201) {
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
  var meth = method || "GET";
  ajax.open(meth, url);
  if(data)
    ajax.send(data);
  else
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
    inject_hover_for_element(post);
  } else {
    post.setAttribute("class", "notfound post");
    post.appendChild(document.createTextNode("post not found"));
  }
  parent.appendChild(post);
}

