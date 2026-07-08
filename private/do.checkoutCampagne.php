<?php

session_start();
require_once __DIR__ . "/../api.php";

if (!isset($_SESSION['token'])) {
    header('Location: ../connection.php');
    exit();
}

$montant = intval($_POST['montant']);
if ($montant < 100 || $montant > 500) {
    header('Location: ../pro/projets.php?limit=10&offset=0&error=1');
    exit();
}

$session = stripe_request('/v1/checkout/sessions', array(
    'mode' => 'payment',
    'line_items[0][price_data][currency]' => 'eur',
    'line_items[0][price_data][product_data][name]' => 'Campagne: ' . $_POST['titre'],
    'line_items[0][price_data][unit_amount]' => $montant * 100,
    'line_items[0][quantity]' => 1,
    'success_url' => APP_BASE_URL . '/private/do.finaliserCampagne.php?session_id={CHECKOUT_SESSION_ID}',
    'cancel_url' => APP_BASE_URL . '/pro/projets.php?limit=10&offset=0&error=1'
));

if ($session['status'] != 200 || empty($session['body']['url'])) {
    header('Location: ../pro/projets.php?limit=10&offset=0&error=1');
    exit();
}

$_SESSION['stripe_campagne'] = array('session_id' => $session['body']['id']);

header('Location: ' . $session['body']['url']);
exit();
