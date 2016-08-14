<?php

function gennntp($headers, $files) {
	if (count($files) == 0) {
	}
	else if (count($files) == 1 && $files[0]['type'] == 'text/plain') {
		$content = $files[0]['text'] . "\r\n";
		$headers['Content-Type'] = "text/plain; charset=UTF-8";
	}
	else {
		$boundary = sha1($headers['Message-Id']);
		$content = "";
		$headers['Content-Type'] = "multipart/mixed; boundary=$boundary";
		foreach ($files as $file) {
			$content .= "--$boundary\r\n";
			if (isset($file['name'])) {
				$file['name'] = preg_replace('/[\r\n\0"]/', '', $file['name']);
				$content .= "Content-Disposition: form-data; filename=\"$file[name]\"; name=\"attachment\"\r\n";
			}
			$type = explode('/', $file['type'])[0];
			if ($type == 'text') {
				$file['type'] .= '; charset=UTF-8';
			}
			$content .= "Content-Type: $file[type]\r\n";
			if ($type != 'text' && $type != 'message') {
				$file['text'] = base64_encode($file['text']);
				$content .= "Content-Transfer-Encoding: base64\r\n";
			}
			$content .= "\r\n";
			$content .= $file['text'];
			$content .= "\r\n";
		}
		$content .= "--$boundary--\r\n";
	}

	//$headers['Content-Length'] = strlen($content);
	$headers['Mime-Version'] = '1.0';
	$headers['Date'] = date('r', $headers['Date']);

	$out = "";
	foreach ($headers as $id => $val) {
		$val = str_replace("\n", "\n\t", $val);
		$out .= "$id: $val\r\n";
	}
	$out .= "\r\n";
	$out .= $content;

	return $out;
}

function shoveitup($msg, $id) {
	$s = fsockopen("tcp://localhost:1119");
	fgets($s);
	fputs($s, "MODE STREAM\r\n");
	fgets($s);
	fputs($s, "TAKETHIS $id\r\n");
	fputs($s, $msg);
	fputs($s, "\r\n.\r\n");
	fgets($s);
	fclose($s);
}

$time = time();

echo "\n@@@@ Thread:\n";
echo $m0 = gennntp(["From" => "czaks <marcin@6irc.net>", "Message-Id" => "<1234.0000.".$time."@example.vichan.net>", "Newsgroups" => "overchan.test", "Date" => time(), "Subject" => "None"],
[['type' => 'text/plain', 'text' => "THIS IS A NEW TEST THREAD"]]);

echo "\n@@@@ Single msg:\n";
echo $m1 = gennntp(["From" => "czaks <marcin@6irc.net>", "Message-Id" => "<1234.1234.".$time."@example.vichan.net>", "Newsgroups" => "overchan.test", "Date" => time(), "Subject" => "None", "References" => "<1234.0000.".$time."@example.vichan.net>"],
[['type' => 'text/plain', 'text' => "hello world, with no image :("]]);

echo "\n@@@@ Single msg and pseudoimage:\n";
echo $m2 = gennntp(["From" => "czaks <marcin@6irc.net>", "Message-Id" => "<1234.2137.".$time."@example.vichan.net>", "Newsgroups" => "overchan.test", "Date" => time(), "Subject" => "None", "References" => "<1234.0000.".$time."@example.vichan.net>"],
[['type' => 'text/plain', 'text' => "hello world, now with an image!"],
 ['type' => 'image/gif', 'text' => base64_decode("R0lGODlhAQABAIAAAAUEBAAAACwAAAAAAQABAAACAkQBADs="), 'name' => "urgif.gif"]]);

echo "\n@@@@ Single msg and two pseudoimages:\n";
echo $m3 = gennntp(["From" => "czaks <marcin@6irc.net>", "Message-Id" => "<1234.1488.".$time."@example.vichan.net>", "Newsgroups" => "overchan.test", "Date" => time(), "Subject" => "None", "References" => "<1234.0000.".$time."@example.vichan.net>"],
[['type' => 'text/plain', 'text' => "hello world, now WITH TWO IMAGES!!!"],
 ['type' => 'image/gif', 'text' => base64_decode("R0lGODlhAQABAIAAAAUEBAAAACwAAAAAAQABAAACAkQBADs="), 'name' => "urgif.gif"],
 ['type' => 'image/gif', 'text' => base64_decode("R0lGODlhAQABAIAAAAUEBAAAACwAAAAAAQABAAACAkQBADs="), 'name' => "urgif2.gif"]]);

shoveitup($m0, "<1234.0000.".$time."@example.vichan.net>");
sleep(1);
shoveitup($m1, "<1234.1234.".$time."@example.vichan.net>");
sleep(1);
shoveitup($m2, "<1234.2137.".$time."@example.vichan.net>");
shoveitup($m3, "<1234.2131.".$time."@example.vichan.net>");
