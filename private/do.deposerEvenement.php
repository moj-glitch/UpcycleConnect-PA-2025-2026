<?php

session_start();
require_once __DIR__ . "/../api.php";

if (!isset($_SESSION['token'])) {
    header('Location: ../connection.php');
    exit();
}

$body = http_build_query(array(
    'nom' => $_POST['nom'],
    'description' => $_POST['description'],
    'date' => $_POST['date'],
    'statut' => $_POST['statut'],
    'categorie' => $_POST['categorie']
));

$headers = array(api_bearer_header(), 'Content-Type: application/x-www-form-urlencoded');
api_request(API_URL . '/api/v1/evenements', 'PUT', $headers, $body);

header('Location: ../bo/admin_evenements.php?limit=10&offset=0');
exit();
