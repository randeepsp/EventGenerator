FROM mysql:5.7.15

ENV MYSQL_DATABASE=eventsdb \
    MYSQL_ROOT_PASSWORD=pass1

ADD events_schema.sql /docker-entrypoint-initdb.d

EXPOSE 3306
