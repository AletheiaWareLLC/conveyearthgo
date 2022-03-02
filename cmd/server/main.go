package main

import (
	"aletheiaware.com/authgo"
	"aletheiaware.com/authgo/authtest"
	"aletheiaware.com/authgo/email"
	authhandler "aletheiaware.com/authgo/handler"
	"aletheiaware.com/conveyearthgo"
	"aletheiaware.com/conveyearthgo/conveytest"
	"aletheiaware.com/conveyearthgo/database"
	"aletheiaware.com/conveyearthgo/filesystem"
	"aletheiaware.com/conveyearthgo/handler"
	"aletheiaware.com/netgo"
	nethandler "aletheiaware.com/netgo/handler"
	"crypto/tls"
	"embed"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/stripe/stripe-go/v72"
	"html/template"
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
	logFile, err := netgo.SetupLogging()
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()
	log.Println("Log File:", logFile.Name())

	secure := netgo.IsSecure()

	scheme := conveyearthgo.Scheme()

	host := conveyearthgo.Host()
	if host == "" {
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
	nethandler.AttachStaticFSHandler(mux, staticFS, false, fmt.Sprintf("public, max-age=%d", 60*60*24*7*52)) // 52 week max-age

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
	var ev authgo.EmailVerifier
	if secure {
		server := os.Getenv("SMTP_SERVER")
		sender := os.Getenv("SMTP_SENDER")
		ev = email.NewSmtpEmailVerifier(server, host, sender, templates.Lookup("email-verification.go.html"))
	} else {
		ev = authtest.NewEmailVerifier()
	}

	// Create an Authenticator
	auth := authgo.NewAuthenticator(db, ev)

	// Attach Authentication Handlers
	// authhandler.AttachAccountHandler(mux, auth, templates) - replaced with custom handler
	authhandler.AttachAccountPasswordHandler(mux, auth, templates)
	authhandler.AttachAccountRecoveryHandler(mux, auth, templates)
	authhandler.AttachAccountDeactivateHandler(mux, auth, templates)
	authhandler.AttachSignInHandler(mux, auth, templates)
	authhandler.AttachSignOutHandler(mux, auth, templates)
	authhandler.AttachSignUpHandler(mux, auth, templates)

	// Create a Account Manager
	am := conveyearthgo.NewAccountManager(db)

	// Create a Notification Manager
	var ns conveyearthgo.NotificationSender
	if secure {
		server := os.Getenv("SMTP_SERVER")
		sender := os.Getenv("SMTP_SENDER")
		ns = conveyearthgo.NewSmtpNotificationSender(scheme, host, server, host, sender, templates)
	} else {
		ns = conveytest.NewNotificationSender()
	}
	nm := conveyearthgo.NewNotificationManager(db, ns)

	// Handle Notification Preferences
	handler.AttachNotificationPreferencesHandler(mux, auth, nm, templates)

	uploads, ok := os.LookupEnv("UPLOAD_DIRECTORY")
	if !ok {
		uploads = "uploads"
	}
	if err := os.MkdirAll(uploads, os.ModePerm); err != nil {
		log.Fatal(err)
	}
	log.Println("Uploads Directory:", uploads)

	// Create a Content Manager
	cm := conveyearthgo.NewContentManager(db, filesystem.NewOnDisk(uploads))

	// Handle Content
	handler.AttachContentHandler(mux, cm, fmt.Sprintf("public, immutable, max-age=%d", 60*60*24*7*52)) // 52 week max-age

	digests, ok := os.LookupEnv("DIGEST_DIRECTORY")
	if !ok {
		digests = "digests"
	}
	if err := os.MkdirAll(digests, os.ModePerm); err != nil {
		log.Fatal(err)
	}
	log.Println("Digests Directory:", digests)

	// Handle Digest
	handler.AttachDigestHandler(mux, auth, templates, digests, fmt.Sprintf("public, max-age=%d", 60*60*24*7*52)) // 52 week max-age

	// Handle Account
	handler.AttachAccountHandler(mux, auth, am, nm, templates)

	// Handle Buy Coins
	handler.AttachCoinBuyHandler(mux, auth, am, templates)

	// Create a Stripe Manager
	sm := conveyearthgo.NewStripeManager(db)

	// Handle Stripe
	handler.AttachStripeHandler(mux, auth, sm, templates)

	// Handle Stripe Webhook
	handler.AttachStripeWebhookHandler(mux, am, os.Getenv("STRIPE_WEBHOOK_SECRET_KEY"), host)

	// Handle Conversation
	handler.AttachConversationHandler(mux, auth, cm, templates)

	// Handle Publish
	handler.AttachPublishHandler(mux, auth, am, cm, nm, templates)

	// Handle Reply
	handler.AttachReplyHandler(mux, auth, am, cm, nm, templates)

	// Handle Gift
	handler.AttachGiftHandler(mux, auth, am, cm, nm, templates)

	// Handle Delete
	handler.AttachDeleteHandler(mux, auth, am, cm, templates)

	// Handle Best
	handler.AttachBestHandler(mux, auth, cm, templates, 8, 100)

	// Handle Recent
	handler.AttachRecentHandler(mux, auth, cm, templates, 8, 100)

	// Handle About
	handler.AttachAboutHandler(mux, templates)

	// Handle favicon.ico
	mux.Handle("/favicon.ico", nethandler.Log(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static/convey.svg", http.StatusFound)
	})))

	// Handle robots.txt
	mux.Handle("/robots.txt", nethandler.Log(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static/robots.txt", http.StatusFound)
	})))

	// Handle sitemap.txt
	mux.Handle("/sitemap.txt", nethandler.Log(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static/sitemap.txt", http.StatusFound)
	})))

	// Handle markdown
	mux.Handle("/markdown", nethandler.Log(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := templates.ExecuteTemplate(w, "markdown.go.html", &struct {
			Live bool
		}{
			Live: netgo.IsLive(),
		}); err != nil {
			log.Println(err)
		}
	})))

	// Handle Index
	handler.AttachIndexHandler(mux, auth, cm, templates, digests)

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
		config := &tls.Config{MinVersion: tls.VersionTLS12}
		server := &http.Server{
			Addr:              ":443",
			Handler:           mux,
			TLSConfig:         config,
			ReadTimeout:       time.Hour,
			ReadHeaderTimeout: time.Hour,
			WriteTimeout:      time.Hour,
			IdleTimeout:       time.Hour,
		}
		if err := server.ListenAndServeTLS(path.Join(certificates, "fullchain.pem"), path.Join(certificates, "privkey.pem")); err != nil {
			log.Fatal(err)
		}
	} else {
		// Serve HTTP Requests
		log.Println("HTTP Server Listening on :80")
		if err := http.ListenAndServe(":80", mux); err != nil {
			log.Fatal(err)
		}
	}
}
