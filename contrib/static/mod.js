/*
 * mod.js, moderator page js stuff
 */


// TODO: implement mod panel all the way

document.onload = function(ev) {
  // populate the mod page with stuff
}

function get_longhash(str) {
  var idx = str.indexOf("#") + 1;
  if ( idx > 0 ) {
    str = str.substr(idx);
  }
  console.log(str);
  return str;
}

// handle ban command
function nntpchan_ban() {
  nntpchan_mod({
    parser: get_longhash,
    name: "ban",
    handle: function(j) {
      if (j.banned) {
        return document.createTextNode(j.banned);
      }
    }
  });
}

function nntpchan_unban() {
  nntpchan_mod({
    name: "unban",
    handle: function(j) {
      if (j.result) {
        return document.createTextNode(j.result);
      }
    }
  })
}

function get_board_target() {
  var e = document.getElementById("nntpchan_board_target");
  return e.value;
}

function get_key_target() {
  var e = document.getElementById("nntpchan_key_target");
  return e.value;
}

function nntpchan_key_del() {
  nntpchan_admin("pubkey.del", {
    pubkey: get_key_target()
  });
}

function nntpchan_key_add() {
  nntpchan_admin("pubkey.add", {
    pubkey: get_key_target()
  });
}

function get_nntp_username() {
  var e = document.getElementById("nntpchan_nntp_username");
  return e.value;  
}

function get_nntp_passwd() {
  var e = document.getElementById("nntpchan_nntp_passwd");
  return e.value;  
}

function nntpchan_admin_nntp(method) {
  nntpchan_admin(method, {
    username: get_nntp_username(),
    passwd: get_nntp_passwd()
  })
}

function nntpchan_admin_board(method) {
  nntpchan_admin(method, {
    newsgroup: get_board_target()
  })
}

function nntpchan_admin(method, param, handler_cb, result_elem) {
  if (handler_cb) {
    // we got a handler already set
  } else {
    // no handler set
    var handler_cb = function(j) {
      if (j.result) {
        return document.createTextNode(j.result);
      } else {
        return "nothing happened?";
      }
    }
  }
  nntpchan_mod({
    name:"admin",
    parser: function(target) {
      return method;
    },
    handle: handler_cb,
    method: ( param && "POST" ) || "GET",
    data: param
  }, result_elem)
}


// handle delete command
function nntpchan_delete() {
  nntpchan_mod({
    parser: get_longhash,
    name: "del",
    handle: function(j) {
      var elem = document.createElement("div");
      if (j.deleted) {
        for ( var idx = 0 ; idx < j.deleted.length ; idx ++ ) {
          var msg = "deleted: " + j.deleted[idx];
          var e = document.createTextNode(msg);
          var el = document.createElement("div");
          el.appendChild(e);
          elem.appendChild(el);
        }
      }
      if (j.notdeleted) {
        for ( var idx = 0 ; idx < j.notdeleted.length ; idx ++ ) {
          var msg = "not deleted: " + j.notdeleted[idx];
          var e = document.createTextNode(msg);
          var el = document.createElement("div");
          el.appendChild(e);
          elem.appendChild(el);
        }
      }
      return elem;
    }
  });
}

function nntpchan_mod(mod_action, result_elem) {

  // get the element
  var input = document.getElementById("nntpchan_mod_target");
  var target = null;
  if (input) {
    target = input.value;
  }
  if (mod_action.parser) {
    target = mod_action.parser(target);
  }
  var elem; 
  if (result_elem) {
    elem = result_elem;
  } else {
    elem = document.getElementById("nntpchan_mod_result");
  }
  
  // clear old results
  while( elem.firstChild ) {
    elem.removeChild(elem.firstChild);
  }


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
    var url = mod_action.name + "/" + target;
    ajax.open(mod_action.method || "GET", url);
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
