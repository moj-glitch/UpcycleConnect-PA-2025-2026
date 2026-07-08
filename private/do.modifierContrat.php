<?php

session_start();
require_once __DIR__ . "/../api.php";

if (!isset($_SESSION['token'])) {
    header('Location: ../connection.php');
    exit();
}

$fields = array(
    'description' => $_POST['description'],
    'depense' => $_POST['depense'],
    'gain' => $_POST['gain']
);
if (!empty($_POST['date_fin'])) {
    $fields['date_fin'] = $_POST['date_fin'];
}

$parts = array();
foreach ($fields as $key => $value) {
    $parts[] = urlencode($key) . '=' . urlencode($value);
}
if (!empty($_POST['tiers'])) {
    foreach ($_POST['tiers'] as $tiersId) {
        $parts[] = 'tiers=' . urlencode($tiersId);
    }
}
$body = implode('&', $parts);

$headers = array(api_bearer_header(), 'Content-Type: application/x-www-form-urlencoded');
api_request(API_URL . '/api/v1/contrats?id=' . urlencode($_GET['id']), 'PATCH', $headers, $body);

header('Location: ../pro/contrat.php?id=' . urlencode($_GET['id']));
exit();
