

function buildCaptcha(domElem, prefix) {
  var captcha_widget = document.createElement("div");
  
  captcha_widget.className = "livechan_captcha_inner";

  var outer = document.createElement("div");

  outer.className = "livechan_captcha";
  
  var text = document.createElement("div");
  text.textContent = "solve the captcha";
  captcha_widget.appendChild(text);
  
  var captcha_image = document.createElement("img");
  captcha_image.className = "livechan_captcha_image";
  var div = document.createElement("div");
  div.appendChild(captcha_image);
  captcha_widget.appendChild(div);
  
  var captcha_entry = document.createElement("input");
  captcha_entry.className = "livechan_captcha_input";
  var div = document.createElement("div");
  div.appendChild(captcha_entry);
  captcha_widget.appendChild(div);
  
  var captcha_submit = document.createElement("input");
  captcha_submit.setAttribute("type", "button");
  captcha_submit.value = "solve";
  var div = document.createElement("div");
  div.appendChild(captcha_submit);
  captcha_widget.appendChild(div);

  outer.appendChild(captcha_widget);
  domElem.appendChild(outer);
  
  return {
    widget: outer,
    button: captcha_submit,
    image: captcha_image,
    entry: captcha_entry,
    prefix: prefix,
  }
}

function Captcha(domElem, options, callback) {
  if (options) {
    this.options = options;
  } else {
    this.options = {};
  }
  
  this.prefix = options.prefix || "/";  
  this.widget = buildCaptcha(domElem, this.prefix);
  var self = this;
  this.widget.button.addEventListener("click", function() { self.process(callback); });
}

Captcha.prototype.load = function() {
  var self = this;
  var xhr = new XMLHttpRequest();

  var url = location.protocol + "//" + location.host + this.prefix ;

  xhr.open('get', url +"captcha/new");
  xhr.onreadystatechange = function () {
    if (xhr.readyState == 4 && xhr.status == 200) {
      var jdata = JSON.parse(xhr.responseText);
      if ( jdata ) {
        self.setCaptchaId(jdata);
      }
    }
  }
    
  xhr.send();
}

/**
 * @brief set captcha id
 */
Captcha.prototype.setCaptchaId = function(data) {
  this.captcha_id = data.id;
  this.setImageUrl(data.url);
}

Captcha.prototype.setImageUrl = function(url) {
  this.widget.image.setAttribute("src", url);
}

/**
 * @brief process captcha form
 */
Captcha.prototype.process = function(callback) {
  console.log("process");
  console.log(this);
  if (this.captcha_id) {
    var solution = this.widget.entry.value;
    var self = this;
    callback(this.captcha_id, solution , function(solved) {
      if (solved) {
        // woot we solved it 
        self.hide();
      } else {
        // TODO: inform user of bad solution
        self.load();
      }
    });
  } else {
    // TODO: inform user of no captcha entred
    self.load();
  }
}

/**
 * @brief show the captcha pane
 */
Captcha.prototype.show = function () {
  console.log("hide captcha");
  var widget = this.widget.widget;
  if ( widget.style ) {
    widget.style.zIndex = 5;
  } else {
    widget.style = {zIndex: 5};
  }
}
/**
 * @brief hide the captcha pane
 */
Captcha.prototype.hide = function () {
  console.log("hide captcha");
  var widget = this.widget.widget;
  if ( widget.style ) {
    widget.style.zIndex = -1;
  } else {
    widget.style = {zIndex: -1};
  }
}

/**
 * build login widget
 * for captcha / mod login 
 */
function buildLogin(domElem) {
  var widget = document.createElement("div");
  widget.className = "livechan_login_widget";
  widget.style.zIndex = -1;
  

  var mod_div = document.createElement("div");
  mod_div.className = "livechan_login";

  var mod_form = document.createElement("form");
  
  var mod_username = document.createElement("input");
  mod_form.appendChild(mod_username);
  mod_username.className = "livechan_login_username";
  
  var mod_password = document.createElement("input");
  mod_password.className = "livechan_login_password";
  mod_password.setAttribute("type", "password");
  mod_form.appendChild(mod_password);
  
  var mod_submit = document.createElement("input");
  mod_password.className = "livechan_login_submit";
  mod_submit.setAttribute("type", "submit");
  mod_submit.setAttribute("value", "login");
  mod_form.appendChild(mod_submit);
  mod_div.appendChild(mod_form);
  widget.appendChild(mod_div);
  domElem.appendChild(widget);
  return {
    widget: widget,
    mod: {
      form: mod_form,
      username: mod_username,
      password: mod_password,
      submit: mod_submit
    }
  }
}


function Login(domElem) {
  this._login = buildLogin(domElem);
}

Login.prototype.show = function() {
  var self = this;
  self._login.widget.style.zIndex = 5;
  console.log("show login widget");
}



/* 
 *  @brief build livechan navbar
 *  @param domElem the root element to put the navbar in
 */
function LivechanNavbar(domElem) {

  this.navbar = document.createElement("div");
  this.navbar.className = 'livechan_navbar';

  var container = document.createElement("div");
  // channel name label
  var channelLabel = document.createElement("span");
  channelLabel.className = 'livechan_navbar_channel_label';

  this.channel = channelLabel;
  
  // mod indicator
  this.mod = document.createElement("span");
  this.mod.className = 'livechan_navbar_mod_indicator_inactive';

  // TODO: don't hardcode
  this.mod.textContent = "Anon";

  // usercounter
  this.status = document.createElement("span");
  this.status.className = 'livechan_navbar_status';

  container.appendChild(this.mod);
  container.appendChild(this.channel);
  container.appendChild(this.status);

  this.navbar.appendChild(container);
  
  domElem.appendChild(this.navbar);

}


/* @brief called when there is an "event" for the navbar */
LivechanNavbar.prototype.onLivechanEvent = function (evstr) {
  if ( evstr === "login:mod" ) {
    // set indicator
    this.mod.className = "livechan_mod_indicator_active";
    this.mod.textContent = "Moderator";
  } else if ( evstr === "login:admin" ) {
    this.mod.className = "livechan_mod_indicator_admin";
    this.mod.textContent = "Admin";
  }
}

/* @brief called when there is a notification for us */
LivechanNavbar.prototype.onLivechanNotify = function(evstr) {
  // do nothing for now
  // maybe have some indicator that shows number of messages unread?
}

/* @brief update online user counter */
LivechanNavbar.prototype.updateUsers = function(count) {
  this.updateStatus("Online: "+count);
}

/* @brief update status label */
LivechanNavbar.prototype.updateStatus = function(str) {
  this.status.textContent = str;
}

/* @brief set channel name */
LivechanNavbar.prototype.setChannel = function(str) {
  this.channel.textContent = str;
}

var modCommands = [
  // login command
  [/l(login)? (.*)/, function(m) {
    var chat = this;
    // mod login
      chat.modLogin(m[2]);
  },
   "login as user", "/l user:password",
  ],
  [/cp (\d+)/, function(m) {
    var chat = this;
    // permaban the fucker
    chat.modAction(3, 4, m[1], "CP", -1);
  },
   "handle illegal content", "/cp postnum",
  ],
  [/cnuke (\d+) (.*)/, function(m) {
    var chat = this;
    // channel ban + nuke files
    chat.modAction(2, 4, m[1], m[2], -1);
  },
   "channel level ban+nuke", "/cnuke postnum reason goes here",
  ],
  [/purge (\d+) (.*)/, function(m) {
    var chat = this;
    // channel ban + nuke files
    chat.modAction(2, 9, m[1], m[2], -1);
  },
   "channel level ban+nuke", "/cnuke postnum reason goes here",
  ],
  [/gnuke (\d+) (.*)/, function(m) {
    var chat = this;
    // global ban + nuke with reason
    chat.modAction(3, 4, m[1], m[2], -1);
  },
   "global ban+nuke", "/gnuke postnum reason goes here",
  ],
  [/gban (\d+) (.*)/, function(m) {
    var chat = this;
    // global ban with reason
    chat.modAction(3, 3, m[1], m[2], -1);
  },
   "global ban (no nuke)", "/gban postnum reason goes here",
  ],
  [/cban (\d+) (.*)/, function(m) {
    var chat = this;
    // channel ban with reason
    chat.modAction(2, 3, m[1], m[2], -1);
  },
   "channel level ban (no nuke)", "/cban postnum reason goes here",
  ],
  [/dpost (\d+)/, function(m) {
    var chat = this;
    // channel level delete post
    chat.modAction(1, 2, m[1]);
  },
   "delete post and file", "/dpost postnum",
  ],
  [/dfile (\d+)/, function(m) {
    var chat = this;
    // channel level delete file
    chat.modAction(1, 1, m[1]);
  },
   "delete just file", "/dpost postnum",
  ]
]


/*
 * @brief Build Notification widget
 * @param domElem root element to put widget in
 */
function buildNotifyPane(domElem) {
  var pane = document.createElement("div");
  pane.className = "livechan_notify_pane";
  domElem.appendChild(pane);
  return pane;
}

/*
 * @brief Livechan Notification system
 * @param domElem root element to put Notification Pane in.
 */
function LivechanNotify(domElem) {
  this.pane = buildNotifyPane(domElem);
}

/* @brief inform the user with a message */
LivechanNotify.prototype.inform = function(str) {
  // new Notify("livechan", {body: str}).show();
  var elem = document.createElement("div");
  elem.className = "livechan_notify_node";
  elem.textContent = Date.now() + ": " + str;
  this.pane.appendChild(elem);
  this.rollover();
}


/* @brief roll over old messages */
LivechanNotify.prototype.rollover = function() {
  while ( this.pane.childNodes.length > this.scrollback ) {
     this.pane.childNodes.removeChild(this.pane.childNodes[0]);
   }
}




/* @brief Creates a structure of html elements for the
 *        chat.
 *
 * @param domElem The element to be populated with the
 *        chat structure.
 * @param chatName The name of this chat
 * @return An object of references to the structure
 *         created.
 */
function buildChat(chat, domElem, channel) {
  channel = channel.toLowerCase();
  // build the navbar
  // see nav.js
  var navbar = new LivechanNavbar(domElem);

  // build the notification system
  // see notify.js
  var notify = new LivechanNotify(domElem);

  var output = document.createElement('div');
  output.className = 'livechan_chat_output';

  var input_left = document.createElement('div');
  input_left.className = 'livechan_chat_input_left';
  
  var input = document.createElement('form');
  input.className = 'livechan_chat_input';
  
  var name = document.createElement('input');
  name.className = 'livechan_chat_input_name';
  name.setAttribute('placeholder', 'Anonymous');
  
  var file = document.createElement('input');
  file.className = 'livechan_chat_input_file';
  file.setAttribute('type', 'file');
  file.setAttribute('value', 'upload');
  file.setAttribute('id', channel+'_input_file');
 
    
  var messageDiv = document.createElement('div');
  messageDiv.className = 'livechan_chat_input_message_div';
  
  var message = document.createElement('textarea');
  message.className = 'livechan_chat_input_message';

  var submit = document.createElement('input');
  submit.className = 'livechan_chat_input_submit';
  submit.setAttribute('type', 'submit');
  submit.setAttribute('value', 'send');
  var convobar = new ConvoBar(chat, domElem);
  input_left.appendChild(name); 
  input_left.appendChild(convobar.subject);
  input_left.appendChild(file);
  input.appendChild(input_left);
  messageDiv.appendChild(message);
  input.appendChild(messageDiv);
  input.appendChild(submit);
  domElem.appendChild(output);
  domElem.appendChild(input);
  
  return {
    convobar : convobar,
    notify: notify,
    navbar: navbar,
    output: output,
    input: {
      subject: convobar.subject,
      form: input,
      message: message,
      name: name,
      submit: submit,
      file: file
    }
  };
}

function Connection(chat, ws, url) {
  this.ws = ws;
  this.chat = chat;
  this.url = url;
  this.ws.onmessage = function(ev) {
    chat.handleData(JSON.parse(ev.data));
  }
  this.ws.onclose = function() {
    self.ws = null;
  }
  var self = this;
  setInterval(function() {
    if (self.ws == null) {
      // try reconnecting
      initWebSocket(self.chat, self.url, self);
    }
  }, 5000);
}

Connection.prototype.ban = function(reason) {
  if (this.ws) {
    this.ws.close();
    this.ws.close = null;
    alert("You have been banned for the following reason: "+reason);
  }
}

Connection.prototype.send = function(obj) {
  /* Jsonify the object and send as string. */
  this.sendBinary(JSON.stringify(obj));
}

Connection.prototype.sendBinary = function(obj) {
  if (this.ws) {
    this.ws.send(obj);
  }
}

Connection.prototype.onclose = function(callback) {
  if(this.ws)
    this.ws.onclose = callback;
}

/* @brief Initializes the websocket connection.
 *
 * @param channel The channel to open a connection to.
 * @return A connection the the websocket.
 */
function initWebSocket(chat, url, connection) {
  var ws = null;
  if (window['WebSocket']) {
    try {
      ws = new WebSocket(url);
    } catch (ex) {
      ws = null;
    }
  }
  if (ws) {
    ws.onclose = function() {
      if (connection) {
        connection.ws = null;
      }
    };
    if (connection) {
      connection.ws = ws;
      connection.ws.onmessage = function(ev) {
        chat.handleData(JSON.parse(ev.data));
      }
      chat.clear();
      chat.handleMessage({Type: "post", PostMessage: "reconnecting..."});
      return connection;
    } else {
      return new Connection(chat, ws, url);
    }
  }
}

/* @brief Parses and returns a message div.
 *
 * @param data The message data to be parsed.
 * @return A dom element containing the message.
 */
function parse(text, rules, end_tag) {
  var output = document.createElement('div'); 
  var position = 0;
  var end_matched = false;
  if (end_tag) {
    var end_handler = function(m) {
      end_matched = true;
    }
    rules = [[end_tag, end_handler]].concat(rules);
  }
  do {
    var match = null;
    var match_pos = text.length;
    var handler = null;
    for (var i = 0; i < rules.length; i++) {
      rules[i][0].lastIndex = position;
      var result = rules[i][0].exec(text);
      if (result !== null && position <= result.index && result.index < match_pos) {
        match = result;
        match_pos = result.index;
        handler = rules[i][1];
      }
    }
    var unmatched_text = text.substring(position, match_pos);
    output.appendChild(document.createTextNode(unmatched_text));
    position = match_pos;
    if (match !== null) {
      position += match[0].length;
      output.appendChild(handler(match));
    }
  } while (match !== null && !end_matched);
  return output;
}

var messageRules = [
  [/>>([0-9a-f]+)/g, function(m) {
    var out = document.createElement('span');
    out.className = 'livechan_internallink';
    out.addEventListener('click', function() {
      var selected = document.getElementById('livechan_chat_'+m[1]);
      console.log(selected.convo);
      selected.select();
      selected.scrollIntoView(true);
      // TODO: highlight
    });
    out.appendChild(document.createTextNode('>>'+m[1]));
    return out;
  }],
  [/^>.+/mg, function(m) {
    var out = document.createElement('span');
    out.className = 'livechan_greentext';
    out.appendChild(document.createTextNode(m));
    return out;
  }],
  [/\[code\]\n?([\s\S]+)\[\/code\]/g, function(m) {
    var out;
    if (m.length >= 2 && m[1].trim !== '') {
      out = document.createElement('pre');
      out.textContent = m[1];
    } else {
      out = document.createTextNode(m);
    }
    return out;
  }],
  [/\[b\]\n?([\s\S]+)\[\/b\]/g, function(m) {
    var out;
    if (m.length >= 2 && m[1].trim !== '') {
      out = document.createElement('span');
      out.className = 'livechan_boldtext';
      out.textContent = m[1];
    } else {
      out = document.createTextNode(m);
    }
    return out;
  }],
  [/\[spoiler\]\n?([\s\S]+)\[\/spoiler\]/g, function(m) {
    var out;
    if ( m.length >= 2 && m[1].trim !== '') {
      out = document.createElement('span');
      out.className = 'livechan_spoiler';
      out.textContent = m[1];
    } else {
      out = document.createTextNode(m);
    }
    return out;
  }],
  [/\r?\n/g, function(m) {
    return document.createElement('br');
  }],
  [/==(.*)==/g, function(m) {
    var out;
    out = document.createElement("span");
    out.className = "livechan_redtext";
    out.textContent = m[1];
    return out;
  }],
  [/((([A-Za-z]{3,9}:(?:\/\/)?)(?:[-;:&=\+\$,\w]+@)?[A-Za-z0-9.-]+|(?:www.|[-;:&=\+\$,\w]+@)[A-Za-z0-9.-]+)((?:\/[\+~%\/.\w-_]*)?\??(?:[-\+=&;%@.\w_]*)#?(?:[\w]*))?)/g, function(m) {
    var out = document.createElement("a");
    out.href = m[1];
    out.textContent = m[1];
    return out;
  }],
]

/* @brief build the convorsation bar's elements
*/
function buildConvoBar(domElem) {
  var elem = document.createElement("div");
  elem.className = "livechan_convobar_root";
  
  var subject = document.createElement("input");
  subject.className = "livechan_chat_input_convo";
  elem.appendChild(subject);
  domElem.appendChild(elem);
  
  return {
    subject: subject,
    widget: elem,
  }
}

/* @brief create the chat's convorsation bar
 * @param domElem the element to place everything in
 */
function ConvoBar(chat, domElem) {
  this.parent = chat;
  this.holder = {};
  this.domElem = domElem;
  var convo = buildConvoBar(domElem);
  this.widget = convo.widget;
  this.subject = convo.subject;
  this.active = null;
  this._num = 0;
}

ConvoBar.prototype.count = function() {
  return this._num;
}

/* @brief remove convorstation given root message-id */
ConvoBar.prototype.removeConvo = function(msgid) {
  var self = this;
  console.log("remove "+msgid);
  var c = self.holder[msgid];
  if (c) {
    var e = document.getElementById("livechan_convobar_item_"+c.id);
    if (e) e.remove();
    for(var idx = 0; idx < c.posts.length; idx ++ ) {
      var id = "livechan_chat_"+c.posts[idx].HashShort;
      var child = document.getElementById(id);
      if(child) child.remove();
    }
    
    delete self.holder[msgid];
    self._num -- ;
  }
}

/* @brief update the convo bar
 * @param convoId the name of this covnorsattion
 */
ConvoBar.prototype.update = function(msgid, data) {
  var self = this;
  if ( ! self.holder[msgid]) {
    // new convo
    // register convo
    self.registerConvo(msgid, data);
  } 
  // bump existing convo
  var convoId = self.holder[msgid].id;
  var convoElem = document.getElementById("livechan_convobar_item_"+convoId);
  var convoParent = convoElem.parentElement;
  if ( convoParent.children.length > 1 ) {
    convoParent.removeChild(convoElem);
    convoParent.insertBefore(convoElem, convoParent.childNodes[0]);
  }
  // add post to convo
  self.holder[msgid].posts.push(data);
  // do roll over
  var scrollback = self.parent.options.scrollback || 30;
  while(self.holder[msgid].posts.length > scrollback) {
    // remove oldest from convo tracker
    var child_data = self.holder[msgid].posts.shift();
    var child = document.getElementById("livechan_chat_"+child_data.ShortHash);
    if(child) {
      child.remove();
    }
  }
  // lol quartic time
  while(self.count() > 10 ) {
    var minid = -1;
    var minmsgid = null;
    for( var i in self.holder ) {
      if (minid == -1 || self.holder[i].id < minid) {
        minid = self.holder[i].id;
        minmsgid = i;
      }
    }
    if(minmsgid)
      self.removeConvo(minmsgid);
  }
}



/** @brief register a new convorsation
 */
ConvoBar.prototype.registerConvo = function(msgid, data) {
  var self = this;
  var max_id = 0;
  // get the highest convo id
  for ( c in self.holder ) {
    var id = self.holder[c].id;
    if (id > max_id ) {
      max_id = id
    }
  }

  self.holder[msgid] = {
    subject: data.PostSubject,
    msgid: data.Message_id,
    id: max_id + 1,
    posts: [],
    group: data.Newsgroup,
    ShortHash: data.ShortHash,
    select: function() {
      console.log("selected convo "+msgid);
      if ( self.active !== msgid ) {
        self.show(msgid);
      }
    },
  }
  // make a new entry in the convo bar
  var elem = document.createElement("div");
  elem.className = "livechan_convobar_item";
  elem.setAttribute("id", "livechan_convobar_item_"+ self.holder[msgid].id);
  var link = document.createElement("span");
  elem.addEventListener("click", function() { self.show(msgid); });
  link.appendChild(document.createTextNode(data.PostSubject));
  elem.appendChild(link);
  // prepend the element
  if (self.widget.children.length > 0 ) {
    self.widget.insertBefore(elem, self.widget.childNodes[0]);
  } else {
    self.widget.appendChild(elem);
  }
  self._num ++;
}

/* @brief Only Show chats from a convorsation
 */
ConvoBar.prototype.show = function(msgid) {
  var self = this;
  var sheet = null;
  for(var idx = 0; idx < document.styleSheets.length; idx++ ) {
    var s = document.styleSheets[idx];
    if (s.ownerNode && s.ownerNode.id === "convo_filter") {
      sheet = s;
      break;
    }
  }
  var rules = null;
  // delete all filtering rules
  if (sheet.rules) {
    rules = sheet.rules;
  } else {
    rules = sheet.cssRules;
  }
  while (rules.length > 0 ) {
    if (sheet.deleteRule) {
      sheet.deleteRule(0);
    } else if (sheet.removeRule) {
      sheet.removeRule(0);
    } else {
      break;
    }
  }

  if (msgid === self.active) {
    // this is resetting the view
    if (sheet.insertRule) {  // firefox
      sheet.insertRule(".livechan_chat_output_chat {  display: block; }", 0);
    } else if (sheet.addRule) { // not firefox
      sheet.addRule(".livechan_chat_output_chat", "display: block");
    }
    // unset active highlight
    var convoId = self.holder[self.active].id;
    var itemElem = document.getElementById("livechan_convobar_item_"+convoId);
    itemElem.style.background = null;
    self.active = null;
  } else {
    // unset active highlight if it's there
    if (self.active) {
      var convoId = self.holder[self.active].id;
      var itemElem = document.getElementById("livechan_convobar_item_"+convoId);
      itemElem.style.background = null;
    }
    // set active highlight to new element
    convoId = self.holder[msgid].id;
    itemElem = document.getElementById("livechan_convobar_item_"+convoId);
    itemElem.style.background = "red";
    var elemClass = ".livechan_chat_convo_" + convoId;
    if (sheet.insertRule) {  // firefox
      sheet.insertRule(elemClass+ " { display: block; }", 0);
      sheet.insertRule(".livechan_chat_output_chat {  display: none; }", 0);
    } else if (sheet.addRule) { // not firefox
      sheet.addRule(".livechan_chat_output_chat", "display: none");
      sheet.addRule(elemClass, "display: block");
    }
    // this convo is now active
    self.active = msgid;
  }
  // scroll view
  self.parent.scroll();
}

/* @brief Creates a chat.
 *
 * @param domElem The element to populate with chat
 *        output div and input form.
 * @param channel The channel to bind the chat to.
 *
 * @param options Channel Specific options
 */
function Chat(domElem, channel, options) {
  var self = this;
  this.name = channel.toLowerCase();
  this.domElem = domElem;
  this.lastOp = null;
  if (options) {
    this.options = options;
  } else {
    this.options = {};
  }

  
  this.chatElems = buildChat(this, this.domElem, this.name);
  var prefix = this.options.prefix || "/";
  var scheme = "wss://";
  if (location.protocol == "http:") scheme = "ws://";
  var url = scheme + location.host + prefix + "live?"+ this.name;
  this.connection = initWebSocket(this, url);
  this.initOutput();
  this.initInput();
  // set navbar channel name
  this.chatElems.navbar.setChannel(this.name);
  // prepare captcha callback
  this.captcha_callback = null;
  // create captcha
  this.captcha = new Captcha(this.domElem, this.options, function(id, solution, callback) {
    // set callback to handle captcha success
    self.captcha_callback = callback;
    // send captcha solution
    self.connection.send({Captcha: { ID: id, Solution: solution}});
  });
  this.captcha.hide();
}

Chat.prototype.clear = function () {
  for ( var convo in this.chatElems.convobar.holder ) {
    this.chatElems.convobar.removeConvo(convo);
  }
}

/**
 * @brief begin login sequence
 */
Chat.prototype.login = function() {
  this.captcha.show();
  this.captcha.load();
}

/**
 * @brief do mod login
 */
Chat.prototype.modLogin = function(str) {
  var self = this;
  self.connection.send({ModLogin: str});
}

Chat.prototype.modAction = function(scope, action, postID, reason, expire) {
  var self = this;
  self.connection.send({
    ModReason: reason,
    ModScope: parseInt(scope),
    ModAction: parseInt(action),
    ModPostID: parseInt(postID),
    ModExpire: parseInt(expire),
  });
}

/* @brief called when our post got mentioned
 *
 * @param event the event that has this mention
 */
Chat.prototype.Mentioned = function(event, chat) {
  var self = this;
  self.notify("mentioned: "+chat);
}

Chat.prototype.onNotifyShow = function () {

}


Chat.prototype.readImage = function (elem, callback) {
  var self = this;

  if (elem.files.length > 0 ) {
    var reader = new FileReader();
    var file = elem.files[0];
    reader.onloadend = function(ev) {
      if (ev.target.readyState == FileReader.DONE) {
        callback(window.btoa(ev.target.result), file.name, file.type);
      }
    }
    reader.readAsBinaryString(file);
  } else {
    callback(null, null, null);
  }
}

/* @brief Sends the message in the form.
 *
 * @param event The event causing a message to be sent.
 */
Chat.prototype.sendInput = function(event) {
  var inputElem = this.chatElems.input;
  var connection = this.connection;
  var self = this;
    
  if (inputElem.message.value[0] == '/') {
    var inp = inputElem.message.value;
    var helpRegex = /(help)? (.*)/;
    var helpMatch = helpRegex.exec(inp.slice(1));
    if (helpMatch) {
      
    }
    if ( self.options.customCommands ) {
      for (var i in self.options.customCommands) {
        var regexPair = self.options.customCommands[i];
        var match = regexPair[0].exec(inp.slice(1));
        if (match) {
          (regexPair[1]).call(self, match);
          inputElem.message.value = '';
        }
      }
    }
    // modCommands is defined in mod.js
    for ( var i in modCommands ) {
      var command = modCommands[i];
      var match = command[0].exec(inp.slice(1));
      if (match) {
        (command[1]).call(self, match);
        // don't clear input for mod command
      }
    }
    event.preventDefault();
    return false;
  }
  if (inputElem.submit.disabled == false) {
    var message = inputElem.message.value;
    var name = inputElem.name.value;
    var convo = self.chatElems.convobar.active;
    var board;
    if(convo)
      board = self.chatElems.convobar.holder[convo].group;
    if (!board) board = "overchan.random";
    console.log(board);
    var subject = self.chatElems.input.subject.value;
    self.readImage(inputElem.file, function(fdata, fname, ftype) {
      if (fdata) {
        connection.send({Type: "post", Post: {
          message: message,
          subject: subject,
          name: name,
          reference: convo,
          newsgroup: board,
          files: [{name: fname, data: fdata, type: ftype}],
        }});
      } else {
        connection.send({Type: "post", Post: {
          subject: subject,
          message: message,
          reference: convo,
          name: name,
          newsgroup: board,
        }});
      }
      inputElem.file.value = "";
      inputElem.message.value = '';
    });
    inputElem.submit.disabled = true;
    var i = parseInt(self.options.cooldown);
    // fallback
    if ( i == NaN ) { i = 4; } 
    inputElem.submit.setAttribute('value', i);
    var countDown = setInterval(function(){
      inputElem.submit.setAttribute('value', --i);
    }, 1000);
    setTimeout(function(){
      clearInterval(countDown);
      inputElem.submit.disabled = false;
      inputElem.submit.setAttribute('value', 'send');
    }, i * 1000);
    event.preventDefault();
    return false;
  }
}

/* @brief Binds the form submission to websockets.
 */
Chat.prototype.initInput = function() {
  var inputElem = this.chatElems.input;
  var connection = this.connection;
  var self = this;
  inputElem.form.addEventListener('submit', function(event) {
    self.sendInput(event);
  });
  
  inputElem.message.addEventListener('keydown', function(event) {
    /* If enter key. */
    if (event.keyCode === 13 && !event.shiftKey) {
      self.sendInput(event);
    }
  });
  inputElem.message.focus();
}


/* @brief show a notification to the user */
Chat.prototype.notify = function(message) {
  // show notification pane
  this.showNotifyPane();

  var notifyPane = this.chatElems.notify;

  notifyPane.inform(message);
}

/* @brief show the notification pane */
Chat.prototype.showNotifyPane = function () {
  var pane = this.chatElems.notify.pane;
  pane.style.zIndex = 5;
  pane.style.visibility = 'visible';
}

/* @brief hide the notification pane */
Chat.prototype.showNotifyPane = function () {
  var pane = this.chatElems.notify.pane;
  pane.style.zIndex = -1;
  pane.style.visibility = 'hidden';
}

Chat.prototype.error = function(message) {
  var self = this;
  console.log("error: "+message);
  self.notify("an error has occured: "+message);
}

/** handle inbound websocket message */
Chat.prototype.handleMessage = function (data) {
  var self = this;

  var mtype = data.Type.toLowerCase();

  if (mtype == "captcha" ) {
    if (self.captcha_callback) {
      // captcha reply
      self.captcha_callback(data.Success);
      // reset
      self.captcha_callback = null;
    } else {
      // captcha challenge
      self.login();
    }
  } else if (mtype == "posted" ) {
    // user posted something
    if ( data.OP ) {
      // they made a new thread, focus on it once it comes up
      self.lastOp = data.Msgid
      console.log("made new thread: "+data.Msgid)
    }
  } else if (mtype == "post" ) {
    self.insertChat(self.generateChat(data), data);
  } else if (mtype == "count" ) {
    self.chatElems.navbar.updateUsers(data.UserCount);
  } else if (mtype == "ban" ) {
    self.connection.ban(data.Reason);
  } else if (mtype == "error") {
    self.insertChat(self.generateChat({PostMessage: data.Error, PostSubject: "Server Error", PostName: "Server"}))
    console.log("server error: "+data.Error);
  } else {
    console.log("unknown message type "+mtype);
  }
}


Chat.prototype.handleData = function(data) {
  var self = this;
  if( Object.prototype.toString.call(data) === '[object Array]' ) {
    for (var i = 0; i < data.length; i++) {
      self.handleMessage(data[i]);
    }
  } else {
    self.handleMessage(data);
  }
}

/* @brief Binds messages to be displayed to the output.
 */
Chat.prototype.initOutput = function() {
  var outputElem = this.chatElems.output;
  var connection = this.connection;
  var self = this;
}

/* @brief update the user counter for number of users online
 */
Chat.prototype.updateUserCount = function(count) {
  var elem = this.chatElems.navbar.userCount;
  elem.textContent = "Online: "+count;
}

/* @brief Scrolls the chat to the bottom.
 */
Chat.prototype.scroll = function() {
  this.chatElems.output.scrollTop = this.chatElems.output.scrollHeight;
}

/** @brief roll over old posts, remove them from ui */
Chat.prototype.rollover = function() {
  /*
  var self = this;
  var chatSize = self.options.scrollback || 50;
  var e = self.chatElems.output;
  while ( e.childNodes.length > chatSize ) {
    e.childNodes[0].remove();
  }
*/
}

/* @brief Inserts the chat into the DOM, overwriting if need be.
 *
 * @TODO: Actually scan and insert appropriately for varying numbers.
 *
 * @param outputElem The dom element to insert the chat into.
 * @param chat The dom element to be inserted.
 * @param number The number of the chat to keep it in order.
 */
Chat.prototype.insertChat = function(chat, data) {
  //var number = data.Count;
  //var convo = data.Convo;
  //if (!number) {
  //  this.error("Error: invalid chat number.");
  //}
  var self = this;
  // append to main output
  var outputElem = this.chatElems.output;
  outputElem.appendChild(chat);
  // scroll to end
  self.scroll();
  self.rollover();
  // show new thread
  if (self.lastOp) {
    self.chatElems.convobar.show(self.lastOp);
    self.lastOp = null;
  }
}


/* @brief Generates a chat div.
 *
 * @param data Data passed in via websocket.
 * @return A dom element.
 */
Chat.prototype.generateChat = function(data) {
  var self = this;

  var chat = document.createElement('div');
  self.chatElems.convobar.update(data.Parent, data);
  var convo = self.chatElems.convobar.holder[data.Parent];
  chat.select = function() {
    console.log("selecting...");
    convo.select();
  }
  chat.className = 'livechan_chat_output_chat livechan_chat_convo_' + convo.id;
  var convoLabel = document.createElement('span');
  convoLabel.className = 'livechan_convo_label';
  convoLabel.appendChild(document.createTextNode(data.PostSubject));
  
  var header = document.createElement('div');
  header.className = 'livechan_chat_output_header';
  var name = document.createElement('span');
  name.className = 'livechan_chat_output_name';
  var trip = document.createElement('span');
  trip.className = 'livechan_chat_output_trip';
  var date = document.createElement('span');
  date.className = 'livechan_chat_output_date';
  var count = document.createElement('span');
  count.className = 'livechan_chat_output_count';

  var body = document.createElement('div');
  body.className = 'livechan_chat_output_body';
  var message = document.createElement('div');
  message.className = 'livechan_chat_output_message';
  

  if (data.PostName) {
    name.appendChild(document.createTextNode(data.PostName));
  } else {
    name.appendChild(document.createTextNode('Anonymous'));
  }

  if (data.Files) {
    for (var idx = 0 ; idx < data.Files.length ; idx ++ ) {
      var file = data.Files[idx];
      if(!file) continue;
      var a = document.createElement('a');
      a.setAttribute('target', '_blank');
      // TODO: make these configurable
      var filepath = file.Path;
      var thumb_url = self.options.prefix + 'thm/'+filepath + ".jpg";
      var src_url = self.options.prefix + 'img/'+filepath;
      
      a.setAttribute('href',src_url);
      var fl = filepath.toLowerCase();
      var img = document.createElement('img');
      img.setAttribute('src', thumb_url);
      img.className = 'livechan_image_thumb';
      a.appendChild(img);
      message.appendChild(a);
      img.onload = function() { self.scroll(); }
      
      img.addEventListener('mouseover', function () {
        
        var e = document.createElement("div");
        e.setAttribute("id", "hover_"+data.ShortHash);
        e.setAttribute("class", "hover");
        
        if (fl.match(/\.(webm|mp4|mkv)$/)) {
          // video
          var v = document.createElement("video");
          v.src = src_url;
          e.appendChild(v);
        } else if (fl.match(/\.(mp3|ogg|oga|flac|opus)$/)) {
          // audio
          var a = document.createElement("audio");
          a.src = src_url;
          e.appendChild(a);
        } else if (fl.match(/\.txt$/)) {
          // 
        } else {
          // image
          var i = document.createElement("img");
          i.src = src_url;
          e.appendChild(i);
        }
        chat.appendChild(e);
      });
      img.addEventListener('mouseout', function () {
        // unload image
        var e = document.getElementById("hover_"+data.ShortHash);
        e.parentElement.removeChild(e);
      });
    }
  }
    
  /* Note that parse does everything here.  If you want to change
   * how things are rendered modify messageRules. */
  if (data.PostMessage) {
    message.appendChild(parse(data.PostMessage, messageRules));
  } else {
    message.appendChild(document.createTextNode(''));
  }

  if (data.Posted) {
    date.appendChild(document.createTextNode((new Date(data.Posted * 1000)).toLocaleString()));
  }

  if (data.Tripcode) {
    var et = document.createElement('span');
    et.innerHTML = data.Tripcode;
    trip.appendChild(et);
  }

  if (data.HashShort) {
    var h = data.HashShort;
    chat.setAttribute('id', 'livechan_chat_'+h);
    count.appendChild(document.createTextNode(h));
    count.addEventListener('click', function() {
      self.chatElems.input.message.value += '>>'+h+'\n';
      self.chatElems.input.message.focus();
      chat.select();
    });
  }

  header.appendChild(name);
  header.appendChild(trip);
  header.appendChild(date);
  header.appendChild(convoLabel);
  header.appendChild(count);
  body.appendChild(message);

  chat.appendChild(header);
  chat.appendChild(body);
  return chat;
}
