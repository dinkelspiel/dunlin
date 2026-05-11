CREATE TABLE
    IF NOT EXISTS `migrations` (
        `id` INT PRIMARY KEY AUTO_INCREMENT,
        `name` VARCHAR(16) NOT NULL
    );

INSERT INTO
    `migrations` (`name`)
VALUES
    ("0001_init");

CREATE TABLE
    IF NOT EXISTS `users` (
        `id` INT PRIMARY KEY AUTO_INCREMENT,
        `username` VARCHAR(32) NOT NULL UNIQUE,
        `email` VARCHAR(128) NOT NULL UNIQUE,
        `updated_at` TIMESTAMP,
        `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
    );

CREATE TABLE
    IF NOT EXISTS `user_auth_codes` (
        `id` INT PRIMARY KEY AUTO_INCREMENT,
        `user_id` INT NOT NULL,
        `code` INT NOT NULL,
        `used` BOOLEAN NOT NULL,
        `updated_at` TIMESTAMP,
        `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
    );

CREATE TABLE
    IF NOT EXISTS `user_sessions` (
        `id` INT PRIMARY KEY AUTO_INCREMENT,
        `user_id` INT NOT NULL,
        `token` UUID NOT NULL UNIQUE,
        `updated_at` TIMESTAMP,
        `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
    );

CREATE TABLE
    IF NOT EXISTS `teams` (
        `id` INT PRIMARY KEY AUTO_INCREMENT,
        `name` VARCHAR(64) NOT NULL,
        `slug` VARCHAR(64) NOT NULL,
        `owner_id` INT NOT NULL,
        `updated_at` TIMESTAMP,
        `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
    );

CREATE TABLE
    IF NOT EXISTS `team_users` (
        `id` INT PRIMARY KEY AUTO_INCREMENT,
        `user_id` INT NOT NULL,
        `team_id` INT NOT NULL,
        `updated_at` TIMESTAMP,
        `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
    );

CREATE TABLE
    IF NOT EXISTS `team_projects` (
        `id` INT PRIMARY KEY AUTO_INCREMENT,
        `name` VARCHAR(64) NOT NULL,
        `slug` VARCHAR(64) NOT NULL,
        `team_id` INT NOT NULL,
        `updated_at` TIMESTAMP,
        `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
    );

CREATE TABLE
    IF NOT EXISTS `cached_images` (
        `id` INT PRIMARY KEY AUTO_INCREMENT,
        `width` INT NOT NULL,
        `height` INT NOT NULL,
        `cache_file` VARCHAR(64) NOT NULL,
        `directory` VARCHAR(64) NOT NULL,
        `file` VARCHAR(64) NOT NULL,
        `size_bytes` INT NOT NULL,
        `team_project_id` INT NOT NULL,
        `updated_at` TIMESTAMP,
        `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
    );