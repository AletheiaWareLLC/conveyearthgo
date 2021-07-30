package handler

import (
	"aletheiaware.com/conveyearthgo"
	"aletheiaware.com/netgo/handler"
	"encoding/json"
	"fmt"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/webhook"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const MaxBodyBytes = int64(65536)

func AttachStripeWebhookHandler(m *http.ServeMux, am conveyearthgo.AccountManager, webhookSecretKey string, domain string) {
	m.Handle("/stripe-webhook", handler.Log(StripeWebhook(am, webhookSecretKey, domain)))
}

func StripeWebhook(am conveyearthgo.AccountManager, webhookSecretKey string, domain string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, err := ioutil.ReadAll(http.MaxBytesReader(w, r.Body, MaxBodyBytes))
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
