CREATE TABLE tbl_gifts (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user INT NULL NOT NULL,
    conversation INT NOT NULL,
    message INT NULL,
    amount INT NOT NULL,
    created_unix INT UNSIGNED NOT NULL,
    FOREIGN KEY (user) REFERENCES tbl_users(id),
    FOREIGN KEY (conversation) REFERENCES tbl_conversations(id),
    FOREIGN KEY (message) REFERENCES tbl_messages(id)
);