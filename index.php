<?php

// Require our Composer-generated auto-loader
require_once( __DIR__ . '/vendor/autoload.php' );

use \Paulus\Paulus;
use \Paulus\Router;
use \Predis\Client;

$app = new Paulus();

// Respond to a base-call
Router::any( '/?', function( $request, $response, $service ) {
	// Create a new Predis Client (connect)
	$redis = new Client( array(
		'host' => '127.0.0.1',
		'port' => 6379
	));

	// Increment our number of given "fucks"
	$fucks_given = $redis->incr('fucks');
	
	// Set the response data as an object
	$response->data = (object) array(
		'fucks_given' => $fucks_given,
	);
});

$app->run();
