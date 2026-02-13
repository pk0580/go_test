FROM php:8.2-fpm-alpine

RUN apk add --no-cache \
    mysql-client \
    libzip-dev \
    zip \
    unzip \
    git \
    $PHPIZE_DEPS \
    pcre-dev \
    protobuf-dev \
    grpc-dev

RUN docker-php-ext-install pdo_mysql zip

RUN pecl install redis grpc && docker-php-ext-enable redis grpc

RUN apk del $PHPIZE_DEPS pcre-dev

COPY --from=composer:latest /usr/bin/composer /usr/bin/composer

WORKDIR /var/www/html

# Исправляем права при сборке (хотя при монтировании томов они могут перекрыться)
# Но лучше иметь это в Dockerfile для продакшена
RUN mkdir -p storage/framework/views storage/framework/cache storage/framework/sessions bootstrap/cache && \
    chown -R www-data:www-data storage bootstrap/cache && \
    chmod -R 775 storage bootstrap/cache

EXPOSE 9000
CMD ["php-fpm"]
