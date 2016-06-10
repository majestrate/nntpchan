
function createConnectionElement(j) {
  var e = document.createElement("div");
  e.setAttribute("class", "connection");
  var auth = document.createElement("div");
  auth.appendChild(document.createTextNode("Connection: "+j.name));
  // authentication state
  if (j.authed) {
    auth.setAttribute("class", "authed");
    auth.appendChild(document.createTextNode("(authenticated)"));
  } else {
    auth.appendChild(document.createTextNode("(not authenticated)"));
  }
  e.appendChild(auth);

  // connection mode
  var mode = document.createElement("div");
  mode.setAttribute("class", "mode");
  mode.appendChild(document.createTextNode("mode: "+j.mode));
  e.appendChild(mode);

  var pending = document.createElement("div");
  pending.setAttribute("class", "pending");
  // pending articles
  var articles = Object.keys(j.pending);
  pending.appendChild(document.createTextNode("pending articles: "+articles.length));
  for ( var idx = 0 ; idx < articles.length; idx ++ ) {
    var msgid = articles[idx];
    var state = j.pending[msgid];
    var elem = document.createElement("div");
    elem.appendChild(document.createTextNode(msgid + ": " + state));
    elem.setAttribute("class", "pending_item "+state);
    pending.appendChild(elem);
  }
  e.appendChild(pending);
  // e.appendChild(document.createTextNode(JSON.stringify(j)));
  return e;
}

function inject_nntp_feed_element(feed, elem) {
  elem.appendChild(document.createElement("hr"));
  var name = document.createElement("div");
  name.setAttribute("class", "feeds_name");
  name_elem = document.createTextNode("Name: "+feed.State.Config.Name);
  name.appendChild(name_elem);
  elem.appendChild(name);
  var conns = document.createElement("div");
  conns.setAttribute("class", "connections");
  for ( var idx = 0 ; idx < feed.Conns.length; idx ++ ) {
    conns.appendChild(createConnectionElement(feed.Conns[idx]));
  }
  elem.appendChild(conns);
}

function update_nntpchan_feed_ticker(elem, result_elem) {
  nntpchan_admin("feed.list", null, function(j) {
    if (j) {
      if (j.error) {
        console.log("nntpchan_feed_ticker: error, "+j.error);
      } else {
        // remove all children
        while(elem.children.length) {
          elem.children[0].remove();
        }        
        var result = j.result;
        for (var idx = 0; idx < result.length; idx++) {
          var item = result[idx];
          var entry = document.createElement("div");
          inject_nntp_feed_element(item, entry);
          elem.appendChild(entry);
        }
      }
    }
  }, result_elem);
}


function nntp_feed_add() {
  var param = {};
  
  var e = document.getElementById("add_feed_name");
  param.name = e.value;
  e = document.getElementById("add_feed_host");
  param.host = e.value;
  e = document.getElementById("add_feed_port");
  param.port = parseInt(e.value);
  
  e = document.getElementById("nntpchan_feed_result");
  nntpchan_admin("feed.add", param, null, e);
}

function nntp_feed_del() {
  var e = document.getElementById("del_feed_name");
  var name = e.value;
  e = document.getElementById("nntpchan_feed_result");
  nntpchan_admin("feed.del", {name: name}, null, e);
}

function nntp_feed_update() {
  var e = document.getElementById("nntpchan_feeds");
  if (e) {
    setInterval(function(){
      var e1 = document.getElementById("nntpchan_feed_result");
      update_nntpchan_feed_ticker(e, e1);
    }, 1000);
  }
}
