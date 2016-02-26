Overchan is a newsgroup meant to be served on web frontends in an effort to create a decentralized imageboard. Moderation takes place on each frontend itself. Message and image transport is using MIME multipart messages and Base64 as encoding for Images. All messages need to be valid NNTP messages, the transport of messages need to follow NNTP specifications. It is possible to use an existing NNTP daemon like INN or to implement the NNTP sync part as well.

# Sync Protocol (NNTP) 

## Article Format  

### Monopart

Message without images can be sent without delimiting the message.


    Content-Type: text/plain; charset=UTF-8
    Content-Transfer-Encoding: 8bit
    From: anonymous <foo@bar.ano>
    Date: Thu, 02 May 2013 12:16:44 +0000
    Message-ID: <pmc8xpmgyf2@foo.bar>
    Newsgroups: overchan.test
    Subject: none
    References: <referenced message-id>
    Path: hschan.ano
    X-Sage: optional

    some visible message text


### Multipart 

This is necessary for posting files.

    Mime-Version: 1.0
    Content-Type: multipart/mixed; boundary="abcdEFGH-1234"
    From: anonymous <foo@bar.ano>
    Date: Thu, 02 May 2013 12:16:44 +0000
    Message-ID: <pmc8xpmgyf2@foo.bar>
    Newsgroups: overchan.test
    Subject: none
    References: <referenced message-id>
    Path: hschan.ano
    X-Sage: optional

    This is a multi-part message in MIME format.
    --abcdEFGH-1234
    Content-Type: text/plain; charset=UTF-8
    Content-Transfer-Encoding: 8bit
    some visible message text
    --abcdEFGH-1234
    Content-Type: image/jpeg; name="RosenFessel_LM_030-lg.jpg"
    Content-Transfer-Encoding: base64
    Content-Disposition: attachment; filename="RosenFessel_LM_030-lg.jpg"
    /9j/4AAQSkZJRgABAQAAAQABAAD//gA8Q1JFQVRPUjogZ2QtanBlZyB2MS4wICh1c2luZyBJ
    SkcgSlBFRyB2NjIpLCBxdWFsaXR5ID0gMTAwCv/bAEMAAQEBAQEBAQEBAQEBAQEBAQEBAQEB
    [..]
    e3ykVkO1lOwlSMkfvDngbSOmRuqaDQo2Kqi5yUwfKPXkk8kAjIA5B4wPUVi8TKblJ1Oy969r
    ea8+nXc6I4WnFLROMVZRai1L4bK+vdfM/9k=
    --abcdEFGH-1234-
    Content-Type boundary="$IDENTIFIER"


Where ``$identifier`` should be a rather long random string (at least 0-9, a-z, A-Z, - are allowed). The ``$identifier`` should not occur in the message text itself, so it usually begins with multiple - characters because these will never occur in base64. The content type ``multipart/mixed`` allows to have different parts inside a message body. The first part would be the actual message, the second part could be a Base64 encoded picture. See also: MIME


In ~~2013~~ 2015 we can send UTF-8 messages, although this is not part of the old testament.

### Date 

It is recommend to use ``UTC (+0000)`` as timezone for new messages. If a received message is not already in UTC, the date may be converted to UTC for display purposes.

### Message-ID: (see RFC 3977) 

    "A message-id MUST begin with "<", end with ">", and MUST NOT contain the latter except at the end."
    "A message-id MUST be between 3 and 250 octets in length."
    "A message-id MUST NOT contain octets other than printable US-ASCII characters."
    a possible valid message-id could be in the format <{random}{timestamp}@${frontend}> where:
        ${random} == a random 10 char ascii value
        ${timestamp} == the current unix_timestamp
        ${frontend} == web.hschan.ano

Which would result in ``jbUdn73KxN1369733675@web.hschan.ano`` This format makes it easier to block massive spam/inapropiate content based on the frontend and a timespan.

### References 

If reference is not given or empty, the message is considered an original (root) post.

### X-Sage 

If ``X-Sage`` is given, the message shall not bump the corresponding thread.

## Transport Format 

NNTP requires line endings with ``\r\n``

### Sending

if a line in the message body starts with `.` in needs another `.` prepended. the last line must be a single `.\r\n`
 
### Receiving: 

If a line in the message body starts with `.` but is not `.\r\n` the `.` needs to be removed.


# Frontend 

## Postnumbers 

The first ten characters of a sha1sum of field message-id. The probability for a unique post number (at time of generation) on a board with a maximum of 30k messages is:

    (1-(1/16^10))^30000 = 0.99999997271515937221

In case of several message forgers exhaust obscure post numbers, it will become much more likely for a quote to be 'shadowed'.

There was a hash collision on October 13 2015, the post hash has been bumped from 10 to 18 bytes.

### Quotes 

Quotes reference postnumbers and work across all boards on overchan. The comment field may contain serveral lines such as:

\>>postnumber

to quote someone.

Valid quotes match this regex: `>+ ?[0-9a-f]+`

#### Optional

* Resolve quotes to corresponding articles and append them to references - this will aid newsreaders. 
* Parse message IDs as quotes

## Implementations 

Because of the decentralized nature of Overchan, many different entry points to using the service can exist. In the following we discuss different implementations all serving from the newsgroup 'overchan'.

### negromancy.ano 

negromancy.ano uses `breaking-news`, a web frontend compiler for imageboards, pastebins, etc.

`breaking-news` consists of an Happstack application and a daemon that will generate static html from NNTP files. It depends on InterNetNews (INN) and a load balancer that can distinguish between POST and GET, preferably nginx. Furthermore it relies on imagemagick (mogrify) for generation of thumbnails. Hchloride are bindings to libsodium in haskell that breaking-news uses for singing messages.

It utilizes blaze-html for fast Html templating and happstack-lite for serving POST requests.

You can visit http://boards.negromancy.ano/

for browsing the imageboard.

#### GET request 

nginx will serve a static html from dir.

#### POST request 

nginx will reverse to the happstack web application which generates and sends a NNTP message to a local INN daemon. INNd will place the new article in dir and feed it to its configured peers. Pictures are encoded in base64 or base91a. Root posting will not work without attaching an image.

#### Generation of Html 

The daemon will poll for new articles in dir. If new files are found, it generates 10 main pages ranging from 0.html to 9.html and a html each for any altered threads. For each new article the corresponding thread, starting with the original post, will be bumped to the first page (0.html). For new posts, corresponding pictures are created and 'thumbnailed' through mogrify.

### overchan.sfor.ano 

overchan.sfor.ano uses SRNd, a complete NNTP server implemented in Python.

It provides a plugin interface (among other hook possibilities) which loads plug-in overchan and postman. Plug-in overchan is notified about new messages in `overchan.*` and creates static HTML files. Plug-in postman receives new messages via HTTP POST request and adds those messages to SRNd where they are send to configured outfeeds. It depends on a reverse proxy like nginx which delivers generated HTML files and proxies POST requests back to postman.

You can visit http://overchan.sfor.ano

for browsing the imageboard and ``git clone git://git.sfor.ano/SRNd.git`` for source.

#### GET request 

nginx will serve a static html or image from dir.

#### POST request 

nginx will proxy to postman which generates and delivers a NNTP message to SRNd which then will notify overchan plugin about the new message and also deliver it to its configured NNTP peers (which can run SRNd or another NNTPd software like INN). Pictures are encoded in base64.

#### Generation of Html 

Plugin overchan is notified by SRNd about new articles and (re)generates `thread-$id.html` and its parent board with up to 10 root posts for each site. For each new article without `X-sage` header the corresponding thread will be bumped to the first page. For new posts, corresponding pictures are created and thumbnailed.

### NNTP News reader applications 

Through the use of the standard MIME format, news reader applications like Mozilla Thunderbird can also read and post directly to the chan newsserver. Each chan will appear as a root post, while additional posts will appear as replies directly to the root post.

News readers have some features the chan software may not have: multiple attachments, non-image attachments, subject, posts referencing non-root posts, HTML text. 

Open question: how should this be handled by the chan software for viewing?

# Extensions

## Control suggestion 

A control suggestion is a single message containing lines with commands, message-ID and extra information separated by spaces.

### Commands

    sticky: sticky this thread
    delete-x-all: delete all attachments from this article
    delete: delete the whole article


### Format 

    Content-Type: text/plain; charset=UTF-8
    Content-Transfer-Encoding: 8bit
    From: anonymous <foo@bar.ano>
    Date: Thu, 02 May 2013 12:16:44 +0000
    Message-ID: <h2cykk1lwlmuqao2qiy@foo.bar>
    Newsgroups: ctl
    Subject: none
    Path: censorship.fleet
    X-Sage: optional

    delete-x-all <message-ID>
    delete <message-ID>
    delete <message-ID>


Messages to control are separated by at least one line break.


### Examples

Delete all attachments from message with ID ``message-ID``

    delete-x-all <message-ID>

Please sticky thread with OP ``message-ID`` till UNIX timestamp ``1380000000``

    sticky <message-ID> unix_timestamp 1380000000




### Convention 

We send control suggestions to newsgroup ``ctl``. Full deletion of a root post results in removal of corresponding thread.

### Signatures 

As users give their secret key to the frontend they expect every form field to be verified on all ends. This includes the comment field and headers. In the following we suggest a protocol to sign optional headers.

We sign a SHA512 hash of the message body using primitive Ed25519 as defined by SUPERCOP and libsodium. Therefore this system does not inherit any collision resilience from Ed25519, a hash collision is a signature collision.

Signing M vs. Signing H(M) 

    method	space	time

    S(M)	O(n)	O(n)

    S(H(M))	O(1)	O(n)

    S: Sign
    H: Hash
    M: Message


Input for block based hashing algorithms like SHA-512 can be streamed, only keeping a fixed blocked size in memory instead of all blocks. In case of SHA-512 these message blocks are 1024 bit and the hash to sign 512 bit. Optimized ``S(H(M))`` implementations require a constant amount memory as opposed to a linear requirement in ``S(M)``.

### Format for signing messages (RFC 822) 

Outer headers start with ``Content-Type: message/rfc822`` when there are signed headers or at least an attachment,
which requires ``Content-Type: multipart/mixed`` to be signed as well as an inner header.
Otherwise you can use ``Content-Type: text/plain``, in which case you just sign the body.
Outer headers include ``X-pubkey-ed25519`` and ``X-signature-ed25519-sha512``, inner headers need verification.
``X-pubkey-ed25519`` is 64 characters long, 32 byte public key in base 16: ``Base16(PK)``
``X-signature-ed25519-sha512`` is 128 characters long, 64 byte signature in base 16: ``Base16(S(SK,H(M)))``
The signed message equals body of the outer message. It begins at first inner header (in this example ``Content-Type: text/plain``) and includes the inner body. Lines are separated by ``<CRLF>``.
Please include a ``Content-Type`` header in the inner message as suggested by RFC822.

    Symbol	Function

    Base16	function that will take an arbitrary amount of octets and encode them to Base 16 with character set "0123456789abcdef"
    SK	64 bytes secret key
    PK	32 bytes public key, can be generated from signSeedKeypair(take32(SK))
    M	message body
    H	function that will hash an arbitrary amount of octets using SHA-512, returning 64 bytes
    S(SK,M)	function that will sign an arbitrary amount of octets M using Ed25519 with secret key SK, returning only the first 64 bytes
    take32	function that takes any amount of binary data and returns the first 32 bytes


#### Example

    Content-Type: message/rfc822; charset=UTF-8
    Content-Transfer-Encoding: 8bit
    From: anonymous <foo@bar.ano>
    Date: Thu, 02 May 2013 12:16:44 +0000
    Message-ID: <h2cykk1lwlmuqao2qiy@foo.bar>
    Newsgroups: ctl
    Subject: none
    Path: censorship.fleet
    X-pubkey-ed25519: 37c16fa40c2bade813b53b65107a064d02becfa5635acf3241003a61cb137ea3
    X-signature-ed25519-sha512: a850ccd788d71ed19de8dfa061b9f1f4f506810a01ed1391433e893a3e6305b4944168760d97f2517bcfe786aef1ccfc34fb7bb1b77531  82aebf2bdd0303150f
    
    Content-Type: text/plain; charset=UTF-8
    Date: Thu, 02 May 2013 12:16:44 +0000
    
    delete-x-all <message-ID>
    delete <message-ID>
    
    delete <message-ID


In this example header Date needs verification, too.
The following part is signed:


    Content-Type: text/plain; charset=UTF-8
    Date: Thu, 02 May 2013 12:16:44 +0000

    delete-x-all <message-ID>
    delete <message-ID>

    delete <message-ID


Above example in octets:

    Content-Type: text/plain; charset=UTF-8\\r\\nDate: Thu, 02 May 2013 12:16:44 +0000\\r\\n\\r\\ndelete-x-all <message-ID>\\r\\ndelete <message-ID>\\r\\n\\r\\ndelete <message-ID>
    
## RPC 
Remote procedure calls can be sent via ``ctl`` or on a group basis by using the group ``ctl.overchan.*``, where * is the group for which you want to execute a certain operation.

### Default format
The default format uses the MIME type ``text/plain`` where the first line of the body opens an array with ``[`` the next line is the name of the procedure you want to call, and on the lines following you can add one or more parameters. Each of these lines is terminated with `,` and indention can be added as well. The arry is closed with `]`.

#### Exapmple
    Content-Type: text/plain; charset=UTF-8
    Content-Transfer-Encoding: 8bit
    From: anonymous <foo@bar.ano>
    Date: Thu, 02 May 2013 12:16:44 +0000
    Message-ID: <h2cykk1lwlmuqao2qiy@foo.bar>
    Newsgroups: ctl.overchan.foo
    Subject: RPC
    Path: hschan.ano

    [
        setSetting,
        bumplimit,
        350,
    ]
    
    
### JSON RPC
If the MIME type is specified as ``application/json`` the body is interpreted as [JSON RPC](http://json-rpc.org/).

#### Exapmple
    Content-Type: application/json; charset=UTF-8
    Content-Transfer-Encoding: 8bit
    From: anonymous <foo@bar.ano>
    Date: Thu, 02 May 2013 12:16:44 +0000
    Message-ID: <h2cykk1lwlmuqao2qiy@foo.bar>
    Newsgroups: ctl.overchan.foo
    Subject: RPC
    Path: hschan.ano

    {"method": "setSetting", "params": ["bumplimit", "350"], "id": null}
    
### Additional details
As described above, muliple RPC's can be sent via the multipart format. It is also expected that these articles are signed.

# Glossary 

## chan specific

### root post
original post

### OP
original post

### thread
a collection of messages starting with the original post followed by messages referencing it ordered by date

### bump
newest post will be shown first with corresponding thread
    
### sticky
thread is temporarily 'bumped' by the frontend and sticks there regardless of newer posts
