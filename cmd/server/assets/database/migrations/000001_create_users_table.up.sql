CREATE TABLE tbl_users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(320) NOT NULL UNIQUE KEY,
    username VARCHAR(100) NOT NULL UNIQUE KEY,
    password BINARY(60) NOT NULL,
    created_unix INT UNSIGNED NOT NULL,
    verified BOOLEAN NOT NULL DEFAULT FALSE
);
