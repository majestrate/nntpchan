



/** inject search widget */
function inject_search(elem) {
  var inner = document.createElement("div");
  var button = document.createElement("button");
  var input = document.createElement("input");
  var status = document.createElement("span");
  var output = document.createElement("div");

  function inject_search_result(r) {
    var e = document.createElement("div");
    e.innerHTML = r.PostMarkup;
    output.appendChild(e);
    output.appendChild(document.createElement("hr"));
  }

  botton.onclick = function(ev) {
    var text = input.value;
    input.value = "";
    while(output.children.length > 0) 
      output.children[0].remove();
    var ajax = new XMLHttpRequest();
    ajax.onreadystatechange = function() {
      if (ajax.readyState == XMLHttpRequest.DONE) {
        // done
        if(ajax.status == 200) {
          // good
          var result = JSON.parse(ajax.responseText);
          if (result.length == 0) {
            status.innerHTML = "no results";
          } else {
            status.innerHTML = "found "+result.length+"results";
            for (var idx = 0 ; idx < result.length; idx ++ ) {
              inject_search_result(result[idx]);
            }
          }
        } else {
          status.innerHTML = "HTTP "+ajax.status;
        }
      }
    }
  }

  inner.appendChild(input);
  inner.appendChild(button);
  inner.appendChild(status);
  
  elem.appendChild(inner);
  elem.appendChild(output);
}
