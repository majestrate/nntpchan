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
  return "overchan.test.endchan." + board;
}

// call function for each op on a board
var foreachOpsOnBoard = function(db, board, callback) {
  var cur = db.collection("threads").find({boardUri: board});
  cur.each(function(err, doc) {
    if (doc) {
      callback(doc);
    } else {
      console.log("error fetching op on "+board+" "+err);
    }
  });
}

// call a callback for each reply to op
// pass in the memegod post
var foreachReplyForOP = function(db, op, callback) {
  getMessageIDForMemegodID(idPrefix+op._id, function(msgid) {
    var cur = db.collection("posts").find({ threadId: op.threadId});
    cur.each(function (err, doc) {
      callback(doc);
    });
  }, function() {
    // not found?
    console.log("op not in database!? "+op._id);
  });
}

// find all boards in memegod
// call callback for each board
var foreachBoard = function(db, callback) {
   var cursor = db.collection('boards').find();
   cursor.each(function(err, doc) {
     if (doc) {
       callback(doc)
     }
   });
};


// convert a memegod post from board into an overchan article
// call a callback with the created post
var createArticle = function(post, board, callback) {
  var article = {
    ip: post.ip.join("."),
    message: post.message || " ",
    frontend: frontendName,
    newsgroup: makeGroupName(board),
    headers: {
    }
  };
  article.headers["X-Memegod-Post-Id"] = post.postId;
  article.headers["X-Memegod-Thread-Id"] = post.threadId;
  article.headers["X-Memegod-Id"] = idPrefix+post._id;
  article.headers["X-Migrated-From"] = "MemeGod";
  article.name = post.name || "Stephen Lynx";
  article.subject = post.subject || "MongoDB is Web Scale";
  
  callback(article);
}

// get message id given a memegod id
// must have already been inserted
// call callback pass in message id
var getMessageIDForMemegodID = function(id, callback) {
  checkViaHeader("X-Memegod-Id", id, function(msgid) {
    callback(msgid);
  }, function() { console.log("message is not there?! "+id); });
}

// post an overchan article via the api
// call callback passing in the message-id of the new post
var postArticle = function(article, callback) {
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
      if (j) {
        var msgid = j[0];
        callback(msgid);
      }
    })
  });
  req.write(json.stringify(article));
  req.end();
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
}

// check if a post exists 
var checkPostExists = function(post, yescb, nocb) {
  checkViaHeader("X-Memegod-Id", idPrefix + post._id,
                 function(msgid) { yescb(post, msgid); },
                 function() { nocb(post); });
}

var putBoard = function(db, board) {
  // for each op
  foreachOpsOnBoard(db, board.boardUri, function(op) {
    // create OP
    createArticle(op, board.boardUrl, function(opArticle) {
      // post OP
      postArticle(opArticle, function (opMsgId) {
        console.log("posted op "+ opMsgId);
        // for each reply for OP
        foreachReplyForOP(db, op, function(post) {
          // put create reply
          createArticle(post, board.boardUri, function(article) {
            // set references header
            article.headers["References"] = opMsgId;
            postArticle(article, function(replyMsgId) {
              console.log("posted reply to "+opMsgId+" as "+replyMsgId);
            });
          });
        });
      });              
    });
  });
}

memegodClient.connect(url, function(err, db) {
  console.log("connected to the meme god");
  foreachBoard(db, function(board) {
    console.log("put board: "+board.boardUri);
    putBoard(db, board);
  });
});
