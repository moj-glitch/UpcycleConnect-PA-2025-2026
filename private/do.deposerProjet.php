<?php

session_start();
require_once __DIR__ . "/../api.php";

if (!isset($_SESSION['token'])) {
    header('Location: ../connection.php');
    exit();
}

$boundary = '----phpboundary' . bin2hex(random_bytes(16));
$parts = '';

$fields = array(
    'categorie_projet' => $_POST['categorie_projet'],
    'titre' => $_POST['titre'],
    'texte' => $_POST['texte']
);

foreach ($fields as $name => $value) {
    $parts .= "--$boundary\r\n";
    $parts .= "Content-Disposition: form-data; name=\"$name\"\r\n\r\n";
    $parts .= $value . "\r\n";
}

if (isset($_FILES['image']) && $_FILES['image']['error'] === UPLOAD_ERR_OK) {
    $fileContent = file_get_contents($_FILES['image']['tmp_name']);
    $parts .= "--$boundary\r\n";
    $parts .= "Content-Disposition: form-data; name=\"image\"; filename=\"" . basename($_FILES['image']['name']) . "\"\r\n";
    $parts .= "Content-Type: application/octet-stream\r\n\r\n";
    $parts .= $fileContent . "\r\n";
}

$parts .= "--$boundary--\r\n";

$headers = array(
    api_bearer_header(),
    "Content-Type: multipart/form-data; boundary=$boundary"
);

$response = api_request(API_URL . '/api/v1/projets', 'PUT', $headers, $parts);

if ($response['status'] != 201) {
    header('Location: ../pro/projet_new.php?error=1');
    exit();
}

header('Location: ../pro/projets.php?limit=10&offset=0');
exit();
