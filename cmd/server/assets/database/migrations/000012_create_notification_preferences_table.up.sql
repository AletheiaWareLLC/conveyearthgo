CREATE TABLE tbl_notification_preferences (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user INT NOT NULL,
    responses BOOL DEFAULT TRUE,
    mentions BOOL DEFAULT TRUE,
    digests BOOL DEFAULT TRUE,
    FOREIGN KEY (user) REFERENCES tbl_users(id)
);