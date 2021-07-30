CREATE TABLE tbl_charges (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user INT NOT NULL,
    conversation INT NULL,
    message INT NULL,
    amount INT,
    created_unix INT UNSIGNED NOT NULL,
    FOREIGN KEY (user) REFERENCES tbl_users(id),
    FOREIGN KEY (conversation) REFERENCES tbl_conversations(id),
    FOREIGN KEY (message) REFERENCES tbl_messages(id)
);