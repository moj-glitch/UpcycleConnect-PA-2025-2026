<?php

session_start();
require_once __DIR__ . "/../api.php";

if (!isset($_SESSION['token']) || !isset($_SESSION['stripe_pending'])) {
    header('Location: ../connection.php');
    exit();
}

$pending = $_SESSION['stripe_pending'];
unset($_SESSION['stripe_pending']);

if ($_GET['session_id'] != $pending['session_id'] || $_GET['application_id'] != $pending['application_id']) {
    header('Location: ../pro/applications.php?error=1');
    exit();
}

$session = stripe_get('/v1/checkout/sessions/' . urlencode($pending['session_id']));

if ($session['status'] != 200 || $session['body']['payment_status'] != 'paid') {
    header('Location: ../pro/applications.php?error=1');
    exit();
}

$body = http_build_query(array('application_id' => $pending['application_id']));
$headers = array(api_bearer_header(), 'Content-Type: application/x-www-form-urlencoded');
$response = api_request(API_URL . '/api/v1/applications', 'PUT', $headers, $body);

if ($response['status'] != 201) {
    header('Location: ../pro/applications.php?error=1');
    exit();
}

header('Location: ../pro/applications.php?success=1');
exit();
