FROM php:8.2-fpm-alpine

RUN apk add --no-cache \
    mysql-client \
    libzip-dev \
    zip \
    unzip \
    git \
    $PHPIZE_DEPS \
    pcre-dev

RUN docker-php-ext-install pdo_mysql zip

RUN pecl install redis && docker-php-ext-enable redis

RUN apk del $PHPIZE_DEPS pcre-dev

COPY --from=composer:latest /usr/bin/composer /usr/bin/composer

WORKDIR /var/www/html

EXPOSE 9000
CMD ["php-fpm"]
