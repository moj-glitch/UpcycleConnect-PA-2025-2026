<?php

session_start();
require_once __DIR__ . "/../api.php";

if (!isset($_SESSION['token'])) {
    header('Location: ../connection.php');
    exit();
}

$response = api_request(API_URL . '/api/v1/annonces?id=' . urlencode($_POST['id']), 'DELETE', array(api_bearer_header()));

header('Location: ../bo/admin_moderation_annonces.php?limit=10&offset=0');
exit();
