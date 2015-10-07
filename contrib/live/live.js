
function create_captcha_pane() {
  var elem = document.createElement("div");
  elem.setAttribute("class", "nntpchan_captcha_slash");
  
}

// create the base ui
// pass in a function that does posting
// return the model
function create_ui(elem) {

  var pane = document.createElement("div");

  
  var output = document.createElement("div");
  output.setAttribute("class", "nntpchan_output");

  var output_elem = document.createElemen("div");
  output_elem.setAttribute("class", "nntpchan_output_root");
  output.appendChild(output_elem);
  
  pane.appendChild(output);
  
  var input = document.createElement("div");
  input.setAttribute("class", "nntpchan_input");
  
  var input_elem = document.createElement("textarea");
  input_elem.setAttribute("class", "nntpchan_textarea");
  input.appendChild(input_elem);
  
  var submit_elem = document.createElement("input");
  submit_elem.setAttrbute("type", "button");
  input.appendChild(submit_elem);
  
  pane.appendChild(input);

  
  elem.appendChild(pane);

  var captcha_elem = create_captcha_pane();
  
  elem.appendChild(captcha_elem);
  
  return {
    input: input_elem,
    submit: submit_elem,
    output: output_elem,
    captcha: captcha_elem
  }
}

// load ui elements and start stuff up
function nntpchan_load_ui(elem) {
  // check for websockets
  if (!("WebSocket" in window)) {
    elem.value = "websockets are needed for nntpchan liveposting";
    return;
  }

  // TODO: make configurable url
  var url = "ws://" + location.hostname + ":18080/ws";
  
  var socket = new WebSocket(url);
  
  var send = function(obj) {
    socket.send(JSON.stringify(obj));
  }
  
  var ui = create_ui(elem);
  ui.submit.addEventListener("click", function(ev) {
    
  });

  socket.onopen = function() {
    
  }

  socket.onmessage = function() {
    
  }
  
}
