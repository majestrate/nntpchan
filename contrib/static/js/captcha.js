//
// sorry I can't think of any better way to do captcha ;~;
//

window.addEventListener('load', function() {
  // get new captcha
  var ajax = new XMLHttpRequest();
  // get form elements for captcha
  var elem_input = document.getElementById("captcha_input");
  var elem_img = document.getElementById("captcha_img");
  // prepare ajax
  ajax.onreadystatechange = function(ev) {
    if ( ajax.readyState == XMLHttpRequest.DONE && ajax.status == 200 ) {
      // we succeeded
      var captcha_id = ajax.responseText;
      // set captcha id 
      elem_input.value = captcha_id;
      // set captcha image
      elem_img.src = "captcha/" + captcha_id + ".png";
    }
  };
  // open and send the ajax request
  ajax.open("GET", "captcha/new");
  ajax.send();
});
