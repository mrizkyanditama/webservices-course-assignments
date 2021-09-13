<?php

ini_set('memory_limit', '32M'); 

require_once __DIR__ . '/vendor/autoload.php';
use PhpAmqpLib\Connection\AMQPStreamConnection;
use PhpAmqpLib\Message\AMQPMessage;

$connection = new AMQPStreamConnection('127.0.0.1', 5672, guest, guest);

$channel = $connection->channel();

register_shutdown_function(
        function () use ($channel, $connection) {
          $channel->close();
          $connection->close();
        }
);

$exchange = "exchange_ping";
$channel->exchange_declare($exchange, 'fanout', false, false, false);


while(true) {
  $ts = date("Y-m-d H:i:s");
  $jam = date("H:i:s");
  $data = $ts;
  $msg = new AMQPMessage($data);
  $channel->basic_publish($msg, $exchange, "");
  echo " [x] Sent ", $data, "\n";
  sleep(1);
}

$channel->close();
$connection->close();


