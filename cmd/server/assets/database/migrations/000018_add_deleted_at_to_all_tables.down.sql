ALTER TABLE tbl_users
DROP COLUMN deleted_at;

ALTER TABLE tbl_conversations
DROP COLUMN deleted_at;

ALTER TABLE tbl_messages
DROP COLUMN deleted_at;

ALTER TABLE tbl_files
DROP COLUMN deleted_at;

ALTER TABLE tbl_yields
DROP COLUMN deleted_at;

ALTER TABLE tbl_charges
DROP COLUMN deleted_at;

ALTER TABLE tbl_purchases
DROP COLUMN deleted_at;

ALTER TABLE tbl_awards
DROP COLUMN deleted_at;

ALTER TABLE tbl_stripe_account
DROP COLUMN deleted_at;

ALTER TABLE tbl_gifts
DROP COLUMN deleted_at;
