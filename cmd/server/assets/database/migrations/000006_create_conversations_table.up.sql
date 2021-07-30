CREATE TABLE tbl_conversations (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user INT NOT NULL,
    topic VARCHAR(100) NOT NULL,
    created_unix INT UNSIGNED NOT NULL,
    FOREIGN KEY (user) REFERENCES tbl_users(id)
);