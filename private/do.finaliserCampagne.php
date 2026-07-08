<?php

session_start();
require_once __DIR__ . "/../api.php";

if (!isset($_SESSION['token']) || !isset($_SESSION['stripe_campagne'])) {
    header('Location: ../connection.php');
    exit();
}

$pending = $_SESSION['stripe_campagne'];
unset($_SESSION['stripe_campagne']);

if ($_GET['session_id'] != $pending['session_id']) {
    header('Location: ../pro/projets.php?limit=10&offset=0&error=1');
    exit();
}

$session = stripe_get('/v1/checkout/sessions/' . urlencode($pending['session_id']));

if ($session['status'] != 200 || $session['body']['payment_status'] != 'paid') {
    header('Location: ../pro/projets.php?limit=10&offset=0&error=1');
    exit();
}

header('Location: ../pro/projets.php?limit=10&offset=0&campagne=1');
exit();
