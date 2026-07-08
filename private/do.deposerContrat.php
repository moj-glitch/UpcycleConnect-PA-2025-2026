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
    'categorie_contrat_id' => $_POST['categorie_contrat_id'],
    'description' => $_POST['description'],
    'depense' => $_POST['depense'],
    'gain' => $_POST['gain']
);
if (!empty($_POST['date_fin'])) {
    $fields['date_fin'] = $_POST['date_fin'];
}

foreach ($fields as $name => $value) {
    $parts .= "--$boundary\r\n";
    $parts .= "Content-Disposition: form-data; name=\"$name\"\r\n\r\n";
    $parts .= $value . "\r\n";
}

if (isset($_FILES['pdf']) && $_FILES['pdf']['error'] === UPLOAD_ERR_OK) {
    $fileContent = file_get_contents($_FILES['pdf']['tmp_name']);
    $parts .= "--$boundary\r\n";
    $parts .= "Content-Disposition: form-data; name=\"pdf\"; filename=\"" . basename($_FILES['pdf']['name']) . "\"\r\n";
    $parts .= "Content-Type: application/pdf\r\n\r\n";
    $parts .= $fileContent . "\r\n";
}

if (!empty($_POST['tiers'])) {
    foreach ($_POST['tiers'] as $tiersId) {
        $parts .= "--$boundary\r\n";
        $parts .= "Content-Disposition: form-data; name=\"tiers\"\r\n\r\n";
        $parts .= $tiersId . "\r\n";
    }
}

$parts .= "--$boundary--\r\n";

$headers = array(
    api_bearer_header(),
    "Content-Type: multipart/form-data; boundary=$boundary"
);

$response = api_request(API_URL . '/api/v1/contrats', 'PUT', $headers, $parts);

if ($response['status'] != 201) {
    header('Location: ../pro/contrat_new.php?error=1');
    exit();
}

header('Location: ../pro/contrats.php?limit=10&offset=0');
exit();
