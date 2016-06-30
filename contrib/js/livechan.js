function livechan_got_post(widget, j) {
  // do scroll
  while (widget.children.length > 5) {
    // remove top element
    widget.removeChild(widget.children[0]);
  }
  nntpchan_buildpost(widget, j);
  // scroll to bottom
  widget.scrollTop = widget.scrollHeight;
}

// inject post form into an element
function inject_postform(prefix, parent) {
  
}

// inject livechan widget into parent
function inject_livechan_widget(prefix, parent) {
  if ( "WebSocket" in window ) {
    var url = "ws://"+document.location.host+prefix+"live";
    if ( document.location.protocol == "https:" ) {
      url = "wss://"+document.location.host+prefix+"live";
    }
    var socket = new WebSocket(url);
    var progress = function(str) {
      parent.innerHTML = "<pre>livechan: "+str+"</pre>";
    };
    progress("initialize");
    socket.onopen = function () {
      progress("streaming");
    }
    socket.onmessage = function(ev) {
      var j = null;
      try {
        j = JSON.parse(ev.data);
      } catch(e) {
        // ignore
      }
      if (j) {
        livechan_got_post(parent, j);
      }
    }
    socket.onclose = function(ev) {
      progress("connection closed");
      setTimeout(function() {
        inject_livechan_widget(prefix, parent);
      }, 1000);
    }
  } else {
    parent.innerHTML = "<pre>livechan mode requires websocket support</pre>";
    setTimeout(function() {
      parent.innerHTML = "";
    }, 5000);
  }
}

function ukko_livechan(prefix) {
  var ukko = document.getElementById("ukko_threads");
  if (ukko) {
    // remove children
    ukko.innerHTML = "";
    inject_livechan_widget(prefix, ukko);
  }
}

