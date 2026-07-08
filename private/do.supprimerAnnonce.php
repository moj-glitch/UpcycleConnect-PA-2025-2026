<?php

session_start();
require_once __DIR__ . "/../api.php";

if (!isset($_SESSION['token'])) {
    header('Location: ../connection.php');
    exit();
}

$headers = array(api_bearer_header());
$response = api_request(API_URL . '/api/v1/annonces?id=' . urlencode($_POST['id']), 'DELETE', $headers);

if ($response['status'] != 204) {
    header('Location: ../gp/annonce.php?id=' . urlencode($_POST['id']) . '&error=1');
    exit();
}

header('Location: ../gp/annonces.php?limit=10&offset=0');
exit();
