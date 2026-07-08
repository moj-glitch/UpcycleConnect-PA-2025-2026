<?php

session_start();
require_once __DIR__ . "/../api.php";

if (!isset($_SESSION['token'])) {
    header('Location: ../connection.php');
    exit();
}

$response = api_request(API_URL . '/api/v1/applications', 'GET', array(api_bearer_header()));
$applications = json_decode($response['body'], true);

$application = null;
if (!empty($applications)) {
    foreach ($applications as $candidate) {
        if ($candidate['application_id'] == $_POST['application_id']) {
            $application = $candidate;
            break;
        }
    }
}

if ($application === null || $application['achetee']) {
    header('Location: ../pro/applications.php?error=1');
    exit();
}

$session = stripe_request('/v1/checkout/sessions', array(
    'mode' => 'payment',
    'line_items[0][price_data][currency]' => 'eur',
    'line_items[0][price_data][product_data][name]' => $application['nom'],
    'line_items[0][price_data][unit_amount]' => intval(round($application['prix'] * 100)),
    'line_items[0][quantity]' => 1,
    'success_url' => APP_BASE_URL . '/private/do.finaliserApplication.php?application_id=' . urlencode($application['application_id']) . '&session_id={CHECKOUT_SESSION_ID}',
    'cancel_url' => APP_BASE_URL . '/pro/applications.php?error=1'
));

if ($session['status'] != 200 || empty($session['body']['url'])) {
    header('Location: ../pro/applications.php?error=1');
    exit();
}

$_SESSION['stripe_pending'] = array(
    'session_id' => $session['body']['id'],
    'application_id' => $application['application_id']
);

header('Location: ' . $session['body']['url']);
exit();
