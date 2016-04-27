//For source code and license information please check https://github.com/majestrate/nntpchan 

function nntpchan_apicall(url,handler,err_handler){var ajax=new XMLHttpRequest();ajax.onreadystatechange=function(){if(ajax.readyState==XMLHttpRequest.DONE){var status=ajax.status;var j=null;if(status==200){try{j=JSON.parse(ajax.responseText);}catch(e){}}else if(status==410){if(err_handler){err_handler("cannot fetch post: api disabled");}
return;}
handler(j);}};ajax.open("GET",url);ajax.send();}
function nntpchan_buildpost(parent,j){var post=document.createElement("div");if(j){post.innerHTML=j.PostMarkup;inject_hover_for_element(post);}else{post.setAttribute("class","notfound post");post.appendChild(document.createTextNode("post not found"));}
parent.appendChild(post);}function nntpchan_backlink(shorthash){var elem=document.getElementById("postform_message");if(elem)
{elem.value+=">>"+shorthash.substr(0,10)+"\n";}}
function inject_hover(prefix,el,parent){if(!prefix){throw"prefix is not defined";}
var linkhash=el.getAttribute("backlinkhash");if(!linkhash){throw"linkhash undefined";}
console.log("rewrite linkhash "+linkhash);var elem=document.createElement("span");elem.setAttribute("class","backlink_rewritten");elem.appendChild(document.createTextNode(">>"+linkhash.substr(0,10)));if(!parent){parent=el.parentNode;}
parent.removeChild(el);parent.appendChild(elem);elem.onclick=function(ev){if(parent.backlink){nntpchan_apicall(prefix+"api/find?hash="+linkhash,function(j){var wrapper=document.createElement("div");wrapper.setAttribute("class","hover "+linkhash);if(j==null){wrapper.setAttribute("class","hover notfound-hover "+linkhash);wrapper.appendChild(document.createTextNode("not found"));}else{nntpchan_buildpost(wrapper,j);}
parent.appendChild(wrapper);parent.backlink=false;},function(msg){var wrapper=document.createElement("div");wrapper.setAttribute("class","hover "+linkhash);wrapper.appendChild(document.createTextNode(msg));parent.appendChild(wrapper);parent.backlink=false;});}else{var elems=document.getElementsByClassName(linkhash);if(!elems)throw"bad state, no backlinks open?";for(var idx=0;idx<elems.length;idx++){elems[idx].parentNode.removeChild(elems[idx]);}
parent.backlink=true;}};parent.backlink=true;}
function inject_hover_for_element(elem){var elems=elem.getElementsByClassName("backlink");var ls=[];var l=elems.length;for(var idx=0;idx<l;idx++){var e=elems[idx];ls.push(e);}
for(var elem in ls){inject_hover(prefix,ls[elem]);}}
function init(prefix){inject_hover_for_element(document);}var banner_count=3;function nntpchan_inject_banners(elem,prefix){var n=Math.floor(Math.random()*banner_count);var banner=prefix+"static/banner_"+n+".jpg";var e=document.createElement("img");e.src=banner;e.id="nntpchan_banner";elem.appendChild(e);}function livechan_got_post(widget,j){while(widget.children.length>5){widget.removeChild(widget.children[0]);}
nntpchan_buildpost(widget,j);widget.scrollTop=widget.scrollHeight;}
function inject_postform(prefix,parent){}
function inject_livechan_widget(prefix,parent){if("WebSocket"in window){var url="ws://"+document.location.host+prefix+"live";if(document.location.protocol=="https:"){url="wss://"+document.location.host+prefix+"live";}
var socket=new WebSocket(url);var progress=function(str){parent.innerHTML="<pre>livechan: "+str+"</pre>";};progress("initialize");socket.onopen=function(){progress("streaming (read only)");}
socket.onmessage=function(ev){var j=null;try{j=JSON.parse(ev.data);}catch(e){}
if(j){console.log(j);livechan_got_post(parent,j);}}
socket.onclose=function(ev){progress("connection closed");setTimeout(function(){inject_livechan_widget(prefix,parent);},1000);}}else{parent.innerHTML="<pre>livechan mode requires websocket support</pre>";setTimeout(function(){parent.innerHTML="";},5000);}}
function ukko_livechan(prefix){var ukko=document.getElementById("ukko_threads");if(ukko){ukko.innerHTML="";inject_livechan_widget(prefix,ukko);}}function get_storage(){var st=null;if(window.localStorage){st=window.localStorage;}else if(localStorage){st=localStorage;}
return st;}function enable_theme(prefix,name){if(prefix&&name){var theme=document.getElementById("current_theme");if(theme){theme.href=prefix+"static/"+name+".css";var st=get_storage();st.nntpchan_prefix=prefix;st.nntpchan_theme=name;}}}
var st=get_storage();enable_theme(st.nntpchan_prefix,st.nntpchan_theme);