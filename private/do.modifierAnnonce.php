<?php

session_start();
require_once __DIR__ . "/../api.php";

if (!isset($_SESSION['token'])) {
    header('Location: ../connection.php');
    exit();
}

$body = http_build_query(array(
    'id' => $_POST['id'],
    'categorie' => $_POST['categorie'],
    'titre' => $_POST['titre'],
    'prix' => $_POST['prix'],
    'description' => $_POST['description'],
    'etat' => $_POST['etat'],
    'taxe' => $_POST['taxe'],
    'image' => $_POST['image']
));

$headers = array(api_bearer_header(), 'Content-Type: application/x-www-form-urlencoded');
$response = api_request(API_URL . '/api/v1/annonces', 'PATCH', $headers, $body);

if ($response['status'] != 204) {
    header('Location: ../gp/annonce_edit.php?id=' . urlencode($_POST['id']) . '&error=1');
    exit();
}

header('Location: ../gp/annonce.php?id=' . urlencode($_POST['id']));
exit();
