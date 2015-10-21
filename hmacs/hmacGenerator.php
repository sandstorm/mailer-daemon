<?php

function generateCsvSampleFile($algorithm, $encryptionKey, $samples) {
    $hmacs = array();
    foreach ($samples as $index => $sample) {
        $hmacs[$index] = hash_hmac($algorithm, $sample, $encryptionKey);
    }

    $filename = "hmacSamples.$algorithm.csv";
    echo "generating $filename\n";
    file_put_contents($filename, asCsvRow($samples, $encryptionKey, $hmacs));
}

function asCsvRow($samples, $encryptionKey, $hmacs) {
    $result = "# hmac\tencryptionKey\tsample\n";
    foreach ($hmacs as $index => $hmac) {
        $sample = $samples[$index];
        $result .= "\"".csvEscape($hmac)."\"\t";
        $result .= "\"".csvEscape($encryptionKey)."\"\t";
        $result .= "\"".csvEscape($sample)."\"\n";
    }
    return $result;
}

function csvEscape($value) {
    return str_replace("\"", "\"\"", $value);
}

$encryptionKey = "my-super-secret-private-key";
$samples = array(
    "Hello World",
    "",
    '{"email": "stefm@example.com", "firstName": "Stefan", "lastName": "Martin", "language": "en"}',
    '{"email": "foo@bar.de", "firstName": "Egon", "lastName": "Björk", "additional": "Test", "language": "de"}'
);

generateCsvSampleFile('sha1', $encryptionKey, $samples);


