//
// nntpchan.js -- frontend ui niceness
//


// insert a backlink for a post given its short hash
function nntpchan_backlink(shorthash)
{
  var elem = document.getElementById("postform_message");
  if ( elem )
  {
    elem.value += ">>" + shorthash + "\n";
  }
}
