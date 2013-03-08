<?php

// Require our Composer-generated auto-loader
require_once( __DIR__ . '/vendor/autoload.php' );

use \Paulus\Paulus;
use \Paulus\Router;

use \Predis\Client;

$app = new Paulus();

// Always do.
Router::any( function( $request, $response, $service ) {
	// Define available formats
	$available_formats = array( 'json', 'plain' );

	// Grab our format param
	$request->format = strtolower( $request->param( 'format', 'json' ) );

	// Handle our format
	if ( in_array( $request->format, $available_formats ) !== true ) {
		// Return an error
		$service->app->abort( 400, 'INVALID_FORMAT', 'Invalid format type', array(
			'valid_formats' => $available_formats,
		));
	}

	if ( $request->format === 'plain' ) {
		$response->header( 'Content-Type', 'text/plain' );
	}
});

// Respond to a base-call
Router::route( array( 'HEAD', 'GET', 'POST' ), '/?', function( $request, $response, $service ) {
	// Create a new Predis Client (connect)
	$redis = new Client( array(
		'host' => '127.0.0.1',
		'port' => 6379
	));

	// Increment our number of given "fucks"
	$fucks_given = $redis->incr('app:rage fucks');
	
	// Depending on the format..
	switch ( $request->format ) {
		case 'json' :
			// Set the response data as an object
			$response->data = (object) array(
				'fucks_given' => $fucks_given,
			);
			break;

		case 'plain' :
			// Return as plain-text
			echo $fucks_given . ' fucks given' . PHP_EOL;
			exit;
			break;
	}

});

// About
Router::get( '/about/?', function( $request, $response, $service ) {
	// Define our about response
	$about_response = array(
		'body' => array(
			'raw' => 'Wtf is this?! I\'m not exactly sure myself.
					To be honest, I used to work with @abackstrom,
					and he one day decided to {{tweet}} a ridiculous bash alias.
					Naturally, I added it to my own .bashrc file and loved it...
					Until one day, it stopped working.
					
					This is my recreation of that original script, which just
					so happened to allow me to test out my new {{paulus}} and have an excuse to try using a Redis DB
					for the first time. So... yea. Now you know.',
			'parsed' => null,
		),
		'templates' => array(
			'text' => array(
				'tweet' => 'tweet',
				'paulus' => 'PHP micro-framework (Paulus)',
			),
			'sources' => array(
				'tweet' => 'https://twitter.com/abackstrom/status/232898857837662208',
				'paulus' => 'https://github.com/Rican7/Paulus',
			),
		),
	);

	// Strip our tabs...
	$about_response[ 'body' ][ 'raw' ] = str_replace( "\t", '', $about_response[ 'body' ][ 'raw' ] );

	// Get rid of our awkward line-breaks
	$about_response[ 'body' ][ 'raw' ] = preg_replace( '/([^\n])[\n]/', '$1 ', $about_response[ 'body' ][ 'raw' ] );

	$about_response[ 'body' ][ 'parsed' ] = $about_response[ 'body' ][ 'raw' ];

	// Parse our body
	foreach ( $about_response[ 'templates' ][ 'text' ] as $key => $replacement ) {
		$about_response[ 'body' ][ 'parsed' ] = str_replace( '{{' . $key . '}}', $replacement, $about_response[ 'body' ][ 'parsed' ] );
	}

	// Depending on the format..
	switch ( $request->format ) {
		case 'json' :
			// Set the response data as an object
			$response->data = (object) $about_response;
			break;

		case 'plain' :
			// Return as plain-text
			echo $about_response[ 'body' ][ 'parsed' ] . PHP_EOL;
			exit;
			break;
	}

});

$app->run();
