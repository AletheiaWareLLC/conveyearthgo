CREATE TABLE tbl_signups (
    id INT AUTO_INCREMENT PRIMARY KEY,
    token VARCHAR(100) NOT NULL UNIQUE KEY,
    error VARCHAR(100),
    email VARCHAR(320),
    username VARCHAR(100),
    challenge CHAR(8),
    created_unix INT UNSIGNED NOT NULL,
    user INT NULL,
    FOREIGN KEY (user) REFERENCES tbl_users(id)
);