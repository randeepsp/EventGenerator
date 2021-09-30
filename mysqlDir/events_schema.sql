DROP DATABASE IF EXISTS eventsdb;
create database eventsdb;
use eventsdb;

DROP TABLE IF EXISTS events;

CREATE TABLE events (
    dbid INT AUTO_INCREMENT PRIMARY KEY,
    id VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    dept VARCHAR(255),
    empid INT,
    etime DATE
);
