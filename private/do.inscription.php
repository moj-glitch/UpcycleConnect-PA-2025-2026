<?php

require_once __DIR__ . "/../api.php";

$fields = array(
    'client_email' => $_POST['client_email'],
    'client_secret' => $_POST['client_secret'],
    'client_confirm' => $_POST['client_confirm'],
    'client_nom' => $_POST['client_nom'],
    'client_prenom' => $_POST['client_prenom'],
    'client_telephone' => $_POST['client_telephone'],
    'client_adresse' => $_POST['client_adresse'],
    'client_code_postal' => $_POST['client_code_postal'],
    'client_ville' => $_POST['client_ville'],
    'client_siret' => $_POST['client_siret'],
    'account_type' => isset($_POST['account_type']) ? $_POST['account_type'] : 'freemium'
);

if ($fields['account_type'] == 'entreprise') {
    $fields['forfait'] = isset($_POST['forfait']) ? $_POST['forfait'] : 'gratuit';
    $fields['entreprise_nom'] = $_POST['entreprise_nom'];
    $fields['entreprise_adresse'] = $_POST['entreprise_adresse'];
    $fields['entreprise_code_postal'] = $_POST['entreprise_code_postal'];
    $fields['entreprise_ville'] = $_POST['entreprise_ville'];
    $fields['entreprise_siret'] = $_POST['entreprise_siret'];
}

$body = http_build_query($fields);

$headers = array('Content-Type: application/x-www-form-urlencoded');
$response = api_request(OAUTH_URL . '/oauth/v3/inscription', 'POST', $headers, $body);

if ($response['status'] != 201) {
    header('Location: ../inscription.php?error=1');
    exit();
}

header('Location: ../connection.php');
exit();
