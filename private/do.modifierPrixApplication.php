<?php

session_start();
require_once __DIR__ . "/../api.php";

if (!isset($_SESSION['token'])) {
    header('Location: ../connection.php');
    exit();
}

$body = http_build_query(array(
    'id' => $_POST['id'],
    'prix' => $_POST['prix']
));

$headers = array(api_bearer_header(), 'Content-Type: application/x-www-form-urlencoded');
api_request(API_URL . '/api/v1/applications', 'PATCH', $headers, $body);

header('Location: ../bo/admin_applications.php');
exit();
