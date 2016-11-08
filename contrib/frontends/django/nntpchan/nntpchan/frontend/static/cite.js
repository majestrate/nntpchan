/** cite.js */


function thread_citepost(hash) {
  var e = document.getElementById("postform_message");
  if (!e) return; // no postform element
  e.value += ">>"+hash.slice(0, 18) + "\n";
}

function board_postcite(hash) {
  // TODO implement
}

onready(function() {
  // inject postciter
  forEachInClass("post-cite", function(elem) {
    
  });
});
