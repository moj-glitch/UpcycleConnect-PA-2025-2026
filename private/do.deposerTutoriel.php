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
    'titre' => $_POST['titre'],
    'article' => $_POST['article'],
    'categorie' => $_POST['categorie']
);

foreach ($fields as $name => $value) {
    $parts .= "--$boundary\r\n";
    $parts .= "Content-Disposition: form-data; name=\"$name\"\r\n\r\n";
    $parts .= $value . "\r\n";
}

if (isset($_FILES['video']) && $_FILES['video']['error'] === UPLOAD_ERR_OK) {
    $fileContent = file_get_contents($_FILES['video']['tmp_name']);
    $parts .= "--$boundary\r\n";
    $parts .= "Content-Disposition: form-data; name=\"video\"; filename=\"" . basename($_FILES['video']['name']) . "\"\r\n";
    $parts .= "Content-Type: application/octet-stream\r\n\r\n";
    $parts .= $fileContent . "\r\n";
}

$parts .= "--$boundary--\r\n";

$headers = array(
    api_bearer_header(),
    "Content-Type: multipart/form-data; boundary=$boundary"
);

api_request(API_URL . '/api/v1/tutoriels', 'PUT', $headers, $parts);

header('Location: ../bo/admin_tutoriels.php');
exit();
