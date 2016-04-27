// insert a backlink for a post given its short hash
function nntpchan_backlink(shorthash) {
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
}

