function createBoard() {
  var form = document.getElementById("postform");
  var e = document.getElementById("boardname");
  var board = e.value;
  if ( board.indexOf(".") == -1 ) {
    board = "overchan." + board;
  }
  form.action = form.action + board;
  form.submit();
}
