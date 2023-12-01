CREATE DATABASE order_service;
CREATE DATABASE payment_service;

CREATE USER 'payment_service'@'%' IDENTIFIED BY 'verysecretpass';
CREATE USER 'order_service'@'%' IDENTIFIED BY 'verysecretpass';

GRANT ALL PRIVILEGES ON order_service.* TO 'order_service'@'%';
GRANT ALL PRIVILEGES ON payment_service.* TO 'payment_service'@'%';
