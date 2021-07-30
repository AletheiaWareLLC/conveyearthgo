CREATE TABLE tbl_purchases (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user INT NOT NULL,
    stripe_session VARCHAR(255) NOT NULL,
    stripe_customer VARCHAR(255) NOT NULL,
    stripe_payment_intent VARCHAR(255) NOT NULL,
    stripe_currency TEXT(3) NOT NULL,
    stripe_amount INT NOT NULL,
    bundle_size INT NOT NULL,
    created_unix INT UNSIGNED NOT NULL,
    FOREIGN KEY (user) REFERENCES tbl_users(id)
);