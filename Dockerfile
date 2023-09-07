FROM php:8.0-fpm

RUN docker-php-ext-install -j "$(nproc)" opcache

RUN mv "$PHP_INI_DIR/php.ini-production" "$PHP_INI_DIR/php.ini"
COPY .deploy/php-fpm.conf /usr/local/etc/php-fpm.d/zz-docker.conf

RUN apt-get update -y \
    && apt-get install -y nginx

COPY .deploy/nginx.conf /etc/nginx/sites-enabled/default

COPY --chown=www-data:www-data . /var/www/app
WORKDIR /var/www/app

EXPOSE 80

COPY --chmod=755 .deploy/exec /etc/exec
CMD ["/etc/exec"]
