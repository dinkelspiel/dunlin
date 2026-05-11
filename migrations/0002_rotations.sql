CREATE TABLE
    IF NOT EXISTS `image_rotations` (
        `id` INT PRIMARY KEY AUTO_INCREMENT,
        `team_project_id` INT NOT NULL,
        `file_path` VARCHAR(512) NOT NULL,
        `rotation_degrees` INT NOT NULL DEFAULT 0,
        `updated_at` TIMESTAMP,
        `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
        UNIQUE KEY `image_rotations_project_file` (`team_project_id`, `file_path`)
    );

INSERT INTO
    `migrations` (`name`)
SELECT
    "0002_rot"
WHERE
    NOT EXISTS (
        SELECT
            1
        FROM
            `migrations`
        WHERE
            `name` = "0002_rot"
    );