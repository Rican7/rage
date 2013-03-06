<?php

// Require our Composer-generated auto-loader
require_once( __DIR__ . '/vendor/autoload.php' );

$app = new \Paulus\Paulus();

\Paulus\Router::any( '/?', function( $request, $response, $service ) {
	return $response->data = 'works!';
});

$app->run();
