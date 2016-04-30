//For source code and license information please check https://github.com/majestrate/nntpchan 


/* ./contrib/js/main.js_ */
onready_callbacks=[];function onready(fnc){onready_callbacks.push(fnc);}
function ready(){for(var i=0;i<onready_callbacks.length;i++){onready_callbacks[i]();}}
/* ./contrib/js/api.js */
function nntpchan_apicall(url,handler,err_handler){var ajax=new XMLHttpRequest();ajax.onreadystatechange=function(){if(ajax.readyState==XMLHttpRequest.DONE){var status=ajax.status;var j=null;if(status==200){try{j=JSON.parse(ajax.responseText);}catch(e){}}else if(status==410){if(err_handler){err_handler("cannot fetch post: api disabled");}
return;}
handler(j);}};ajax.open("GET",url);ajax.send();}
function nntpchan_buildpost(parent,j){var post=document.createElement("div");if(j){post.innerHTML=j.PostMarkup;inject_hover_for_element(post);}else{post.setAttribute("class","notfound post");post.appendChild(document.createTextNode("post not found"));}
parent.appendChild(post);}
/* ./contrib/js/backlink.js */
var dynreply;function getReplyTo(){if(!dynreply){var e=document.getElementById("postform_container");if(e){dynreply=new DynReply(e);}else{dynreply=new DynReply();}}
return dynreply;}
function table_insert_row(table,header,items){var tr=document.createElement("tr");var th=document.createElement("th");th.appendChild(header);tr.appendChild(th);for(var idx=0;idx<items.length;idx++){var elem=document.createElement("td");elem.appendChild(items[idx]);tr.appendChild(elem);}
table.appendChild(tr);}
function DynReply(existingElem){if(existingElem){this.elem=existingElem;this.form=this.elem.querySelector("form");return;}
var elem=document.createElement("div");elem.setAttribute("id","postform_container");this.elem=elem;this.form=document.createElement("form");this.form.enctype="multipart/form-data";this.form.name="post";this.form.method="post";elem=document.createElement("input");elem.setAttribute("id","postform_reference");elem.name="reference";elem.type="hidden";this.form.appendChild(elem);var table=document.createElement("table");table.setAttribute("class","postform");var tbody=document.createElement("tbody");elem=document.createElement("input");elem.setAttribute("name","name");elem.setAttribute("value","Anonymous");table_insert_row(tbody,document.createTextNode("Name"),[elem])
elem=document.createElement("input");elem.setAttribute("name","subject");elem.setAttribute("value","");var submit=document.createElement("input");submit.setAttribute("type","submit");submit.setAttribute("value","reply");submit.setAttribute("class","button");table_insert_row(tbody,document.createTextNode("Subject"),[elem,submit]);elem=document.createElement("textarea");elem.setAttribute("id","postform_message");elem.setAttribute("name","message");elem.setAttribute("cols","40");elem.setAttribute("rows","5");table_insert_row(tbody,document.createTextNode("Comment"),[elem]);elem=document.createElement("input");elem.setAttribute("class","postform_attachment");elem.setAttribute("id","postform_attachments");elem.setAttribute("type","file");elem.setAttribute("name","attachment_uploaded");elem.setAttribute("multiple","multiple");table_insert_row(tbody,document.createTextNode("Files"),[elem]);elem=document.createElement("input");elem.setAttribute("type","checkbox");elem.setAttribute("name","dubs");table_insert_row(tbody,document.createTextNode("Get Dubs"),[elem]);elem=document.createElement("img");elem.setAttribute("id","captcha_img");elem.alt="captcha";table_insert_row(tbody,document.createTextNode("Captcha"),[elem]);elem=document.createElement("input");elem.name="captcha";elem.autocomplete="off";table_insert_row(tbody,document.createTextNode("Name"),[elem])
table.appendChild(tbody);this.form.appendChild(table);this.elem.appendChild(this.form);document.body.appendChild(this.elem);this.board=null;this.roothash=null;this.prefix=null;}
DynReply.prototype.update=function(){if(this.prefix){this.updateCaptcha();if(this.board&&this.roothash){var ref=document.getElementById("postform_reference");ref.setAttribute("value",this.roothash);this.form.action=this.prefix+"post/"+this.board;}}}
DynReply.prototype.show=function(){console.log("show dynreply");this.update();this.elem.style.display='inline';}
DynReply.prototype.updateCaptcha=function(){if(this.prefix){var captcha_img=document.getElementById("captcha_img");captcha_img.src=this.prefix+"captcha/img";}}
DynReply.prototype.setPrefix=function(prefix){this.prefix=prefix;}
DynReply.prototype.hide=function(){this.elem.style.display='none';}
DynReply.prototype.setBoard=function(boardname){if(boardname){this.board=boardname;}}
DynReply.prototype.setRoot=function(roothash){if(roothash){this.roothash=roothash;}}
function nntpchan_reply(prefix,parent,shorthash){if(prefix&&parent){var boardname=parent.getAttribute("boardname");var roothash=parent.getAttribute("root");var replyto=getReplyTo();replyto.setBoard(boardname);replyto.setRoot(roothash);replyto.setPrefix(prefix);replyto.show();}
var elem=document.getElementById("postform_message");if(elem)
{elem.value+=">>"+shorthash.substr(0,10)+"\n";}}
function inject_hover(prefix,el,parent){if(!prefix){throw"prefix is not defined";}
var linkhash=el.getAttribute("backlinkhash");if(!linkhash){throw"linkhash undefined";}
console.log("rewrite linkhash "+linkhash);var elem=document.createElement("span");elem.setAttribute("class","backlink_rewritten");elem.appendChild(document.createTextNode(">>"+linkhash.substr(0,10)));if(!parent){parent=el.parentNode;}
parent.removeChild(el);parent.appendChild(elem);elem.onclick=function(ev){if(parent.backlink){nntpchan_apicall(prefix+"api/find?hash="+linkhash,function(j){var wrapper=document.createElement("div");wrapper.setAttribute("class","hover "+linkhash);if(j==null){wrapper.setAttribute("class","hover notfound-hover "+linkhash);wrapper.appendChild(document.createTextNode("not found"));}else{nntpchan_buildpost(wrapper,j);}
parent.appendChild(wrapper);parent.backlink=false;},function(msg){var wrapper=document.createElement("div");wrapper.setAttribute("class","hover "+linkhash);wrapper.appendChild(document.createTextNode(msg));parent.appendChild(wrapper);parent.backlink=false;});}else{var elems=document.getElementsByClassName(linkhash);if(!elems)throw"bad state, no backlinks open?";for(var idx=0;idx<elems.length;idx++){elems[idx].parentNode.removeChild(elems[idx]);}
parent.backlink=true;}};parent.backlink=true;}
function inject_hover_for_element(elem){var elems=elem.getElementsByClassName("backlink");var ls=[];var l=elems.length;for(var idx=0;idx<l;idx++){var e=elems[idx];ls.push(e);}
for(var elem in ls){inject_hover(prefix,ls[elem]);}}
function init(prefix){inject_hover_for_element(document);var rpl=getReplyTo();rpl.setPrefix(prefix);var e=rpl.elem;e.setAttribute("draggable","true");e.ondragend=function(ev){ev.preventDefault();var el=document.getElementById("postform_container");el.setAttribute("style","top: "+ev.y+"; left: "+ev.x+"; position: fixed; ");}}
/* ./contrib/js/banner.js */
var banner_count=3;function nntpchan_inject_banners(elem,prefix){var n=Math.floor(Math.random()*banner_count);var banner=prefix+"static/banner_"+n+".jpg";var e=document.createElement("img");e.src=banner;e.id="nntpchan_banner";elem.appendChild(e);}
/* ./contrib/js/expand-image.js */
function filenameIsImage(fname){return/\.(gif|jpeg|jpg|png|webp)/.test(fname);}
function setupInlineImage(thumb,url){if(thumb.inlineIsSetUp)return;thumb.inlineIsSetUp=true;var img=thumb.querySelector("img.thumbnail");var expanded=false;var oldurl=img.src;thumb.onclick=function(){if(expanded){img.setAttribute("class","thumbnail");img.src=oldurl;expanded=false;}else{img.setAttribute("class","expanded-thumbnail");img.src=url;expanded=true;}
return false;}}
function setupInlineImageIn(element){var thumbs=element.querySelectorAll("a.file");for(var i=0;i<thumbs.length;i++){var url=thumbs[i].href;if(filenameIsImage(url)){console.log("matched url",url);setupInlineImage(thumbs[i],url);}}}
onready(function(){setupInlineImageIn(document);if(window.MutationObserver){var observer=new MutationObserver(function(mutations){for(var i=0;i<mutations.length;i++){var additions=mutations[i].addedNodes;if(additions==null)continue;for(var j=0;j<additions.length;j++){var node=additions[j];if(node.nodeType==1){setupInlineImageIn(node);}}}});observer.observe(document.body,{childList:true,subtree:true});}});
/* ./contrib/js/expand-video.js */
var configRoot="";if(typeof _=='undefined'){var _=function(a){return a;};}
function setupVideo(thumb,url){if(thumb.videoAlreadySetUp)return;thumb.videoAlreadySetUp=true;var video=null;var videoContainer,videoHide;var expanded=false;var hovering=false;var loop=true;var loopControls=[document.createElement("span"),document.createElement("span")];var fileInfo=thumb.parentNode.querySelector(".fileinfo");var mouseDown=false;function unexpand(){if(expanded){expanded=false;if(video.pause)video.pause();videoContainer.style.display="none";thumb.style.display="inline";video.style.maxWidth="inherit";video.style.maxHeight="inherit";}}
function unhover(){if(hovering){hovering=false;if(video.pause)video.pause();videoContainer.style.display="none";video.style.maxWidth="inherit";video.style.maxHeight="inherit";}}
function getVideo(){if(video==null){video=document.createElement("video");video.src=url;video.loop=loop;video.innerText=_("Your browser does not support HTML5 video.");videoHide=document.createElement("img");videoHide.src=configRoot+"static/collapse.gif";videoHide.alt="[ - ]";videoHide.title="Collapse video";videoHide.style.marginLeft="-15px";videoHide.style.cssFloat="left";videoHide.addEventListener("click",unexpand,false);videoContainer=document.createElement("div");videoContainer.style.paddingLeft="15px";videoContainer.style.display="none";videoContainer.appendChild(videoHide);videoContainer.appendChild(video);thumb.parentNode.insertBefore(videoContainer,thumb.nextSibling);video.addEventListener("mousedown",function(e){if(e.button==0)mouseDown=true;},false);video.addEventListener("mouseup",function(e){if(e.button==0)mouseDown=false;},false);video.addEventListener("mouseenter",function(e){mouseDown=false;},false);video.addEventListener("mouseout",function(e){if(mouseDown&&e.clientX-video.getBoundingClientRect().left<=0){unexpand();}
mouseDown=false;},false);}}
thumb.addEventListener("click",function(e){if(!e.shiftKey&&!e.ctrlKey&&!e.altKey&&!e.metaKey){getVideo();expanded=true;hovering=false;video.style.position="static";video.style.pointerEvents="inherit";video.style.display="inline";videoHide.style.display="inline";videoContainer.style.display="block";videoContainer.style.position="static";video.parentNode.parentNode.removeAttribute('style');thumb.style.display="none";video.controls=true;if(video.readyState==0){video.addEventListener("loadedmetadata",expand2,false);}else{setTimeout(expand2,0);}
video.play();e.preventDefault();}},false);function expand2(){video.style.maxWidth="100%";video.style.maxHeight=window.innerHeight+"px";var bottom=video.getBoundingClientRect().bottom;if(bottom>window.innerHeight){window.scrollBy(0,bottom-window.innerHeight);}}
thumb.addEventListener("mouseover",function(e){if(false){getVideo();expanded=false;hovering=true;var docRight=document.documentElement.getBoundingClientRect().right;var thumbRight=thumb.querySelector("img, video").getBoundingClientRect().right;var maxWidth=docRight-thumbRight-20;if(maxWidth<250)maxWidth=250;video.style.position="fixed";video.style.right="0px";video.style.top="0px";var docRight=document.documentElement.getBoundingClientRect().right;var thumbRight=thumb.querySelector("img, video").getBoundingClientRect().right;video.style.maxWidth=maxWidth+"px";video.style.maxHeight="100%";video.style.pointerEvents="none";video.style.display="inline";videoHide.style.display="none";videoContainer.style.display="inline";videoContainer.style.position="fixed";video.controls=false;video.play();}},false);thumb.addEventListener("mouseout",unhover,false);thumb.addEventListener("wheel",function(e){if(true){if(e.deltaY>0)volume-=0.1;if(e.deltaY<0)volume+=0.1;if(volume<0)volume=0;if(volume>1)volume=1;if(video!=null){video.muted=(volume==0);video.volume=volume;}
e.preventDefault();}},false);}
function setupVideosIn(element){var thumbs=element.querySelectorAll("a.file");for(var i=0;i<thumbs.length;i++){if(/(\.webm)|(\.mp4)$/.test(thumbs[i].pathname)){setupVideo(thumbs[i],thumbs[i].href);}else{var url=thumbs[i].href;if(/(\.webm)|(\.mp4)$/.test(url))setupVideo(thumbs[i],url);}}}
onready(function(){if(typeof settingsMenu!="undefined"&&typeof Options=="undefined")
document.body.insertBefore(settingsMenu,document.getElementsByTagName("hr")[0]);setupVideosIn(document);if(window.MutationObserver){var observer=new MutationObserver(function(mutations){for(var i=0;i<mutations.length;i++){var additions=mutations[i].addedNodes;if(additions==null)continue;for(var j=0;j<additions.length;j++){var node=additions[j];if(node.nodeType==1){setupVideosIn(node);}}}});observer.observe(document.body,{childList:true,subtree:true});}});
/* ./contrib/js/livechan.js */
function livechan_got_post(widget,j){while(widget.children.length>5){widget.removeChild(widget.children[0]);}
nntpchan_buildpost(widget,j);widget.scrollTop=widget.scrollHeight;}
function inject_postform(prefix,parent){}
function inject_livechan_widget(prefix,parent){if("WebSocket"in window){var url="ws://"+document.location.host+prefix+"live";if(document.location.protocol=="https:"){url="wss://"+document.location.host+prefix+"live";}
var socket=new WebSocket(url);var progress=function(str){parent.innerHTML="<pre>livechan: "+str+"</pre>";};progress("initialize");socket.onopen=function(){progress("streaming (read only)");}
socket.onmessage=function(ev){var j=null;try{j=JSON.parse(ev.data);}catch(e){}
if(j){livechan_got_post(parent,j);}}
socket.onclose=function(ev){progress("connection closed");setTimeout(function(){inject_livechan_widget(prefix,parent);},1000);}}else{parent.innerHTML="<pre>livechan mode requires websocket support</pre>";setTimeout(function(){parent.innerHTML="";},5000);}}
function ukko_livechan(prefix){var ukko=document.getElementById("ukko_threads");if(ukko){ukko.innerHTML="";inject_livechan_widget(prefix,ukko);}}
/* ./contrib/js/local_storage.js */
function get_storage(){var st=null;if(window.localStorage){st=window.localStorage;}else if(localStorage){st=localStorage;}
return st;}
/* ./contrib/js/theme.js */
function enable_theme(prefix,name){if(prefix&&name){var theme=document.getElementById("current_theme");if(theme){theme.href=prefix+"static/"+name+".css";var st=get_storage();st.nntpchan_prefix=prefix;st.nntpchan_theme=name;}}}
var st=get_storage();enable_theme(st.nntpchan_prefix,st.nntpchan_theme);