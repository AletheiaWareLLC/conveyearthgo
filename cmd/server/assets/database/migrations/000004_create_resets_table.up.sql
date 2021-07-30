CREATE TABLE tbl_resets (
    id INT AUTO_INCREMENT PRIMARY KEY,
    token VARCHAR(100) NOT NULL UNIQUE KEY,
    error VARCHAR(100),
    username VARCHAR(100),
    created_unix INT UNSIGNED NOT NULL,
    user INT NULL,
    FOREIGN KEY (user) REFERENCES tbl_users(id)
);