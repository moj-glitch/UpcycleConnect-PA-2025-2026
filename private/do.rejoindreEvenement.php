<?php

session_start();
require_once __DIR__ . "/../api.php";

if (!isset($_SESSION['token'])) {
    header('Location: ../connection.php');
    exit();
}

$body = http_build_query(array('evenement_id' => $_POST['evenement_id']));
$headers = array(api_bearer_header(), 'Content-Type: application/x-www-form-urlencoded');
api_request(API_URL . '/api/v1/planning', 'PUT', $headers, $body);

header('Location: ../gp/planning.php?limit=10&offset=0');
exit();
