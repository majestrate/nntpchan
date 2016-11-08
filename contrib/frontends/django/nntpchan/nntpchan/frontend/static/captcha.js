/** captcha.js */

onready(function(){
  var e = document.getElementById("captcha_img");
  if(!e) return; // no captcha
  var original_url = e.src;
  e.onclick = function() {
    e.src = original_url + "?t="+new Date().getTime();
  }
});
