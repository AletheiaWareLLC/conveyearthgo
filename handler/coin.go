package handler

import (
	"aletheiaware.com/authgo"
	"aletheiaware.com/authgo/redirect"
	"aletheiaware.com/conveyearthgo"
	"aletheiaware.com/netgo"
	"aletheiaware.com/netgo/handler"
	"fmt"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/checkout/session"
	"github.com/stripe/stripe-go/v72/price"
	"github.com/stripe/stripe-go/v72/product"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

func AttachCoinBuyHandler(m *http.ServeMux, a authgo.Authenticator, am conveyearthgo.AccountManager, ts *template.Template, scheme, domain string) {
	m.Handle("/coin-buy", handler.Log(CoinBuy(a, am, ts, scheme, domain)))
}

func CoinBuy(a authgo.Authenticator, am conveyearthgo.AccountManager, ts *template.Template, scheme, domain string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		account := a.CurrentAccount(w, r)
		if account == nil {
			redirect.SignIn(w, r, r.URL.String())
			return
		}
		data := &CoinBuyData{
			Live:    netgo.IsLive(),
			Account: account,
		}
		bundles := make(map[string]*BundleData)
		products := product.List(&stripe.ProductListParams{})
		for products.Next() {
			product := products.Product()
			d, ok := product.Metadata["domain"]
			if !ok || !strings.Contains(d, domain) {
				continue
			}
			if !product.Active {
				continue
			}
			prices := price.List(&stripe.PriceListParams{
				Product: stripe.String(product.ID),
			})
			for prices.Next() {
				price := prices.Price()
				if !price.Active {
					continue
				}
				size := int64(1)
				s, ok := price.Metadata["bundle_size"]
				if ok {
					if i, err := strconv.ParseInt(s, 10, 64); err != nil {
						log.Println(err)
					} else {
						size = int64(i)
					}
				}
				b := &BundleData{
					Name:      price.Nickname,
					PriceID:   price.ID,
					ProductID: product.ID,
					Size:      size,
				}
				b.Price = conveyearthgo.FormatStripeAmount(price.UnitAmountDecimal, price.Currency)
				bundles[price.ID] = b
				data.Bundles = append(data.Bundles, b)
			}
		}
		sort.Slice(data.Bundles, func(i, j int) bool {
			return data.Bundles[i].Size < data.Bundles[j].Size
		})
		switch r.Method {
		case "GET":
			executeCoinBuyTemplate(w, ts, data)
		case "POST":
			id := strings.TrimSpace(r.FormValue("bundle"))
			data.Bundle = id

			bundle, ok := bundles[id]
			if !ok {
				err := conveyearthgo.ErrBundleUnrecognized
				log.Println(err)
				data.Error = err.Error()
				executeCoinBuyTemplate(w, ts, data)
				return
			}

			params := &stripe.CheckoutSessionParams{
				CustomerEmail: stripe.String(account.Email),
				PaymentMethodTypes: stripe.StringSlice([]string{
					"card",
				}),
				LineItems: []*stripe.CheckoutSessionLineItemParams{
					&stripe.CheckoutSessionLineItemParams{
						Price:    stripe.String(bundle.PriceID),
						Quantity: stripe.Int64(1),
					},
				},
				Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
				SuccessURL: stripe.String(fmt.Sprintf("%s://%s/account", scheme, domain)),
				CancelURL:  stripe.String(fmt.Sprintf("%s://%s/coin-buy", scheme, domain)),
			}
			params.AddMetadata("domain", domain)
			params.AddMetadata("bundle_size", strconv.FormatInt(bundle.Size, 10))
			params.AddMetadata("account_id", strconv.FormatInt(account.ID, 10))
			s, err := session.New(params)
			if err != nil {
				log.Println(err)
				data.Error = err.Error()
				executeCoinBuyTemplate(w, ts, data)
				return
			}
			http.Redirect(w, r, s.URL, http.StatusSeeOther)
		}
	})
}

func executeCoinBuyTemplate(w http.ResponseWriter, ts *template.Template, data *CoinBuyData) {
	if err := ts.ExecuteTemplate(w, "coin-buy.go.html", data); err != nil {
		log.Println(err)
	}
}

type CoinBuyData struct {
	Live    bool
	Error   string
	Account *authgo.Account
	Bundles []*BundleData
	Bundle  string
}

type BundleData struct {
	Name      string
	ProductID string
	PriceID   string
	Price     string
	Size      int64
}
