<?php

session_start();
require_once __DIR__ . "/../api.php";

if (!isset($_SESSION['token'])) {
    header('Location: ../connection.php');
    exit();
}

$fields = array(
    'thread_id' => $_POST['thread_id'],
    'message' => $_POST['message']
);
if (!empty($_POST['parent'])) {
    $fields['parent'] = $_POST['parent'];
}
$body = http_build_query($fields);

$headers = array(api_bearer_header(), 'Content-Type: application/x-www-form-urlencoded');
api_request(API_URL . '/api/v1/threads/messages', 'PUT', $headers, $body);

header('Location: ../gp/forum.php?id=' . urlencode($_POST['thread_id']));
exit();
