<?php

session_start();
require_once __DIR__ . "/../api.php";

if (!isset($_SESSION['token'])) {
    header('Location: ../connection.php');
    exit();
}

$headers = array(api_bearer_header());
api_request(API_URL . '/api/v1/admin/roles?client_id=' . urlencode($_POST['client_id']) . '&libelle=' . urlencode($_POST['libelle']), 'DELETE', $headers);

header('Location: ../bo/admin_clients.php?limit=10&offset=0');
exit();
