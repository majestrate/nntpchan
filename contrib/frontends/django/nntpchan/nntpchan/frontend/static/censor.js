/** censor tools */


function show_censortools() {
  var e = document.getElementById("censor-tools");
  e.checked = true;
}

function nntpchan_report_thread(posthash) {
  var thread = document.getElementById(posthash);
  if (!thread) return;
  var posts = thread.getElementsByClassName("post");
  for (var idx = 0; idx < posts.length; idx ++ ) {
    var post = posts[idx];
    nntpchan_report(post.dataset.msgid);
  }
}

function nntpchan_report(msgid) {
  var e = document.getElementById("modactions");
  if (!e) return;
  e.value += "delete "+msgid+"\n";
  show_censortools();
}

onready(function() {
  
});
