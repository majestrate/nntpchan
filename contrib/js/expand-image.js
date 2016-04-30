//
// expand images inline
//
// released into the public domain by Jeff on 2016-04-30
//

// is the filename matching an image?
function filenameIsImage(fname) {
  return /\.(gif|jpeg|jpg|png|webp)/.test(fname);
}

// setup image inlining for 1 element
function setupInlineImage(thumb, url) {
  if(thumb.inlineIsSetUp) return;
  thumb.inlineIsSetUp = true;
  var img = thumb.querySelector("img.thumbnail");
  var expanded = false;
  var oldurl = img.src;
  thumb.onclick = function() {
    if (expanded) {
      img.setAttribute("class", "thumbnail");
      img.src = oldurl;
      expanded = false;
    } else {
      img.setAttribute("class", "expanded-thumbnail");
      img.src = url;
      expanded = true;
    }
    return false;
  }
}

// set up image inlining for all applicable children in an element
function setupInlineImageIn(element) {
  var thumbs = element.querySelectorAll("a.file");
  for ( var i = 0 ; i < thumbs.length ; i++ ) {
    var url = thumbs[i].href;
    if (filenameIsImage(url)) {
      // match
      console.log("matched url", url);
      setupInlineImage(thumbs[i], url);
    }
  }
}


onready(function(){

  // Setup Javascript events for document 
  setupInlineImageIn(document);
  
  
  // Setup Javascript events via updatoer
  if (window.MutationObserver) {
    var observer = new MutationObserver(function(mutations) {
      for (var i = 0; i < mutations.length; i++) {
        var additions = mutations[i].addedNodes;
        if (additions == null) continue;
        for (var j = 0; j < additions.length; j++) {
          var node = additions[j];
          if (node.nodeType == 1) {
            setupInlineImageIn(node);
          }
        }
      }
    });
    observer.observe(document.body, {childList: true, subtree: true});
  }
  
});
