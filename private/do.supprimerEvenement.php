<?php

session_start();
require_once __DIR__ . "/../api.php";

if (!isset($_SESSION['token'])) {
    header('Location: ../connection.php');
    exit();
}

$headers = array(api_bearer_header());
api_request(API_URL . '/api/v1/evenements?id=' . urlencode($_POST['id']), 'DELETE', $headers);

header('Location: ../bo/admin_evenements.php?limit=10&offset=0');
exit();
