<?php

session_start();
require_once __DIR__ . "/../api.php";

if (!isset($_SESSION['token'])) {
    header('Location: ../connection.php');
    exit();
}

$body = http_build_query(array(
    'id' => $_POST['id'],
    'categorie_thread' => $_POST['categorie_thread'],
    'titre' => $_POST['titre'],
    'message' => $_POST['message'],
    'resolu' => '1'
));

$headers = array(api_bearer_header(), 'Content-Type: application/x-www-form-urlencoded');
api_request(API_URL . '/api/v1/threads', 'PATCH', $headers, $body);

header('Location: ../gp/forum.php?id=' . urlencode($_POST['id']));
exit();
