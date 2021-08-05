package main

import (
	"aletheiaware.com/authgo"
	"aletheiaware.com/authgo/authtest"
	"aletheiaware.com/authgo/email"
	authhandler "aletheiaware.com/authgo/handler"
	"aletheiaware.com/conveyearthgo"
	"aletheiaware.com/conveyearthgo/database"
	"aletheiaware.com/conveyearthgo/filesystem"
	"aletheiaware.com/conveyearthgo/handler"
	"aletheiaware.com/netgo"
	nethandler "aletheiaware.com/netgo/handler"
	"crypto/tls"
	"embed"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/stripe/stripe-go/v72"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

//go:embed assets
var embeddedFS embed.FS

func main() {
	// Configure Logging
	logs, ok := os.LookupEnv("LOG_DIRECTORY")
	if !ok {
		logs = "logs"
	}
	if err := os.MkdirAll(logs, os.ModePerm); err != nil {
		log.Fatal(err)
	}
	logFile, err := os.OpenFile(path.Join(logs, time.Now().UTC().Format(time.RFC3339)), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Log:", logFile.Name())

	secure := netgo.IsSecure()

	scheme := "http"
	if secure {
		scheme = "https"
	}

	host, ok := os.LookupEnv("HOST")
	if !ok {
		log.Fatal(errors.New("Missing HOST environment variable"))
	}

	// Configure Stripe
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// Create Multiplexer
	mux := http.NewServeMux()

	nethandler.AttachHealthHandler(mux)

	// Handle Static Assets
	staticFS, err := fs.Sub(embeddedFS, path.Join("assets", "html", "static"))
	if err != nil {
		log.Fatal(err)
	}
	nethandler.AttachStaticFSHandler(mux, staticFS, false)

	// Parse Templates
	templateFS, err := fs.Sub(embeddedFS, path.Join("assets", "html", "template"))
	if err != nil {
		log.Fatal(err)
	}
	templates, err := template.ParseFS(templateFS, "*.go.html")
	if err != nil {
		log.Fatal(err)
	}

	// Create Database
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbSecure := secure
	if dbHost == "" || dbHost == "localhost" {
		// XXX FIXME Disable TLS for local connections
		dbSecure = false
	}
	db, err := database.NewSql(dbName, dbUser, dbPassword, dbHost, dbPort, dbSecure)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Migrate Database
	migrationFS, err := fs.Sub(embeddedFS, path.Join("assets", "database", "migrations"))
	if err != nil {
		log.Fatal(err)
	}
	m, err := db.Migrator(migrationFS)
	if err != nil {
		log.Fatal(err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}

	// Create Email Verifier
	ev := authtest.NewEmailVerifier()
	if secure {
		server := os.Getenv("SMTP_SERVER")
		sender := os.Getenv("SMTP_SENDER")
		ev = email.NewSmtpEmailVerifier(server, host, sender, templates.Lookup("email-verification.go.html"))
	}

	// Create an Authenticator
	auth := authgo.NewAuthenticator(db, ev)

	// Attach Authentication Handlers
	// authhandler.AttachAccountHandler(mux, auth, templates) - replaced with custom handler
	authhandler.AttachAccountPasswordHandler(mux, auth, templates)
	authhandler.AttachAccountRecoveryHandler(mux, auth, templates)
	authhandler.AttachSignInHandler(mux, auth, templates)
	authhandler.AttachSignOutHandler(mux, auth, templates)
	authhandler.AttachSignUpHandler(mux, auth, templates)

	// Create a Account Manager
	am := conveyearthgo.NewAccountManager(db)

	uploads, ok := os.LookupEnv("USER_CONTENT_DIRECTORY")
	if !ok {
		uploads = "uploads"
	}
	if err := os.MkdirAll(uploads, os.ModePerm); err != nil {
		log.Fatal(err)
	}
	log.Println("User Content Directory:", uploads)

	// Create a Content Manager
	cm := conveyearthgo.NewContentManager(db, filesystem.NewOnDisk(uploads))

	// Handle Content
	handler.AttachContentHandler(mux, cm)

	digests, ok := os.LookupEnv("DIGEST_DIRECTORY")
	if !ok {
		digests = "digests"
	}
	if err := os.MkdirAll(digests, os.ModePerm); err != nil {
		log.Fatal(err)
	}
	log.Println("Digests Directory:", digests)

	// Create a Digest Manager
	dm := conveyearthgo.NewDigestManager(filesystem.NewOnDisk(digests))

	// Handle Digest
	handler.AttachDigestHandler(mux, auth, dm, templates)

	// Handle Account
	handler.AttachAccountHandler(mux, auth, am, templates)

	// Handle Buy Coins
	handler.AttachCoinBuyHandler(mux, auth, am, templates, scheme, host)

	// Handle Stripe Webhook
	handler.AttachStripeWebhookHandler(mux, am, os.Getenv("STRIPE_WEBHOOK_SECRET_KEY"), host)

	// Handle Conversation
	handler.AttachConversationHandler(mux, auth, cm, templates)

	// Handle Message
	handler.AttachMessageHandler(mux, auth, cm, templates)

	// Handle Publish
	handler.AttachPublishHandler(mux, auth, am, cm, templates)

	// Handle Reply
	handler.AttachReplyHandler(mux, auth, am, cm, templates)

	// Handle Best
	handler.AttachBestHandler(mux, auth, cm, templates)

	// Handle Recent
	handler.AttachRecentHandler(mux, auth, cm, templates)

	// Handle Demo
	handler.AttachDemoHandler(mux, auth, templates)

	// Handle robots.txt
	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static/robots.txt", http.StatusFound)
	})

	// Handle Index
	handler.AttachIndexHandler(mux, auth, templates)

	// Start Server
	if secure {
		routeMap := make(map[string]bool)

		routes, ok := os.LookupEnv("ROUTES")
		if ok {
			for _, route := range strings.Split(routes, ",") {
				routeMap[route] = true
			}
		}

		// Redirect HTTP Requests to HTTPS
		go http.ListenAndServe(":80", http.HandlerFunc(netgo.HTTPSRedirect(host, routeMap)))

		certificates, ok := os.LookupEnv("CERTIFICATE_DIRECTORY")
		if !ok {
			certificates = "certificates"
		}
		log.Println("Certificate Directory:", certificates)

		// Serve HTTPS Requests
		config := &tls.Config{MinVersion: tls.VersionTLS10}
		server := &http.Server{Addr: ":443",
			Handler:           mux,
			TLSConfig:         config,
			ReadTimeout:       5 * time.Second,
			ReadHeaderTimeout: 5 * time.Second,
			WriteTimeout:      5 * time.Second,
			IdleTimeout:       5 * time.Second,
		}
		if err := server.ListenAndServeTLS(path.Join(certificates, "fullchain.pem"), path.Join(certificates, "privkey.pem")); err != nil {
			log.Fatal(err)
		}
	} else {
		// Server HTTP Requests
		log.Println("HTTP Server Listening on :80")
		if err := http.ListenAndServe(":80", mux); err != nil {
			log.Fatal(err)
		}
	}
}
