package handler

import (
	"aletheiaware.com/authgo"
	"aletheiaware.com/authgo/redirect"
	"aletheiaware.com/conveyearthgo"
	"aletheiaware.com/netgo"
	"aletheiaware.com/netgo/handler"
	"encoding/json"
	"fmt"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/account"
	"github.com/stripe/stripe-go/v72/accountlink"
	"github.com/stripe/stripe-go/v72/balance"
	"github.com/stripe/stripe-go/v72/loginlink"
	"github.com/stripe/stripe-go/v72/webhook"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const MaxBodyBytes = int64(65536)

func AttachStripeHandler(m *http.ServeMux, a authgo.Authenticator, sm conveyearthgo.StripeManager, ts *template.Template) {
	m.Handle("/stripe", handler.Log(handler.Compress(Stripe(a, sm, ts))))
}

func Stripe(a authgo.Authenticator, sm conveyearthgo.StripeManager, ts *template.Template) http.Handler {
	scheme := conveyearthgo.Scheme()
	domain := conveyearthgo.Host()
	url := fmt.Sprintf("%s://%s/stripe", scheme, domain)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		acc := a.CurrentAccount(w, r)
		if acc == nil {
			redirect.SignIn(w, r, r.URL.String())
			return
		}
		data := &StripeData{
			Live:    netgo.IsLive(),
			Account: acc,
		}
		sa, err := sm.StripeAccount(acc)
		if err != nil {
			log.Println(err)
		}
		if sa != nil {
			data.StripeAccount = sa

			params := &stripe.BalanceParams{}
			params.SetStripeAccount(sa.ID)
			bal, err := balance.Get(params)
			if err != nil {
				log.Println(err)
			} else {
				for _, a := range bal.Available {
					s := conveyearthgo.FormatStripeAmount(float64(a.Value), a.Currency)
					if s != "" {
						data.StripeAvailableBalance = append(data.StripeAvailableBalance, s)
					}
				}
				for _, a := range bal.Pending {
					s := conveyearthgo.FormatStripeAmount(float64(a.Value), a.Currency)
					if s != "" {
						data.StripePendingBalance = append(data.StripePendingBalance, s)
					}
				}
			}

			if sa.ChargesEnabled && sa.PayoutsEnabled {
				ll, err := loginlink.New(&stripe.LoginLinkParams{
					Account: stripe.String(sa.ID),
				})
				if err != nil {
					log.Println(err)
				} else {
					data.StripeLoginLink = ll.URL
				}
			} else if !sa.DetailsSubmitted {
				al, err := accountlink.New(&stripe.AccountLinkParams{
					Account:    stripe.String(sa.ID),
					RefreshURL: stripe.String(url),
					ReturnURL:  stripe.String(url),
					Type:       stripe.String("account_onboarding"),
				})
				if err != nil {
					log.Println(err)
				} else {
					data.StripeOnboardingLink = al.URL
				}
			}
		}
		switch r.Method {
		case "GET":
			executeStripeTemplate(w, ts, data)
		case "POST":
			if sa != nil {
				http.Redirect(w, r, "/stripe", http.StatusFound)
				return
			}
			params := &stripe.AccountParams{
				Capabilities: &stripe.AccountCapabilitiesParams{
					Transfers: &stripe.AccountCapabilitiesTransfersParams{
						Requested: stripe.Bool(true),
					},
				},
				Country: stripe.String(r.FormValue("country")),
				Email:   stripe.String(acc.Email),
				Type:    stripe.String("express"),
				TOSAcceptance: &stripe.AccountTOSAcceptanceParams{
					ServiceAgreement: stripe.String("recipient"),
				},
			}
			params.AddMetadata("domain", domain)
			params.AddMetadata("account_id", strconv.FormatInt(acc.ID, 10))
			a, err := account.New(params)
			if err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			if err := sm.NewStripeAccount(acc, a); err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			al, err := accountlink.New(&stripe.AccountLinkParams{
				Account:    stripe.String(a.ID),
				RefreshURL: stripe.String(url),
				ReturnURL:  stripe.String(url),
				Type:       stripe.String("account_onboarding"),
			})
			if err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			http.Redirect(w, r, al.URL, http.StatusFound)
		}
	})
}

func executeStripeTemplate(w http.ResponseWriter, ts *template.Template, data *StripeData) {
	if err := ts.ExecuteTemplate(w, "stripe.go.html", data); err != nil {
		log.Println(err)
	}
}

type StripeData struct {
	Live                   bool
	Error                  string
	Account                *authgo.Account
	StripeAccount          *stripe.Account
	StripeAvailableBalance []string
	StripePendingBalance   []string
	StripeOnboardingLink   string
	StripeLoginLink        string
}

func AttachStripeWebhookHandler(m *http.ServeMux, am conveyearthgo.AccountManager, webhookSecretKey string, domain string) {
	m.Handle("/stripe-webhook", handler.Log(StripeWebhook(am, webhookSecretKey, domain)))
}

func StripeWebhook(am conveyearthgo.AccountManager, webhookSecretKey string, domain string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, err := io.ReadAll(http.MaxBytesReader(w, r.Body, MaxBodyBytes))
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		event, err := webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), webhookSecretKey)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		fmt.Printf("Stripe Event: %+v\n", event)

		// Unmarshal the event data into an appropriate struct depending on its Type
		switch event.Type {
		case "checkout.session.completed":
			var session stripe.CheckoutSession
			if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			fmt.Printf("Checkout Session Completed: %+v\n", session)
			// TODO
			// - Send the customer a receipt email.

			// TODO check payment status == "paid"

			d, ok := session.Metadata["domain"]
			if !ok || !strings.Contains(d, domain) {
				log.Println("Incorrect Domain")
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			user := int64(0)
			u, ok := session.Metadata["account_id"]
			if !ok {
				log.Println("Missing Account ID")
				w.WriteHeader(http.StatusBadRequest)
				return
			} else {
				if i, err := strconv.ParseInt(u, 10, 64); err != nil {
					log.Println(err)
				} else {
					user = int64(i)
				}
			}

			size := int64(1)
			s, ok := session.Metadata["bundle_size"]
			if !ok {
				log.Println("Missing Bundle Size")
				w.WriteHeader(http.StatusBadRequest)
				return
			} else {
				if i, err := strconv.ParseInt(s, 10, 64); err != nil {
					log.Println(err)
				} else {
					size = int64(i)
				}
			}
			if err := am.NewPurchase(user, session.ID, session.Customer.ID, session.PaymentIntent.ID, string(session.Currency), session.AmountTotal, size); err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		default:
			log.Println("Unhandled event type:", event.Type)
		}

		w.WriteHeader(http.StatusOK)
	})
}
