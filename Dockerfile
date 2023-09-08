FROM php:8.0-fpm

RUN docker-php-ext-install -j "$(nproc)" opcache
RUN mkdir /tmp/php-fpm

RUN mv "$PHP_INI_DIR/php.ini-production" "$PHP_INI_DIR/php.ini"
COPY .deploy/php-fpm.conf /usr/local/etc/php-fpm.d/zz-docker.conf
COPY .deploy/php.cloudrun.ini /usr/local/etc/php/conf.d/zz-php.cloudrun.ini.disabled

RUN apt-get update -y \
    && apt-get install -y inotify-tools nginx

COPY .deploy/nginx.conf /etc/nginx/sites-enabled/default

COPY --chown=www-data:www-data . /var/www/app

# We should be able to use `COPY --chmod=755 ...`, but that requires "Docker Buildkit", which Google Cloud Build doesn't natively support
COPY .deploy/exec /etc/exec
RUN chmod 755 /etc/exec

WORKDIR /var/www/app
EXPOSE 80

CMD ["/etc/exec"]
