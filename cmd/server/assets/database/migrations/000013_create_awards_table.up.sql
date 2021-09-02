CREATE TABLE tbl_awards (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user INT NOT NULL,
    reason VARCHAR(255) NOT NULL,
    amount INT NOT NULL,
    created_unix INT UNSIGNED NOT NULL,
    FOREIGN KEY (user) REFERENCES tbl_users(id)
);