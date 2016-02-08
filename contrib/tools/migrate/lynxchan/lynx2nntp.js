#!/usr/bin/env nodejs

var memegod = require("mongodb");
var memegodClient = memegod.MongoClient;
var assert = require("assert");
var http = require("http");


// memegod daemon url
var url = "mongodb://127.0.0.1:27017/lynxchan";

// nntp frontend name for article import
var frontendName = "memegod.import";
// srndv2 http frontend server to do api requests against 
var srndApiServ = "localhost";
var srndApiPort = 18000;
// username password combo for using api
var srndApiLogin = "user:pass";
// some memegod thing
var idPrefix="v1-";


// create a newsgroup name given board name
var makeGroupName = function(board) {
  return "overchan.test.endchan." + board.boardUri;
}

// call function for each op on a board
var foreachOpsOnBoard = function(db, board, callback) {
  var cur = db.collection("threads").find({boardUri: board});
  var ops = [];
  cur.each(function(err, doc) {
    if (doc) {
      ops.push(doc)
    } else {
      for (var idx = 0 ; ops.length > idx ; idx ++ ) {
        callback(op[idx]);
      }
    }
  });
}

// call a callback for each reply to op
// pass in the memegod post
var foreachReplyForOP = function(db, op, callback) {
  checkViaHeader("X-Mememgod-Id", idPrefix+op._id, function(msgid) {
    // we do already has got it
    callback(null);
  }, function(msgid) {
    // we don't has got it
    var cur = db.collection("posts").find({ threadId: op.threadId});
    var repls = [];
    cur.each(function(err, doc) {
      if (doc) {
        repls.push(doc)
      } else {
        for (var idx = 0 ; repls.length > idx ; idx ++ ) {
          callback(repls[idx]);
        }
      }
    });
  });
}

// find all boards in memegod
// call callback for each board
var foreachBoard = function(db, callback, done) {
  var cursor = db.collection('boards').find();
  var boards = [];
  cursor.each(function(err, doc) {
    if (doc) {
      boards.push(doc)
    } else {
      for (var idx = 0 ; boards.length > idx ; idx ++ ) {
        callback(boards[idx]);
      }
      done(db);
    }
  });
};


// convert a memegod post from board into an overchan article
// call a callback with the created post
var createArticle = function(post, board, callback) {
  if (post == null) return;
      
  var article = {
    ip: post.ip.join("."),
    message: post.message || " ",
    frontend: frontendName,
    newsgroup: makeGroupName(board),
    headers: {
    }
  };
  if (post.postId) {
    article.headers["X-Memegod-Post-Id"] = idPrefix+post.postId
  }
  if (post.threadId) {
    article.headers["X-Memegod-Thread-Id"] = idPrefix+post.threadId;
  }
  article.headers["X-Memegod-Id"] = idPrefix+post._id;
  article.headers["X-Migrated-From"] = "MemeGod";
  article.name = post.name || "Stephen Lynx";
  article.subject = post.subject || "MongoDB is Web Scale";
  callback(article);
}

// post an overchan article via the api
// call callback passing in the message-id of the new post
var postArticle = function(article, callback) {

  checkViaHeader("X-Memegod-Id", article.headers["X-Memegod-Id"], function(msgid) {
    // we has got it already
    callback(msgid);
  }, function(msgid) {
    // we don't has got it
    var req = http.request({
      port: srndApiPort,
      method: "POST",
      path: "/api/post",
      auth: srndApiLogin,
      headers: {
        "Content-Type": "text/json",
      }
    }, function(res) {
      var data = "";
      res.on("data", function (chunk) {
        data += chunk;
      });
      res.on("end", function() {
        var j = JSON.parse(data)
        var msgid = j.id;
        callback(msgid);
        
      })
    });
    req.write(JSON.stringify(article));
    req.end();
  }, function(msgid) {
    // not there
  });
}

// check if an article exists given header name and header value
var checkViaHeader = function(name, value, yesCb, noCb) {
  var req = http.request({
    port: srndApiPort,
    method: "GET",
    path: "/api/header?name="+name+"&value="+value,
  }, function (res) {
    var data = "";
    res.on("data", function(chnk) {
      data += chnk;
    });
    res.on("end", function() {
      var j = JSON.parse(data);
      if ( j.length > 0 ) {
        // it exists
        yesCb(j[0]);
      } else {
        // does not exist
        noCb();
      }
    });
  });
  req.end();
}

// check if a post exists 
var checkPostExists = function(post, yescb, nocb) {
  checkViaHeader("X-Memegod-Id", idPrefix + post._id,
                 function(msgid) { yescb(post, msgid); },
                 function() { nocb(post); });
}

var putBoard = function(db, board) {
  foreachOpsOnBoard(db, board.boardUri, function(op) {
    // create OP
    createArticle(op, board, function(opArticle) {
      // post OP
      postArticle(opArticle, function (opMsgId) {
        // for each reply for OP
        foreachReplyForOP(db, op, function(post) {
          if (post) {
            // put create reply
            createArticle(post, board, function(article) {
              // set references header
              article.headers["References"] = opMsgId;
              // post reply
              postArticle(article, function() {});
            });
          }
        });
      });              
    });
  });
}

memegodClient.connect(url, function(err, db) {
  console.log("connected to the meme god");
  foreachBoard(db, function(board) {
    console.log("updating "+board.boardUri);
    putBoard(db, board);
  }, function(db) {
    console.log("the meme god has spoken");
    db.close();
  });
});
