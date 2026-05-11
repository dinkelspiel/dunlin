ALTER TABLE `cached_images` MODIFY `directory` VARCHAR(512) NOT NULL,
MODIFY `file` VARCHAR(512) NOT NULL,
ADD COLUMN IF NOT EXISTS `rotation_degrees` INT NOT NULL DEFAULT 0;

INSERT INTO
    `migrations` (`name`)
SELECT
    "0003_cache"
WHERE
    NOT EXISTS (
        SELECT
            1
        FROM
            `migrations`
        WHERE
            `name` = "0003_cache"
    );