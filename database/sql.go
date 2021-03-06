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
	SELECT parent, SUM(IFNULL(amount, 0)) AS yield
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
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?tls=%t&multiStatements=true", username, password, host, port, dbname, secure)
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
	if len(email) > authgo.MAXIMUM_EMAIL_LENGTH {
		email = email[:authgo.MAXIMUM_EMAIL_LENGTH]
	}
	if len(username) > authgo.MAXIMUM_USERNAME_LENGTH {
		username = username[:authgo.MAXIMUM_USERNAME_LENGTH]
	}
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
		WHERE deleted_at=0 AND username=?`, username)

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
		WHERE deleted_at=0 AND email=?`, email)
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
		WHERE deleted_at=0 AND username=?`, password, username)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (db *Sql) IsEmailVerified(email string) (bool, error) {
	row := db.QueryRow(`
		SELECT verified
		FROM tbl_users
		WHERE deleted_at=0 AND email=?`, email)

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
		WHERE deleted_at=0 AND email=?`, verified, email)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (db *Sql) DeactivateUser(username string, deleted time.Time) (int64, error) {
	d := deleted.Unix()
	{
		// Deactivate user
		result, err := db.Exec(`
			UPDATE tbl_users
			SET deleted_at=?
			WHERE username=?`, d, username)
		if err != nil {
			return 0, err
		}
		count, err := result.RowsAffected()
		if err != nil || count == 0 {
			return 0, err
		}
	}
	{
		// Sign out all sessions
		result, err := db.Exec(`
			UPDATE tbl_signins
			SET authenticated=FALSE
			WHERE username=?`, username)
		if err != nil {
			return 0, err
		}
		if _, err := result.RowsAffected(); err != nil {
			return 0, err
		}
	}
	return 1, nil
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

func (db *Sql) SelectSignUpSession(token string) (string, string, string, string, string, time.Time, error) {
	row := db.QueryRow(`
		SELECT IFNULL(error, ''), IFNULL(email, ''), IFNULL(username, ''), IFNULL(referrer, ''), IFNULL(challenge, ''), created_unix
		FROM tbl_signups
		WHERE token=?`, token)

	var (
		error     string
		email     string
		username  string
		referrer  string
		challenge string
		created   int64
	)
	if err := row.Scan(&error, &email, &username, &referrer, &challenge, &created); err != nil {
		return "", "", "", "", "", time.Time{}, err
	}
	return error, email, username, referrer, challenge, time.Unix(created, 0), nil
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
	if len(email) > authgo.MAXIMUM_EMAIL_LENGTH {
		email = email[:authgo.MAXIMUM_EMAIL_LENGTH]
	}
	if len(username) > authgo.MAXIMUM_USERNAME_LENGTH {
		username = username[:authgo.MAXIMUM_USERNAME_LENGTH]
	}
	result, err := db.Exec(`
		UPDATE tbl_signups
		SET email=?, username=?
		WHERE token=?`, email, username, token)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (db *Sql) UpdateSignUpSessionReferrer(token, referrer string) (int64, error) {
	if len(referrer) > authgo.MAXIMUM_USERNAME_LENGTH {
		referrer = referrer[:authgo.MAXIMUM_USERNAME_LENGTH]
	}
	var (
		result sql.Result
		err    error
	)
	if referrer == "" {
		result, err = db.Exec(`
			UPDATE tbl_signups
			SET referrer=NULL
			WHERE token=?`, token)
	} else {
		result, err = db.Exec(`
			UPDATE tbl_signups
			SET referrer=?
			WHERE token=?`, referrer, token)
	}
	if err != nil {
		if driverErr, ok := err.(*mysql.MySQLError); ok {
			switch driverErr.Number {
			case 1452: // ER_NO_REFERENCED_ROW_2
				return 0, authgo.ErrInvalidReferrer
			}
		}
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
	if len(username) > authgo.MAXIMUM_USERNAME_LENGTH {
		username = username[:authgo.MAXIMUM_USERNAME_LENGTH]
	}
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
	if len(username) > authgo.MAXIMUM_USERNAME_LENGTH {
		username = username[:authgo.MAXIMUM_USERNAME_LENGTH]
	}
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
	if len(username) > authgo.MAXIMUM_USERNAME_LENGTH {
		username = username[:authgo.MAXIMUM_USERNAME_LENGTH]
	}
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
	if len(email) > authgo.MAXIMUM_EMAIL_LENGTH {
		email = email[:authgo.MAXIMUM_EMAIL_LENGTH]
	}
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
	if len(username) > authgo.MAXIMUM_USERNAME_LENGTH {
		username = username[:authgo.MAXIMUM_USERNAME_LENGTH]
	}
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

func (db *Sql) DeleteConversation(user, id int64, deleted time.Time) (int64, error) {
	result, err := db.Exec(`
		UPDATE tbl_conversations
		SET deleted_at=?
		WHERE user=? AND id=?`, deleted.Unix(), user, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (db *Sql) SelectConversation(id int64) (*authgo.Account, string, time.Time, error) {
	row := db.QueryRow(`
		SELECT tbl_conversations.user, tbl_users.username, tbl_users.email, tbl_users.created_unix, tbl_conversations.topic, tbl_conversations.created_unix
		FROM tbl_conversations
		INNER JOIN tbl_users ON tbl_conversations.user=tbl_users.id
		WHERE tbl_users.deleted_at=0 AND tbl_conversations.deleted_at=0 AND tbl_conversations.id=?`, id)

	var (
		user     int64
		username string
		email    string
		joined   int64
		topic    string
		created  int64
	)
	if err := row.Scan(&user, &username, &email, &joined, &topic, &created); err != nil {
		return nil, "", time.Time{}, err
	}
	return &authgo.Account{
		ID:       user,
		Username: username,
		Email:    email,
		Created:  time.Unix(joined, 0),
	}, topic, time.Unix(created, 0), nil
}

func (db *Sql) SelectBestConversations(callback func(int64, *authgo.Account, string, time.Time, int64, int64) error, since time.Time, limit int64) error {
	rows, err := db.Query(`
		SELECT tbl_conversations.id, tbl_conversations.user, tbl_users.username, tbl_users.email, tbl_users.created_unix, tbl_conversations.topic, tbl_conversations.created_unix, tbl_charges.amount, IFNULL(yields.yield, 0)
		FROM tbl_conversations
		INNER JOIN tbl_users ON tbl_conversations.user=tbl_users.id
		INNER JOIN tbl_messages ON tbl_conversations.id=tbl_messages.conversation AND tbl_messages.parent IS NULL
		INNER JOIN tbl_charges ON tbl_messages.id=tbl_charges.message
		LEFT JOIN (
			SELECT parent, SUM(IFNULL(amount, 0)) AS yield
			FROM tbl_yields
			WHERE deleted_at=0
			GROUP BY parent
		) AS yields ON tbl_messages.id=yields.parent
		WHERE tbl_users.deleted_at=0 AND tbl_conversations.deleted_at=0 AND tbl_messages.deleted_at=0 AND tbl_charges.deleted_at=0 AND tbl_conversations.created_unix>=? AND yields.yield>0
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
			joined   int64
			topic    string
			created  int64
			cost     int64
			yield    int64
		)
		if err := rows.Scan(&id, &user, &username, &email, &joined, &topic, &created, &cost, &yield); err != nil {
			return err
		}
		if err := callback(id, &authgo.Account{
			ID:       user,
			Username: username,
			Email:    email,
			Created:  time.Unix(joined, 0),
		}, topic, time.Unix(created, 0), cost, yield); err != nil {
			return err
		}
	}
	return rows.Err()
}

func (db *Sql) SelectRecentConversations(callback func(int64, *authgo.Account, string, time.Time, int64, int64) error, limit int64) error {
	rows, err := db.Query(`
		SELECT tbl_conversations.id, tbl_conversations.user, tbl_users.username, tbl_users.email, tbl_users.created_unix, tbl_conversations.topic, tbl_conversations.created_unix, tbl_charges.amount, IFNULL(yields.yield, 0)
		FROM tbl_conversations
		INNER JOIN tbl_users ON tbl_conversations.user=tbl_users.id
		INNER JOIN tbl_messages ON tbl_conversations.id=tbl_messages.conversation AND tbl_messages.parent IS NULL
		INNER JOIN tbl_charges ON tbl_messages.id=tbl_charges.message
		LEFT JOIN (
			SELECT parent, SUM(IFNULL(amount, 0)) AS yield
			FROM tbl_yields
			WHERE deleted_at=0
			GROUP BY parent
		) AS yields ON tbl_messages.id=yields.parent
		WHERE tbl_users.deleted_at=0 AND tbl_conversations.deleted_at=0 AND tbl_messages.deleted_at=0 AND tbl_charges.deleted_at=0
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
			joined   int64
			topic    string
			created  int64
			cost     int64
			yield    int64
		)
		if err := rows.Scan(&id, &user, &username, &email, &joined, &topic, &created, &cost, &yield); err != nil {
			return err
		}
		if err := callback(id, &authgo.Account{
			ID:       user,
			Username: username,
			Email:    email,
			Created:  time.Unix(joined, 0),
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

func (db *Sql) DeleteMessage(user, id int64, deleted time.Time) (int64, error) {
	d := deleted.Unix()
	{
		// Delete message
		result, err := db.Exec(`
			UPDATE tbl_messages
			SET deleted_at=?
			WHERE user=? AND id=? AND id NOT IN (
				SELECT parent
				FROM tbl_messages
				WHERE deleted_at=0 AND parent IS NOT NULL
			) AND id NOT IN (
				SELECT message
				FROM tbl_gifts
				WHERE deleted_at=0
			)`, d, user, id)
		if err != nil {
			return 0, err
		}
		count, err := result.RowsAffected()
		if err != nil || count == 0 {
			return 0, err
		}
	}
	{
		// Delete associated files
		result, err := db.Exec(`
			UPDATE tbl_files
			SET deleted_at=?
			WHERE message=?`, d, id)
		if err != nil {
			return 0, err
		}
		count, err := result.RowsAffected()
		if err != nil || count == 0 {
			return 0, err
		}
	}
	{
		// Delete associated charge
		result, err := db.Exec(`
			UPDATE tbl_charges
			SET deleted_at=?
			WHERE user=? AND message=?`, d, user, id)
		if err != nil {
			return 0, err
		}
		count, err := result.RowsAffected()
		if err != nil || count == 0 {
			return 0, err
		}
	}
	{
		// Delete associated yields
		result, err := db.Exec(`
			UPDATE tbl_yields
			SET deleted_at=?
			WHERE user=? AND message=?`, d, user, id)
		if err != nil {
			return 0, err
		}
		if _, err := result.RowsAffected(); err != nil {
			return 0, err
		}
	}
	return 1, nil
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
		SELECT tbl_messages.user, tbl_users.username, tbl_users.email, tbl_users.created_unix, tbl_messages.conversation, IFNULL(tbl_messages.parent, 0), tbl_messages.created_unix, tbl_charges.amount, IFNULL(yields.yield, 0)
		FROM tbl_messages
		INNER JOIN tbl_users ON tbl_messages.user=tbl_users.id
		INNER JOIN tbl_charges ON tbl_messages.id=tbl_charges.message
		LEFT JOIN (
			SELECT parent, SUM(IFNULL(amount, 0)) AS yield
			FROM tbl_yields
			WHERE deleted_at=0
			GROUP BY parent
		) AS yields ON tbl_messages.id=yields.parent
		WHERE tbl_users.deleted_at=0 AND tbl_messages.deleted_at=0 AND tbl_charges.deleted_at=0 AND tbl_messages.id=?`, id)

	var (
		user         int64
		username     string
		email        string
		joined       int64
		conversation int64
		parent       int64
		created      int64
		cost         int64
		yield        int64
	)
	if err := row.Scan(&user, &username, &email, &joined, &conversation, &parent, &created, &cost, &yield); err != nil {
		return nil, 0, 0, time.Time{}, 0, 0, err
	}
	return &authgo.Account{
		ID:       user,
		Username: username,
		Email:    email,
		Created:  time.Unix(joined, 0),
	}, conversation, parent, time.Unix(created, 0), cost, yield, nil
}

func (db *Sql) SelectMessages(conversation int64, callback func(int64, *authgo.Account, int64, time.Time, int64, int64) error) error {
	rows, err := db.Query(`
		SELECT tbl_messages.id, tbl_messages.user, tbl_users.username, tbl_users.email, tbl_users.created_unix, IFNULL(tbl_messages.parent, 0), tbl_messages.created_unix, tbl_charges.amount, IFNULL(yields.yield, 0)
		FROM tbl_messages
		INNER JOIN tbl_users ON tbl_messages.user=tbl_users.id
		INNER JOIN tbl_charges ON tbl_messages.id=tbl_charges.message
		LEFT JOIN (
			SELECT parent, SUM(amount) AS yield
			FROM tbl_yields
			WHERE deleted_at=0
			GROUP BY parent
		) AS yields ON tbl_messages.id=yields.parent
		WHERE tbl_users.deleted_at=0 AND tbl_messages.deleted_at=0 AND tbl_charges.deleted_at=0 AND tbl_messages.conversation=?`, conversation)
	if err != nil {
		return err
	}
	for rows.Next() {
		var (
			id       int64
			user     int64
			username string
			email    string
			joined   int64
			parent   int64
			created  int64
			cost     int64
			yield    int64
		)
		if err := rows.Scan(&id, &user, &username, &email, &joined, &parent, &created, &cost, &yield); err != nil {
			return err
		}
		if err := callback(id, &authgo.Account{
			ID:       user,
			Username: username,
			Email:    email,
			Created:  time.Unix(joined, 0),
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
			WHERE deleted_at=0 AND id=?`, id)
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
		SELECT message, hash, mime, created_unix
		FROM tbl_files
		WHERE deleted_at=0 AND id=?`, id)

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
		SELECT id, hash, mime, created_unix
		FROM tbl_files
		WHERE deleted_at=0 AND message=?`, message)
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

func (db *Sql) SelectChargesForUser(user int64) (int64, error) {
	row := db.QueryRow(`
		SELECT SUM(IFNULL(total_charges, 0))
		FROM tbl_users
		LEFT JOIN (
			SELECT id, user
			FROM tbl_messages
			WHERE deleted_at=0
		) AS ms ON ms.user=tbl_users.id
		LEFT JOIN (
			SELECT message, SUM(IFNULL(amount, 0)) AS total_charges
			FROM tbl_charges
			WHERE deleted_at=0
			GROUP BY message
		) AS cs ON cs.message=ms.id
		WHERE tbl_users.deleted_at=0 AND tbl_users.id=?`, user)
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

func (db *Sql) SelectYieldsForUser(user int64) (int64, error) {
	row := db.QueryRow(`
		SELECT SUM(IFNULL(total_yields, 0))
		FROM tbl_users
		LEFT JOIN (
			SELECT id, user
			FROM tbl_messages
			WHERE deleted_at=0
		) AS ms ON ms.user=tbl_users.id
		LEFT JOIN (
			SELECT parent, SUM(IFNULL(amount, 0)) AS total_yields
			FROM tbl_yields
			WHERE deleted_at=0
			GROUP BY parent
		) AS ys ON ys.parent=ms.id
		WHERE tbl_users.deleted_at=0 AND tbl_users.id=?`, user)
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

func (db *Sql) SelectPurchasesForUser(user int64) (int64, error) {
	row := db.QueryRow(`
		SELECT IFNULL(SUM(IFNULL(bundle_size, 0)), 0)
		FROM tbl_purchases
		WHERE deleted_at=0 AND user=?`, user)
	var (
		purchases int64
	)
	if err := row.Scan(&purchases); err != nil {
		return 0, err
	}
	return purchases, nil
}

func (db *Sql) SelectNotificationPreferences(user int64) (int64, bool, bool, bool, bool, error) {
	row := db.QueryRow(`
		SELECT id, responses, mentions, gifts, digests
		FROM tbl_notification_preferences
		WHERE user=?`, user)

	var (
		id        int64
		responses bool
		mentions  bool
		gifts     bool
		digests   bool
	)
	if err := row.Scan(&id, &responses, &mentions, &gifts, &digests); err != nil {
		if err == sql.ErrNoRows {
			// Notification preferences default to enabled
			return 0, true, true, true, true, nil
		}
		return 0, false, false, false, false, err
	}
	return id, responses, mentions, gifts, digests, nil
}

func (db *Sql) UpdateNotificationPreferences(id, user int64, responses, mentions, gifts, digests bool) (int64, error) {
	var (
		result sql.Result
		err    error
	)
	if id == 0 {
		result, err = db.Exec(`
		INSERT INTO tbl_notification_preferences (user, responses, mentions, gifts, digests)
		VALUES (?, ?, ?, ?, ?)`, user, responses, mentions, gifts, digests)
	} else {
		result, err = db.Exec(`
		UPDATE tbl_notification_preferences
		SET user=?, responses=?, mentions=?, gifts=?, digests=?
		WHERE id=?`, user, responses, mentions, gifts, digests, id)
	}
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (db *Sql) SelectAwardsForUser(user int64) (int64, error) {
	row := db.QueryRow(`
		SELECT IFNULL(SUM(IFNULL(amount, 0)), 0)
		FROM tbl_awards
		WHERE deleted_at=0 AND user=?`, user)
	var (
		awards int64
	)
	if err := row.Scan(&awards); err != nil {
		return 0, err
	}
	return awards, nil
}

func (db *Sql) CreateStripeAccount(user int64, identity string, created time.Time) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO tbl_stripe_account
		SET user=?, identity=?, created_unix=?`, user, identity, created.Unix())
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (db *Sql) SelectStripeAccount(user int64) (string, time.Time, error) {
	row := db.QueryRow(`
		SELECT identity, created_unix
		FROM tbl_stripe_account
		WHERE deleted_at=0 AND user=?`, user)
	var (
		identity string
		created  int64
	)
	if err := row.Scan(&identity, &created); err != nil {
		return "", time.Time{}, err
	}
	return identity, time.Unix(created, 0), nil
}

func (db *Sql) CreateGift(user, conversation, message, amount int64, created time.Time) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO tbl_gifts
		SET user=?, conversation=?, message=?, amount=?, created_unix=?`, user, conversation, message, amount, created.Unix())
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (db *Sql) DeleteGift(user, id int64, deleted time.Time) (int64, error) {
	result, err := db.Exec(`
		UPDATE tbl_gifts
		SET deleted_at=?
		WHERE user=? AND id=?`, deleted.Unix(), user, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (db *Sql) SelectGift(id int64) (int64, int64, *authgo.Account, int64, time.Time, error) {
	row := db.QueryRow(`
		SELECT tbl_gifts.user, tbl_users.username, tbl_users.email, tbl_users.created_unix, tbl_gifts.conversation, tbl_gifts.message, tbl_gifts.amount, tbl_gifts.created_unix
		FROM tbl_gifts
		INNER JOIN tbl_users ON tbl_gifts.user=tbl_users.id
		WHERE tbl_users.deleted_at=0 AND tbl_gifts.deleted_at=0 AND tbl_gifts.id=?`, id)
	var (
		user         int64
		username     string
		email        string
		joined       int64
		conversation int64
		message      int64
		amount       int64
		created      int64
	)
	if err := row.Scan(&user, &username, &email, &joined, &conversation, &message, &amount, &created); err != nil {
		return 0, 0, nil, 0, time.Time{}, err
	}
	return conversation, message, &authgo.Account{
		ID:       user,
		Username: username,
		Email:    email,
		Created:  time.Unix(joined, 0),
	}, amount, time.Unix(created, 0), nil
}

func (db *Sql) SelectGifts(conversation, message int64, callback func(int64, int64, int64, *authgo.Account, int64, time.Time) error) error {
	var (
		rows *sql.Rows
		err  error
	)
	if message <= 0 {
		rows, err = db.Query(`
		SELECT tbl_gifts.id, tbl_gifts.user, tbl_users.username, tbl_users.email, tbl_users.created_unix, tbl_gifts.conversation, tbl_gifts.message, tbl_gifts.amount, tbl_gifts.created_unix
		FROM tbl_gifts
		INNER JOIN tbl_users ON tbl_gifts.user=tbl_users.id
		WHERE tbl_users.deleted_at=0 AND tbl_gifts.deleted_at=0 AND tbl_gifts.conversation=?`, conversation)
	} else {
		rows, err = db.Query(`
		SELECT tbl_gifts.id, tbl_gifts.user, tbl_users.username, tbl_users.email, tbl_users.created_unix, tbl_gifts.conversation, tbl_gifts.message, tbl_gifts.amount, tbl_gifts.created_unix
		FROM tbl_gifts
		INNER JOIN tbl_users ON tbl_gifts.user=tbl_users.id
		WHERE tbl_users.deleted_at=0 AND tbl_gifts.deleted_at=0 AND tbl_gifts.conversation=? AND tbl_gifts.message=?`, conversation, message)
	}
	if err != nil {
		return err
	}
	for rows.Next() {
		var (
			id           int64
			user         int64
			username     string
			email        string
			joined       int64
			conversation int64
			message      int64
			amount       int64
			created      int64
		)
		if err := rows.Scan(&id, &user, &username, &email, &joined, &conversation, &message, &amount, &created); err != nil {
			return err
		}
		if err := callback(id, conversation, message, &authgo.Account{
			ID:       user,
			Username: username,
			Email:    email,
			Created:  time.Unix(joined, 0),
		}, amount, time.Unix(created, 0)); err != nil {
			return err
		}
	}
	return rows.Err()
}

func (db *Sql) SelectGiftsForUser(user int64) (int64, error) {
	row := db.QueryRow(`
		SELECT SUM(IFNULL(total_amounts, 0))
		FROM tbl_users
		LEFT JOIN (
			SELECT id, user
			FROM tbl_messages
			WHERE deleted_at=0
		) AS ms ON ms.user=tbl_users.id
		LEFT JOIN (
			SELECT message, SUM(IFNULL(amount, 0)) AS total_amounts
			FROM tbl_gifts
			WHERE deleted_at=0
			GROUP BY message
		) AS ys ON ys.message=ms.id
		WHERE tbl_users.deleted_at=0 AND tbl_users.id=?`, user)
	var (
		gifts int64
	)
	if err := row.Scan(&gifts); err != nil {
		return 0, err
	}
	return gifts, nil
}

func (db *Sql) SelectGiftsFromUser(user int64) (int64, error) {
	row := db.QueryRow(`
		SELECT IFNULL(SUM(IFNULL(amount, 0)), 0)
		FROM tbl_gifts
		WHERE deleted_at=0 AND user=?`, user)
	var (
		gifts int64
	)
	if err := row.Scan(&gifts); err != nil {
		return 0, err
	}
	return gifts, nil
}
