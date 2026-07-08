<?php

session_start();
require_once __DIR__ . "/../api.php";

$identifier = $_POST['client_id'];
$secret = $_POST['client_secret'];

$authHeader = 'Authorization: Basic ' . base64_encode($identifier . ':' . $secret);
$tokenResponse = api_request(OAUTH_URL . '/oauth/v3/token', 'POST', array($authHeader));

if ($tokenResponse['status'] != 200) {
    header('Location: ../connection.php?error=1');
    exit();
}

$tokenData = json_decode($tokenResponse['body'], true);

$introspectHeaders = array(
    'Authorization: Bearer ' . $tokenData['access_token'],
    'Content-Type: application/x-www-form-urlencoded'
);
$introspectResponse = api_request(OAUTH_URL . '/oauth/v3/introspect', 'POST', $introspectHeaders, '');

if ($introspectResponse['status'] != 200) {
    header('Location: ../connection.php?error=1');
    exit();
}

$introspectData = json_decode($introspectResponse['body'], true);

if (!$introspectData['active']) {
    header('Location: ../connection.php?error=1');
    exit();
}

$_SESSION['token'] = array(
    'access_token' => $tokenData['access_token'],
    'client_id' => $introspectData['client_id'],
    'username' => $introspectData['username'],
    'scope' => $introspectData['scope']
);

header('Location: ../myapps.php');
exit();
