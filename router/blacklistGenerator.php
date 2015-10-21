<?php

$blacklist = array(
    "foo@example.net",
    "BAR@example.net"
);

foreach($blacklist as $email) {
    echo "$email  \n";
}

echo "\n";
echo "     \t some other stuff but no email address at front\n";
echo "\t some other stuff but no email address at front\n";

$delimiters = array(
    ";",
    "\t",
    ",",
    "|"
);

foreach($delimiters as $index => $delimiter) {
    echo " mail$index@example.org  $delimiter"."some additional$delimiter"."information\n";
}

echo "christoph.daehne@sandstorm-media.de\n";