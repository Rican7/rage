<?php

// Require our Composer-generated auto-loader
require_once( __DIR__ . '/vendor/autoload.php' );

use \Paulus\Paulus;
use \Paulus\Router;
use \Predis\Client;

$app = new Paulus();

// Respond to a base-call
Router::any( '/?', function( $request, $response, $service ) {
	// Grab our format param
	$format = $request->param( 'format', 'json' );

	// Create a new Predis Client (connect)
	$redis = new Client( array(
		'host' => '127.0.0.1',
		'port' => 6379
	));

	// Increment our number of given "fucks"
	$fucks_given = $redis->incr('app:rage fucks');
	
	// Depending on the format..
	switch ( $format ) {
		case 'json' :
			// Set the response data as an object
			$response->data = (object) array(
				'fucks_given' => $fucks_given,
			);
			break;

		case 'plain' :
			// Return as plain-text
			echo $fucks_given . ' fucks given';
			exit;
			break;

		default :
			// Return an error
			$service->app->abort( 400, 'INVALID_FORMAT', 'Invalid format type', array(
				'valid_formats' => array(
					'json',
					'plain'
				),
			));
			break;
	}

});

$app->run();
