CREATE TABLE tbl_stripe_account (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user INT NOT NULL,
    identity VARCHAR(255) NOT NULL,
    created_unix INT UNSIGNED NOT NULL,
    FOREIGN KEY (user) REFERENCES tbl_users(id)
);