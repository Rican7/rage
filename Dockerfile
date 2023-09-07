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

# We should be able to use `COPY --chmod=755 ...`, but that requires "Docker Buildkit", which Google Cloud Build doesn't natively support
COPY .deploy/exec /etc/exec
RUN chmod 755 /etc/exec

CMD ["/etc/exec"]