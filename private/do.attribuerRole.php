<?php

session_start();
require_once __DIR__ . "/../api.php";

if (!isset($_SESSION['token'])) {
    header('Location: ../connection.php');
    exit();
}

$body = http_build_query(array(
    'client_id' => $_POST['client_id'],
    'libelle' => $_POST['libelle']
));

$headers = array(api_bearer_header(), 'Content-Type: application/x-www-form-urlencoded');
api_request(API_URL . '/api/v1/admin/roles', 'PUT', $headers, $body);

header('Location: ../bo/admin_clients.php?limit=10&offset=0');
exit();
