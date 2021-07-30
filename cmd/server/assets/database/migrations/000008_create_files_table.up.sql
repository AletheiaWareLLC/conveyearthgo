CREATE TABLE tbl_files (
    id INT AUTO_INCREMENT PRIMARY KEY,
    message INT NULL,
    hash TEXT(86),
    mime VARCHAR(255),
    created_unix INT UNSIGNED NOT NULL,
    FOREIGN KEY (message) REFERENCES tbl_messages(id)
);