ALTER TABLE tbl_users
ADD COLUMN deleted_at INT UNSIGNED DEFAULT 0;

ALTER TABLE tbl_conversations
ADD COLUMN deleted_at INT UNSIGNED DEFAULT 0;

ALTER TABLE tbl_messages
ADD COLUMN deleted_at INT UNSIGNED DEFAULT 0;

ALTER TABLE tbl_files
ADD COLUMN deleted_at INT UNSIGNED DEFAULT 0;

ALTER TABLE tbl_yields
ADD COLUMN deleted_at INT UNSIGNED DEFAULT 0;

ALTER TABLE tbl_charges
ADD COLUMN deleted_at INT UNSIGNED DEFAULT 0;

ALTER TABLE tbl_purchases
ADD COLUMN deleted_at INT UNSIGNED DEFAULT 0;

ALTER TABLE tbl_awards
ADD COLUMN deleted_at INT UNSIGNED DEFAULT 0;

ALTER TABLE tbl_stripe_account
ADD COLUMN deleted_at INT UNSIGNED DEFAULT 0;

ALTER TABLE tbl_gifts
ADD COLUMN deleted_at INT UNSIGNED DEFAULT 0;
