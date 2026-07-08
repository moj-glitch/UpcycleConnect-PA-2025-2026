<?php

session_start();
require_once __DIR__ . "/../api.php";

if (!isset($_SESSION['token'])) {
    header('Location: ../connection.php');
    exit();
}

$fields = array(
    'categorie' => $_POST['categorie'],
    'titre' => $_POST['titre'],
    'prix' => $_POST['prix'],
    'description' => $_POST['description'],
    'etat' => 'D',
    'taxe' => '0.10',
    'image' => $_POST['image']
);

$parts = array();
foreach ($fields as $key => $value) {
    $parts[] = urlencode($key) . '=' . urlencode($value);
}
if (!empty($_POST['materiaux'])) {
    foreach ($_POST['materiaux'] as $materieauId) {
        $parts[] = 'materiaux=' . urlencode($materieauId);
    }
}
$body = implode('&', $parts);

$headers = array(api_bearer_header(), 'Content-Type: application/x-www-form-urlencoded');
$response = api_request(API_URL . '/api/v1/annonces', 'PUT', $headers, $body);

if ($response['status'] != 201) {
    header('Location: ../gp/annonce_new.php?error=1');
    exit();
}

header('Location: ../gp/annonces.php?limit=10&offset=0');
exit();
