function createBoard() {
  var form = document.getElementById("postform");
  var e = document.getElementById("boardname");
  form.action = form.action + e.value;
  form.submit();
}
