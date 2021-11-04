package database

import (
	"aletheiaware.com/authgo"
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	migratemysql "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"io/fs"
	"strings"
	"time"
)

/*
migrate -database 'mysql://tester:tester123@tcp(localhost:3306)/testdb' -path cmd/server/assets/database/migrations/ up
migrate -database 'mysql://tester:tester123@tcp(localhost:3306)/testdb' -path cmd/server/assets/database/migrations/ down

// Drop database
DROP DATABASE IF EXISTS testdb;

// Create database
CREATE DATABASE testdb;

// Select database
USE testdb;

// Show migrations
SELECT *
FROM schema_migrations;

// Show all users
SELECT *
FROM tbl_users;

// Show all sign ups
SELECT *
FROM tbl_signups;

// Show all sign ins
SELECT *
FROM tbl_signins;

// Show all resets
SELECT *
FROM tbl_resets;

// Show all recoveries
SELECT *
FROM tbl_recoveries;

// Show all conversations
SELECT *
FROM tbl_conversations;

// Show all messages
SELECT *
FROM tbl_messages;

// Show all files
SELECT *
FROM tbl_files;

// Show all charges
SELECT *
FROM tbl_charges;

// Show all yields
SELECT *
FROM tbl_yields;

// Show all purchases
SELECT *
FROM tbl_purchases;

// Show all awards
SELECT *
FROM tbl_awards;

// Show best content
SELECT tbl_conversations.id, tbl_conversations.user, tbl_users.username, tbl_conversations.topic, tbl_conversations.created_unix, tbl_charges.amount, IFNULL(yields.yield, 0)
FROM tbl_conversations
INNER JOIN tbl_users ON tbl_conversations.user=tbl_users.id
INNER JOIN tbl_messages ON tbl_conversations.id=tbl_messages.conversation AND tbl_messages.parent IS NULL
INNER JOIN tbl_charges ON tbl_messages.id=tbl_charges.message
LEFT JOIN (
	SELECT parent, SUM(IFNULL(amount,0)) AS yield
	FROM tbl_yields
	GROUP BY parent
) AS yields ON tbl_messages.id=yields.parent
ORDER BY yields.yield DESC
*/

func NewSql(dbname, username, password, host, port string, secure bool) (*Sql, error) {
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "3306"
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?tls=%t", username, password, host, port, dbname, secure)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &Sql{db}, nil
}

type Sql struct {
	*sql.DB
}

func (db *Sql) Migrator(migrations fs.FS) (*migrate.Migrate, error) {
	source, err := iofs.New(migrations, ".")
	if err != nil {
		return nil, err
	}
	driver, err := migratemysql.WithInstance(db.DB, &migratemysql.Config{})
	if err != nil {
		return nil, err
	}
	return migrate.NewWithInstance("iofs", source, "mysql", driver)
}

func (db *Sql) CreateUser(email, username string, password []byte, created time.Time) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO tbl_users
		SET email=?, username=?, password=?, created_unix=?`, email, username, password, created.Unix())
	if err != nil {
		if driverErr, ok := err.(*mysql.MySQLError); ok {
			switch driverErr.Number {
			case 1062: // ER_DUP_ENTRY
				msg := driverErr.Message
				if strings.HasSuffix(msg, "'email'") {
					return 0, authgo.ErrEmailAlreadyRegistered
				} else if strings.HasSuffix(msg, "'username'") {
					return 0, authgo.ErrUsernameAlreadyRegistered
				}
			}
		}
		return 0, err
	}
	return result.LastInsertId()
}

func (db *Sql) SelectUser(username string) (int64, string, []byte, time.Time, error) {
	row := db.QueryRow(`
		SELECT id, email, password, created_unix
		FROM tbl_users
		WHERE username=?`, username)

	var (
		id       int64
		email    string
		password []byte
		created  int64
	)
	if err := row.Scan(&id, &email, &password, &created); err != nil {
		if err == sql.ErrNoRows {
			err = authgo.ErrUsernameNotRegistered
		}
		return 0, "", nil, time.Time{}, err
	}
	return id, email, password, time.Unix(created, 0), nil
}

func (db *Sql) SelectUsernameByEmail(email string) (string, error) {
	row := db.QueryRow(`
		SELECT username
		FROM tbl_users
		WHERE email=?`, email)
	var (
		username string
	)
	if err := row.Scan(&username); err != nil {
		if err == sql.ErrNoRows {
			err = authgo.ErrEmailNotRegistered
		}
		return "", err
	}
	return username, nil
}

func (db *Sql) ChangePassword(username string, password []byte) (int64, error) {
	result, err := db.Exec(`
		UPDATE tbl_users
		SET password=?
		WHERE username=?`, password, username)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (db *Sql) IsEmailVerified(email string) (bool, error) {
	row := db.QueryRow(`
		SELECT verified
		FROM tbl_users
		WHERE email=?`, email)

	var verified bool
	if err := row.Scan(&verified); err != nil {
		if err == sql.ErrNoRows {
			err = authgo.ErrEmailNotRegistered
		}
		return false, err
	}
	return verified, nil
}

func (db *Sql) SetEmailVerified(email string, verified bool) (int64, error) {
	result, err := db.Exec(`
		UPDATE tbl_users 
		SET verified=? 
		WHERE email=?`, verified, email)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (db *Sql) CreateSignUpSession(token string, created time.Time) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO tbl_signups
		SET token=?, created_unix=?`, token, created.Unix())
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (db *Sql) SelectSignUpSession(token string) (string, string, string, string, time.Time, error) {
	row := db.QueryRow(`
		SELECT IFNULL(error, ''), IFNULL(email, ''), IFNULL(username, ''), IFNULL(challenge, ''), created_unix
		FROM tbl_signups
		WHERE token=?`, token)

	var (
		error     string
		email     string
		username  string
		challenge string
		created   int64
	)
	if err := row.Scan(&error, &email, &username, &challenge, &created); err != nil {
		return "", "", "", "", time.Time{}, err
	}
	return error, email, username, challenge, time.Unix(created, 0), nil
}

func (db *Sql) UpdateSignUpSessionError(token, error string) (int64, error) {
	result, err := db.Exec(`
		UPDATE tbl_signups
		SET error=?
		WHERE token=?`, error, token)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (db *Sql) UpdateSignUpSessionIdentity(token, email, username string) (int64, error) {
	result, err := db.Exec(`
		UPDATE tbl_signups
		SET email=?, username=?
		WHERE token=?`, email, username, token)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (db *Sql) UpdateSignUpSessionChallenge(token, challenge string) (int64, error) {
	result, err := db.Exec(`
		UPDATE tbl_signups
		SET challenge=?
		WHERE token=?`, challenge, token)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (db *Sql) CreateSignInSession(token, username string, authenticated bool, created time.Time) (int64, error) {
	result, err := db.Exec(`
			INSERT INTO tbl_signins
			SET token=?, username=?, authenticated=?, created_unix=?`, token, username, authenticated, created.Unix())
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (db *Sql) SelectSignInSession(token string) (string, string, time.Time, bool, error) {
	row := db.QueryRow(`
		SELECT IFNULL(error, ''), IFNULL(username, ''), created_unix, authenticated
		FROM tbl_signins
		WHERE token=?`, token)

	var (
		error         string
		username      string
		created       int64
		authenticated bool
	)
	if err := row.Scan(&error, &username, &created, &authenticated); err != nil {
		return "", "", time.Time{}, false, err
	}
	return error, username, time.Unix(created, 0), authenticated, nil
}

func (db *Sql) UpdateSignInSessionError(token, error string) (int64, error) {
	result, err := db.Exec(`
		UPDATE tbl_signins
		SET error=?
		WHERE token=?`, error, token)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (db *Sql) UpdateSignInSessionUsername(token, username string) (int64, error) {
	result, err := db.Exec(`
		UPDATE tbl_signins
		SET username=?
		WHERE token=?`, username, token)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (db *Sql) UpdateSignInSessionAuthenticated(token string, authenticated bool) (int64, error) {
	result, err := db.Exec(`
		UPDATE tbl_signins
		SET authenticated=?
		WHERE token=?`, authenticated, token)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (db *Sql) CreateAccountPasswordSession(token, username string, created time.Time) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO tbl_resets
		SET token=?, username=?, created_unix=?`, token, username, created.Unix())
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (db *Sql) SelectAccountPasswordSession(token string) (string, string, time.Time, error) {
	row := db.QueryRow(`
		SELECT IFNULL(error, ''), IFNULL(username, ''), created_unix
		FROM tbl_resets
		WHERE token=?`, token)

	var (
		error    string
		username string
		created  int64
	)
	if err := row.Scan(&error, &username, &created); err != nil {
		return "", "", time.Time{}, err
	}
	return error, username, time.Unix(created, 0), nil
}

func (db *Sql) UpdateAccountPasswordSessionError(token, error string) (int64, error) {
	result, err := db.Exec(`
		UPDATE tbl_resets
		SET error=?
		WHERE token=?`, error, token)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (db *Sql) CreateAccountRecoverySession(token string, created time.Time) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO tbl_recoveries
		SET token=?, created_unix=?`, token, created.Unix())
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (db *Sql) SelectAccountRecoverySession(token string) (string, string, string, string, time.Time, error) {
	row := db.QueryRow(`
		SELECT IFNULL(error, ''), IFNULL(email, ''), IFNULL(username, ''), IFNULL(challenge, ''), created_unix
		FROM tbl_recoveries
		WHERE token=?`, token)

	var (
		error     string
		email     string
		username  string
		challenge string
		created   int64
	)
	if err := row.Scan(&error, &email, &username, &challenge, &created); err != nil {
		return "", "", "", "", time.Time{}, err
	}
	return error, email, username, challenge, time.Unix(created, 0), nil
}

func (db *Sql) UpdateAccountRecoverySessionError(token, error string) (int64, error) {
	result, err := db.Exec(`
		UPDATE tbl_recoveries
		SET error=?
		WHERE token=?`, error, token)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (db *Sql) UpdateAccountRecoverySessionEmail(token, email string) (int64, error) {
	result, err := db.Exec(`
		UPDATE tbl_recoveries
		SET email=?
		WHERE token=?`, email, token)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (db *Sql) UpdateAccountRecoverySessionUsername(token, username string) (int64, error) {
	result, err := db.Exec(`
		UPDATE tbl_recoveries
		SET username=?
		WHERE token=?`, username, token)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (db *Sql) UpdateAccountRecoverySessionChallenge(token, challenge string) (int64, error) {
	result, err := db.Exec(`
		UPDATE tbl_recoveries
		SET challenge=?
		WHERE token=?`, challenge, token)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (db *Sql) CreateConversation(user int64, topic string, created time.Time) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO tbl_conversations
		SET user=?, topic=?, created_unix=?`, user, topic, created.Unix())
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (db *Sql) SelectConversation(id int64) (*authgo.Account, string, time.Time, error) {
	row := db.QueryRow(`
		SELECT tbl_conversations.user, tbl_users.username, tbl_users.email, tbl_conversations.topic, tbl_conversations.created_unix
		FROM tbl_conversations
		INNER JOIN tbl_users ON tbl_conversations.user=tbl_users.id
		WHERE tbl_conversations.id=?`, id)

	var (
		user     int64
		username string
		email    string
		topic    string
		created  int64
	)
	if err := row.Scan(&user, &username, &email, &topic, &created); err != nil {
		return nil, "", time.Time{}, err
	}
	return &authgo.Account{
		ID:       user,
		Username: username,
		Email:    email,
	}, topic, time.Unix(created, 0), nil
}

func (db *Sql) SelectBestConversations(callback func(int64, *authgo.Account, string, time.Time, int64, int64) error, since time.Time, limit int64) error {
	rows, err := db.Query(`
		SELECT tbl_conversations.id, tbl_conversations.user, tbl_users.username, tbl_users.email, tbl_conversations.topic, tbl_conversations.created_unix, tbl_charges.amount, IFNULL(yields.yield,0)
		FROM tbl_conversations
		INNER JOIN tbl_users ON tbl_conversations.user=tbl_users.id
		INNER JOIN tbl_messages ON tbl_conversations.id=tbl_messages.conversation AND tbl_messages.parent IS NULL
		INNER JOIN tbl_charges ON tbl_messages.id=tbl_charges.message
		LEFT JOIN (
			SELECT parent, SUM(IFNULL(amount,0)) AS yield
			FROM tbl_yields
			GROUP BY parent
		) AS yields ON tbl_messages.id=yields.parent
		WHERE tbl_conversations.created_unix>=?
		ORDER BY yields.yield DESC
		LIMIT ?`, since.Unix(), limit)
	if err != nil {
		return err
	}
	for rows.Next() {
		var (
			id       int64
			user     int64
			username string
			email    string
			topic    string
			created  int64
			cost     int64
			yield    int64
		)
		if err := rows.Scan(&id, &user, &username, &email, &topic, &created, &cost, &yield); err != nil {
			return err
		}
		if err := callback(id, &authgo.Account{
			ID:       user,
			Username: username,
			Email:    email,
		}, topic, time.Unix(created, 0), cost, yield); err != nil {
			return err
		}
	}
	return rows.Err()
}

func (db *Sql) SelectRecentConversations(callback func(int64, *authgo.Account, string, time.Time, int64, int64) error, limit int64) error {
	rows, err := db.Query(`
		SELECT tbl_conversations.id, tbl_conversations.user, tbl_users.username, tbl_users.email, tbl_conversations.topic, tbl_conversations.created_unix, tbl_charges.amount, IFNULL(yields.yield, 0)
		FROM tbl_conversations
		INNER JOIN tbl_users ON tbl_conversations.user=tbl_users.id
		INNER JOIN tbl_messages ON tbl_conversations.id=tbl_messages.conversation AND tbl_messages.parent IS NULL
		INNER JOIN tbl_charges ON tbl_messages.id=tbl_charges.message
		LEFT JOIN (
			SELECT parent, SUM(IFNULL(amount,0)) AS yield
			FROM tbl_yields
			GROUP BY parent
		) AS yields ON tbl_messages.id=yields.parent
		ORDER BY tbl_conversations.created_unix DESC
		LIMIT ?`, limit)
	if err != nil {
		return err
	}
	for rows.Next() {
		var (
			id       int64
			user     int64
			username string
			email    string
			topic    string
			created  int64
			cost     int64
			yield    int64
		)
		if err := rows.Scan(&id, &user, &username, &email, &topic, &created, &cost, &yield); err != nil {
			return err
		}
		if err := callback(id, &authgo.Account{
			ID:       user,
			Username: username,
			Email:    email,
		}, topic, time.Unix(created, 0), cost, yield); err != nil {
			return err
		}
	}
	return rows.Err()
}

func (db *Sql) CreateMessage(user, conversation, parent int64, created time.Time) (int64, error) {
	var (
		result sql.Result
		err    error
	)
	if parent == 0 {
		result, err = db.Exec(`
			INSERT INTO tbl_messages
			SET user=?, conversation=?, created_unix=?`, user, conversation, created.Unix())
	} else {
		result, err = db.Exec(`
			INSERT INTO tbl_messages
			SET user=?, conversation=?, parent=?, created_unix=?`, user, conversation, parent, created.Unix())
	}
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (db *Sql) CreateFile(message int64, hash, mime string, created time.Time) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO tbl_files
		SET message=?, hash=?, mime=?, created_unix=?`, message, hash, mime, created.Unix())
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (db *Sql) SelectMessage(id int64) (*authgo.Account, int64, int64, time.Time, int64, int64, error) {
	row := db.QueryRow(`
		SELECT tbl_messages.user, tbl_users.username, tbl_users.email, tbl_messages.conversation, IFNULL(tbl_messages.parent, 0), tbl_messages.created_unix, tbl_charges.amount, IFNULL(yields.yield, 0)
		FROM tbl_messages
		INNER JOIN tbl_users ON tbl_messages.user=tbl_users.id
		INNER JOIN tbl_charges ON tbl_messages.id=tbl_charges.message
		LEFT JOIN (
			SELECT parent, SUM(IFNULL(amount,0)) AS yield
			FROM tbl_yields
			GROUP BY parent
		) AS yields ON tbl_messages.id=yields.parent
		WHERE tbl_messages.id=?`, id)

	var (
		user         int64
		username     string
		email        string
		conversation int64
		parent       int64
		created      int64
		cost         int64
		yield        int64
	)
	if err := row.Scan(&user, &username, &email, &conversation, &parent, &created, &cost, &yield); err != nil {
		return nil, 0, 0, time.Time{}, 0, 0, err
	}
	return &authgo.Account{
		ID:       user,
		Username: username,
		Email:    email,
	}, conversation, parent, time.Unix(created, 0), cost, yield, nil
}

func (db *Sql) SelectMessages(conversation int64, callback func(int64, *authgo.Account, int64, time.Time, int64, int64) error) error {
	rows, err := db.Query(`
		SELECT tbl_messages.id, tbl_messages.user, tbl_users.username, tbl_users.email, IFNULL(tbl_messages.parent, 0), tbl_messages.created_unix, tbl_charges.amount, IFNULL(yields.yield, 0)
		FROM tbl_messages
		INNER JOIN tbl_users ON tbl_messages.user=tbl_users.id
		INNER JOIN tbl_charges ON tbl_messages.id=tbl_charges.message
		LEFT JOIN (
			SELECT parent, SUM(amount) AS yield
			FROM tbl_yields
			GROUP BY parent
		) AS yields ON tbl_messages.id=yields.parent
		WHERE tbl_messages.conversation=?`, conversation)
	if err != nil {
		return err
	}
	for rows.Next() {
		var (
			id       int64
			user     int64
			username string
			email    string
			parent   int64
			created  int64
			cost     int64
			yield    int64
		)
		if err := rows.Scan(&id, &user, &username, &email, &parent, &created, &cost, &yield); err != nil {
			return err
		}
		if err := callback(id, &authgo.Account{
			ID:       user,
			Username: username,
			Email:    email,
		}, parent, time.Unix(created, 0), cost, yield); err != nil {
			return err
		}
	}
	return rows.Err()
}

func (db *Sql) SelectMessageParent(id int64) (int64, error) {
	row := db.QueryRow(`
			SELECT IFNULL(parent, 0)
			FROM tbl_messages
			WHERE id=?`, id)
	var (
		parent int64
	)
	if err := row.Scan(&parent); err != nil {
		return 0, err
	}
	return parent, nil
}

func (db *Sql) SelectFile(id int64) (int64, string, string, time.Time, error) {
	row := db.QueryRow(`
		SELECT tbl_files.message, tbl_files.hash, tbl_files.mime, tbl_files.created_unix
		FROM tbl_files
		WHERE tbl_files.id=?`, id)

	var (
		message int64
		hash    string
		mime    string
		created int64
	)
	if err := row.Scan(&message, &hash, &mime, &created); err != nil {
		return 0, "", "", time.Time{}, err
	}
	return message, hash, mime, time.Unix(created, 0), nil
}

func (db *Sql) SelectFiles(message int64, callback func(int64, string, string, time.Time) error) error {
	rows, err := db.Query(`
		SELECT tbl_files.id, tbl_files.hash, tbl_files.mime, tbl_files.created_unix
		FROM tbl_files
		WHERE tbl_files.message=?`, message)
	if err != nil {
		return err
	}
	for rows.Next() {
		var (
			id      int64
			hash    string
			mime    string
			created int64
		)
		if err := rows.Scan(&id, &hash, &mime, &created); err != nil {
			return err
		}
		if err := callback(id, hash, mime, time.Unix(created, 0)); err != nil {
			return err
		}
	}
	return rows.Err()
}

func (db *Sql) CreateCharge(user, conversation, message, amount int64, created time.Time) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO tbl_charges
		SET user=?, conversation=?, message=?, amount=?, created_unix=?`, user, conversation, message, amount, created.Unix())
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (db *Sql) SelectCharges(user int64) (int64, error) {
	row := db.QueryRow(`
		SELECT SUM(IFNULL(total_charges,0))
		FROM tbl_users
		LEFT JOIN (
			SELECT id, user
			FROM tbl_messages
		) AS ms ON ms.user=tbl_users.id
		LEFT JOIN (
			SELECT message, SUM(IFNULL(amount,0)) AS total_charges
			FROM tbl_charges
			GROUP BY message
		) AS cs ON cs.message=ms.id
		WHERE tbl_users.id=?`, user)
	var (
		charges int64
	)
	if err := row.Scan(&charges); err != nil {
		return 0, err
	}
	return charges, nil
}

func (db *Sql) CreateYield(user, conversation, message, parent, amount int64, created time.Time) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO tbl_yields
		SET user=?, conversation=?, message=?, parent=?, amount=?, created_unix=?`, user, conversation, message, parent, amount, created.Unix())
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (db *Sql) SelectYields(user int64) (int64, error) {
	row := db.QueryRow(`
		SELECT SUM(IFNULL(total_yields,0))
		FROM tbl_users
		LEFT JOIN (
			SELECT id, user
			FROM tbl_messages
		) AS ms ON ms.user=tbl_users.id
		LEFT JOIN (
			SELECT parent, SUM(IFNULL(amount,0)) AS total_yields
			FROM tbl_yields
			GROUP BY parent
		) AS ys ON ys.parent=ms.id
		WHERE tbl_users.id=?`, user)
	var (
		yields int64
	)
	if err := row.Scan(&yields); err != nil {
		return 0, err
	}
	return yields, nil
}

func (db *Sql) CreatePurchase(user int64, sessionID, customerID, paymentIntentID, currency string, amount, size int64, created time.Time) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO tbl_purchases
		SET user=?, stripe_session=?, stripe_customer=?, stripe_payment_intent=?, stripe_currency=?, stripe_amount=?, bundle_size=?, created_unix=?`, user, sessionID, customerID, paymentIntentID, currency, amount, size, created.Unix())
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (db *Sql) SelectPurchases(user int64) (int64, error) {
	row := db.QueryRow(`
		SELECT IFNULL(SUM(IFNULL(bundle_size,0)),0)
		FROM tbl_purchases
		WHERE tbl_purchases.user=?`, user)
	var (
		purchases int64
	)
	if err := row.Scan(&purchases); err != nil {
		return 0, err
	}
	return purchases, nil
}

func (db *Sql) SelectNotificationPreferences(user int64) (int64, bool, bool, bool, error) {
	row := db.QueryRow(`
		SELECT id, responses, mentions, digests
		FROM tbl_notification_preferences
		WHERE user=?`, user)

	var (
		id        int64
		responses bool
		mentions  bool
		digests   bool
	)
	if err := row.Scan(&id, &responses, &mentions, &digests); err != nil {
		if err == sql.ErrNoRows {
			// Notification preferences default to enabled
			return 0, true, true, true, nil
		}
		return 0, false, false, false, err
	}
	return id, responses, mentions, digests, nil
}

func (db *Sql) UpdateNotificationPreferences(id, user int64, responses, mentions, digests bool) (int64, error) {
	var (
		result sql.Result
		err    error
	)
	if id == 0 {
		result, err = db.Exec(`
		INSERT INTO tbl_notification_preferences (user, responses, mentions, digests)
		VALUES (?, ?, ?, ?)`, user, responses, mentions, digests)
	} else {
		result, err = db.Exec(`
		UPDATE tbl_notification_preferences
		SET user=?, responses=?, mentions=?, digests=?
		WHERE id=?`, user, responses, mentions, digests, id)
	}
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (db *Sql) SelectAwards(user int64) (int64, error) {
	row := db.QueryRow(`
		SELECT IFNULL(SUM(IFNULL(amount,0)),0)
		FROM tbl_awards
		WHERE tbl_awards.user=?`, user)
	var (
		awards int64
	)
	if err := row.Scan(&awards); err != nil {
		return 0, err
	}
	return awards, nil
}
