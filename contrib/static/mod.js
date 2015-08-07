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
    name: "ban",
    handle: function(j) {
      if (j.banned) {
        return document.createTextNode(j.banned);
      }
    }
  });
}


// handle delete command
function nntpchan_delete() {
  nntpchan_mod({
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


function nntpchan_mod(mod_action) {

  // get the element
  var input = document.getElementById("nntpchan_mod_target");
  // get the long hash
  var longhash = get_longhash(input.value);

  var elem = document.getElementById("nntpchan_mod_result");
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
            } else {
              // fail
              alert("mod action failed, handler returned nothing");
            }
          } else {
            // fail
            alert("mod action has no handler");
          }
        }
      } else if (status) {
        // nah
        // http error
        elem.innerHTML = "error: HTTP "+status;
      }
      // clear input
      input.value = "";
    }
  }
  if (mod_action.name) {
    var url = mod_action.name + "/" + longhash;
    ajax.open("GET", url);
    ajax.send();
  } else {
    alert("mod action has no name");
  }
}
