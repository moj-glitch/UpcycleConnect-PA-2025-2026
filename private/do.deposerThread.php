<?php

session_start();
require_once __DIR__ . "/../api.php";

if (!isset($_SESSION['token'])) {
    header('Location: ../connection.php');
    exit();
}

$body = http_build_query(array(
    'categorie_thread' => $_POST['categorie_thread'],
    'titre' => $_POST['titre'],
    'message' => $_POST['message'],
    'resolu' => '0'
));

$headers = array(api_bearer_header(), 'Content-Type: application/x-www-form-urlencoded');
$response = api_request(API_URL . '/api/v1/threads', 'PUT', $headers, $body);

if ($response['status'] != 201) {
    header('Location: ../gp/forum_new.php?error=1');
    exit();
}

header('Location: ../gp/forums.php?limit=10&offset=0');
exit();
