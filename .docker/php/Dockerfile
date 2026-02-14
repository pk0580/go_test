FROM php:8.4-fpm

# Установщик расширений
COPY --from=mlocati/php-extension-installer /usr/bin/install-php-extensions /usr/local/bin/

# Системные зависимости + базовые PHP extensions
RUN apt-get update && apt-get install -y \
    mariadb-client \
    zip \
    unzip \
    git \
    libzip-dev \
 && docker-php-ext-install pdo_mysql zip \
 && rm -rf /var/lib/apt/lists/*

# Redis + gRPC (сборка из исходников)
RUN install-php-extensions redis grpc

# Composer
COPY --from=composer:latest /usr/bin/composer /usr/bin/composer

WORKDIR /var/www/html

# Laravel permissions
RUN mkdir -p storage/framework/views \
             storage/framework/cache \
             storage/framework/sessions \
             bootstrap/cache \
 && chown -R www-data:www-data storage bootstrap/cache \
 && chmod -R 775 storage bootstrap/cache

EXPOSE 9000
CMD ["php-fpm"]
