CREATE TABLE tbl_messages (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user INT NULL NOT NULL,
    conversation INT NOT NULL,
    parent INT NULL,
    created_unix INT UNSIGNED NOT NULL,
    FOREIGN KEY (user) REFERENCES tbl_users(id),
    FOREIGN KEY (conversation) REFERENCES tbl_conversations(id),
    FOREIGN KEY (parent) REFERENCES tbl_messages(id)
);