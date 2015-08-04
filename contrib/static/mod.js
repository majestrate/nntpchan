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

// handle delete command
function nntpchan_delete() {
  // get the element
  var input = document.getElementById("nntpchan_mod_delete");
  // get the long hash
  var longhash = get_longhash(input.value);
  // TODO: check long hash


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
          if ( j.deleted ) {
            for ( var idx = 0 ; idx < j.deleted.length ; idx ++ ) {
              var deltxt = "deleted " + j.deleted[idx];
              var e = document.createTextNode(deltxt);
              elem.appendChild(e);
            }
          }
          if ( j.notdeleted ) {
            for ( var idx = 0 ; idx < j.notdeleted.length ; idx ++ ) {
              var deltxt = "failed to delete " + j.notdeleted[idx];
              var e = document.createTextNode(deltxt);
              elem.appendChild(e);
            }
          }
        }
      } else {
        // nah
        // http error
        elem.innerHTML = "error: HTTP "+status;
      }
      input.value = "";
    }
  }
  ajax.open("GET", "del/"+longhash);
  ajax.send();
}
