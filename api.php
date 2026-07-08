<?php

define('OAUTH_URL', getenv('OAUTH_URL') ?: 'http://localhost:8080');
define('API_URL', getenv('API_URL') ?: 'http://localhost:8081');
define('APP_BASE_URL', getenv('APP_BASE_URL') ?: 'http://localhost');
define('STRIPE_SECRET_KEY', getenv('STRIPE_SECRET_KEY') ?: '');

function stripe_is_configured() {
    return STRIPE_SECRET_KEY !== '' && strpos(STRIPE_SECRET_KEY, 'sk_test_') === 0;
}

function stripe_request($path, $fields) {
    if (!stripe_is_configured()) {
        return array('status' => 0, 'body' => null);
    }
    $options = array(
        'http' => array(
            'method' => 'POST',
            'header' => "Authorization: Bearer " . STRIPE_SECRET_KEY . "\r\nContent-Type: application/x-www-form-urlencoded",
            'content' => http_build_query($fields),
            'ignore_errors' => true
        )
    );
    $context = stream_context_create($options);
    $result = @file_get_contents('https://api.stripe.com' . $path, false, $context);
    $status = 0;
    if (isset($http_response_header)) {
        $parts = explode(' ', $http_response_header[0]);
        if (isset($parts[1])) {
            $status = intval($parts[1]);
        }
    }
    return array('status' => $status, 'body' => json_decode($result, true));
}

function stripe_get($path) {
    if (!stripe_is_configured()) {
        return array('status' => 0, 'body' => null);
    }
    $options = array(
        'http' => array(
            'method' => 'GET',
            'header' => "Authorization: Bearer " . STRIPE_SECRET_KEY,
            'ignore_errors' => true
        )
    );
    $context = stream_context_create($options);
    $result = @file_get_contents('https://api.stripe.com' . $path, false, $context);
    $status = 0;
    if (isset($http_response_header)) {
        $parts = explode(' ', $http_response_header[0]);
        if (isset($parts[1])) {
            $status = intval($parts[1]);
        }
    }
    return array('status' => $status, 'body' => json_decode($result, true));
}

function api_request($url, $method, $headers = array(), $body = null) {
    $options = array(
        'http' => array(
            'method' => $method,
            'header' => implode("\r\n", $headers),
            'content' => $body,
            'ignore_errors' => true
        )
    );
    $context = stream_context_create($options);
    $result = @file_get_contents($url, false, $context);
    $status = 0;
    if (isset($http_response_header)) {
        $parts = explode(' ', $http_response_header[0]);
        if (isset($parts[1])) {
            $status = intval($parts[1]);
        }
    }
    return array('status' => $status, 'body' => $result);
}

function api_bearer_header() {
    if (isset($_SESSION['token']['access_token'])) {
        return 'Authorization: Bearer ' . $_SESSION['token']['access_token'];
    }
    return '';
}
