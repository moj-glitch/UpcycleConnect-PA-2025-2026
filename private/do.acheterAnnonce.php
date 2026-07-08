<?php

session_start();
require_once __DIR__ . "/../api.php";

if (!isset($_SESSION['token'])) {
    header('Location: ../connection.php');
    exit();
}

$body = http_build_query(array('id' => $_POST['id']));
$headers = array(api_bearer_header(), 'Content-Type: application/x-www-form-urlencoded');
$response = api_request(API_URL . '/api/v1/annonces/achat', 'PATCH', $headers, $body);

if ($response['status'] != 204) {
    header('Location: ../gp/annonce.php?id=' . urlencode($_POST['id']) . '&error=1');
    exit();
}

header('Location: ../gp/annonce.php?id=' . urlencode($_POST['id']));
exit();
