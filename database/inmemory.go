package database

import (
	"aletheiaware.com/authgo"
	"aletheiaware.com/authgo/database"
	"sort"
	"sync"
	"time"
)

func NewInMemory() *InMemory {
	return &InMemory{
		InMemory:                         database.NewInMemory(),
		ConversationId:                   make(map[int64]bool),
		ConversationUser:                 make(map[int64]int64),
		ConversationTopic:                make(map[int64]string),
		ConversationCreated:              make(map[int64]time.Time),
		MessageId:                        make(map[int64]bool),
		MessageUser:                      make(map[int64]int64),
		MessageConversation:              make(map[int64]int64),
		MessageParent:                    make(map[int64]int64),
		MessageCreated:                   make(map[int64]time.Time),
		FileId:                           make(map[int64]bool),
		FileMessage:                      make(map[int64]int64),
		FileHash:                         make(map[int64]string),
		FileMime:                         make(map[int64]string),
		FileCreated:                      make(map[int64]time.Time),
		ChargeId:                         make(map[int64]bool),
		ChargeUser:                       make(map[int64]int64),
		ChargeConversation:               make(map[int64]int64),
		ChargeMessage:                    make(map[int64]int64),
		ChargeAmount:                     make(map[int64]int64),
		ChargeCreated:                    make(map[int64]time.Time),
		YieldId:                          make(map[int64]bool),
		YieldUser:                        make(map[int64]int64),
		YieldConversation:                make(map[int64]int64),
		YieldMessage:                     make(map[int64]int64),
		YieldParent:                      make(map[int64]int64),
		YieldAmount:                      make(map[int64]int64),
		YieldCreated:                     make(map[int64]time.Time),
		PurchaseId:                       make(map[int64]bool),
		PurchaseUser:                     make(map[int64]int64),
		PurchaseStripeSession:            make(map[int64]string),
		PurchaseStripeCustomer:           make(map[int64]string),
		PurchaseStripePaymentIntent:      make(map[int64]string),
		PurchaseStripeCurrency:           make(map[int64]string),
		PurchaseStripeAmount:             make(map[int64]int64),
		PurchaseBundleSize:               make(map[int64]int64),
		PurchaseCreated:                  make(map[int64]time.Time),
		NotificationPreferencesId:        make(map[int64]bool),
		NotificationPreferencesUser:      make(map[int64]int64),
		NotificationPreferencesResponses: make(map[int64]bool),
		NotificationPreferencesMentions:  make(map[int64]bool),
		NotificationPreferencesDigests:   make(map[int64]bool),
		AwardId:                          make(map[int64]bool),
		AwardUser:                        make(map[int64]int64),
		AwardAmount:                      make(map[int64]int64),
	}
}

type InMemory struct {
	sync.RWMutex
	*database.InMemory
	ConversationId                   map[int64]bool
	ConversationUser                 map[int64]int64
	ConversationTopic                map[int64]string
	ConversationCreated              map[int64]time.Time
	MessageId                        map[int64]bool
	MessageUser                      map[int64]int64
	MessageConversation              map[int64]int64
	MessageParent                    map[int64]int64
	MessageCreated                   map[int64]time.Time
	FileId                           map[int64]bool
	FileMessage                      map[int64]int64
	FileHash                         map[int64]string
	FileMime                         map[int64]string
	FileCreated                      map[int64]time.Time
	ChargeId                         map[int64]bool
	ChargeUser                       map[int64]int64
	ChargeConversation               map[int64]int64
	ChargeMessage                    map[int64]int64
	ChargeAmount                     map[int64]int64
	ChargeCreated                    map[int64]time.Time
	YieldId                          map[int64]bool
	YieldUser                        map[int64]int64
	YieldConversation                map[int64]int64
	YieldMessage                     map[int64]int64
	YieldParent                      map[int64]int64
	YieldAmount                      map[int64]int64
	YieldCreated                     map[int64]time.Time
	PurchaseId                       map[int64]bool
	PurchaseUser                     map[int64]int64
	PurchaseStripeSession            map[int64]string
	PurchaseStripeCustomer           map[int64]string
	PurchaseStripePaymentIntent      map[int64]string
	PurchaseStripeCurrency           map[int64]string
	PurchaseStripeAmount             map[int64]int64
	PurchaseBundleSize               map[int64]int64
	PurchaseCreated                  map[int64]time.Time
	NotificationPreferencesId        map[int64]bool
	NotificationPreferencesUser      map[int64]int64
	NotificationPreferencesResponses map[int64]bool
	NotificationPreferencesMentions  map[int64]bool
	NotificationPreferencesDigests   map[int64]bool
	AwardId                          map[int64]bool
	AwardUser                        map[int64]int64
	AwardAmount                      map[int64]int64
}

func (db *InMemory) CreateConversation(user int64, topic string, created time.Time) (int64, error) {
	db.Lock()
	defer db.Unlock()
	id := database.NextId()
	db.ConversationId[id] = true
	db.ConversationUser[id] = user
	db.ConversationTopic[id] = topic
	db.ConversationCreated[id] = created
	return id, nil
}

func (db *InMemory) SelectConversation(id int64) (*authgo.Account, string, time.Time, error) {
	db.Lock()
	defer db.Unlock()
	if _, ok := db.ConversationId[id]; !ok {
		return nil, "", time.Time{}, database.ErrNoSuchRecord
	}
	user := db.ConversationUser[id]
	username := db.username(user)
	email := db.AccountEmail[username]
	topic := db.ConversationTopic[id]
	created := db.ConversationCreated[id]
	return &authgo.Account{
		ID:       user,
		Username: username,
		Email:    email,
	}, topic, created, nil
}

func (db *InMemory) SelectBestConversations(callback func(int64, *authgo.Account, string, time.Time, int64, int64) error, since time.Time, limit int64) error {
	db.Lock()
	defer db.Unlock()
	costs := make(map[int64]int64)
	yields := make(map[int64]int64)
	var results []int64
	for cid := range db.ConversationId {
		if db.ConversationCreated[cid].Before(since) {
			continue
		}
		results = append(results, cid)
		for mid := range db.MessageId {
			if db.MessageConversation[mid] != cid {
				continue
			}
			costs[cid] = db.cost(mid)
			yields[cid] = db.yield(mid)
		}
	}
	sort.Slice(results, func(a, b int) bool {
		return yields[results[a]] < yields[results[b]]
	})
	count := int64(len(results))
	for i := int64(0); i < limit && i < count; i++ {
		cid := results[i]
		user := db.ConversationUser[cid]
		username := db.username(user)
		email := db.AccountEmail[username]
		topic := db.ConversationTopic[cid]
		created := db.ConversationCreated[cid]
		cost := costs[cid]
		yield := yields[cid]
		if err := callback(cid, &authgo.Account{
			ID:       user,
			Username: username,
			Email:    email,
		}, topic, created, cost, yield); err != nil {
			return err
		}
	}
	return nil
}

func (db *InMemory) SelectRecentConversations(callback func(int64, *authgo.Account, string, time.Time, int64, int64) error, limit int64) error {
	db.Lock()
	defer db.Unlock()
	costs := make(map[int64]int64)
	yields := make(map[int64]int64)
	var results []int64
	for cid := range db.ConversationId {
		results = append(results, cid)
		for mid := range db.MessageId {
			if db.MessageConversation[mid] != cid {
				continue
			}
			costs[cid] = db.cost(mid)
			yields[cid] = db.yield(mid)
		}
	}
	sort.Slice(results, func(a, b int) bool {
		return db.ConversationCreated[results[a]].Before(db.ConversationCreated[results[b]])
	})
	count := int64(len(results))
	for i := int64(0); i < limit && i < count; i++ {
		cid := results[i]
		user := db.ConversationUser[cid]
		username := db.username(user)
		email := db.AccountEmail[username]
		topic := db.ConversationTopic[cid]
		created := db.ConversationCreated[cid]
		cost := costs[cid]
		yield := yields[cid]
		if err := callback(cid, &authgo.Account{
			ID:       user,
			Username: username,
			Email:    email,
		}, topic, created, cost, yield); err != nil {
			return err
		}
	}
	return nil
}

func (db *InMemory) CreateMessage(user, conversation, parent int64, created time.Time) (int64, error) {
	db.Lock()
	defer db.Unlock()
	id := database.NextId()
	db.MessageId[id] = true
	db.MessageUser[id] = user
	db.MessageConversation[id] = conversation
	db.MessageParent[id] = parent
	db.MessageCreated[id] = created
	return id, nil
}

func (db *InMemory) SelectMessage(id int64) (*authgo.Account, int64, int64, time.Time, int64, int64, error) {
	db.Lock()
	defer db.Unlock()
	if _, ok := db.MessageId[id]; !ok {
		return nil, 0, 0, time.Time{}, 0, 0, database.ErrNoSuchRecord
	}
	user := db.MessageUser[id]
	username := db.username(user)
	email := db.AccountEmail[username]
	conversation := db.MessageConversation[id]
	parent := db.MessageParent[id]
	created := db.MessageCreated[id]
	cost := db.cost(id)
	yield := db.yield(id)
	return &authgo.Account{
		ID:       user,
		Username: username,
		Email:    email,
	}, conversation, parent, created, cost, yield, nil
}

func (db *InMemory) SelectMessages(conversation int64, callback func(int64, *authgo.Account, int64, time.Time, int64, int64) error) error {
	db.Lock()
	defer db.Unlock()
	for id := range db.MessageId {
		if db.MessageConversation[id] != conversation {
			continue
		}
		user := db.MessageUser[id]
		username := db.username(user)
		email := db.AccountEmail[username]
		parent := db.MessageParent[id]
		created := db.MessageCreated[id]
		cost := db.cost(id)
		yield := db.yield(id)
		if err := callback(id, &authgo.Account{
			ID:       user,
			Username: username,
			Email:    email,
		}, parent, created, cost, yield); err != nil {
			return err
		}
	}
	return nil
}

func (db *InMemory) SelectMessageParent(id int64) (int64, error) {
	db.Lock()
	defer db.Unlock()
	if _, ok := db.MessageId[id]; !ok {
		return 0, database.ErrNoSuchRecord
	}
	return db.MessageParent[id], nil
}

func (db *InMemory) CreateFile(message int64, hash, mime string, created time.Time) (int64, error) {
	db.Lock()
	defer db.Unlock()
	id := database.NextId()
	db.FileId[id] = true
	db.FileMessage[id] = message
	db.FileHash[id] = hash
	db.FileMime[id] = mime
	db.FileCreated[id] = created
	return id, nil
}

func (db *InMemory) SelectFile(id int64) (int64, string, string, time.Time, error) {
	db.Lock()
	defer db.Unlock()
	if _, ok := db.FileId[id]; !ok {
		return 0, "", "", time.Time{}, database.ErrNoSuchRecord
	}
	message := db.FileMessage[id]
	hash := db.FileHash[id]
	mime := db.FileMime[id]
	created := db.FileCreated[id]
	return message, hash, mime, created, nil
}

func (db *InMemory) SelectFiles(message int64, callback func(int64, string, string, time.Time) error) error {
	db.Lock()
	defer db.Unlock()
	for id := range db.FileId {
		if db.FileMessage[id] != message {
			continue
		}
		hash := db.FileHash[id]
		mime := db.FileMime[id]
		created := db.FileCreated[id]
		if err := callback(id, hash, mime, created); err != nil {
			return err
		}
	}
	return nil
}

func (db *InMemory) CreateCharge(user, conversation, message, amount int64, created time.Time) (int64, error) {
	db.Lock()
	defer db.Unlock()
	id := database.NextId()
	db.ChargeId[id] = true
	db.ChargeUser[id] = user
	db.ChargeConversation[id] = conversation
	db.ChargeMessage[id] = message
	db.ChargeAmount[id] = amount
	db.ChargeCreated[id] = created
	return id, nil
}

func (db *InMemory) SelectCharges(user int64) (int64, error) {
	db.Lock()
	defer db.Unlock()
	var charges int64
	for cid := range db.ChargeId {
		if db.ChargeUser[cid] != user {
			continue
		}
		charges += db.ChargeAmount[cid]
	}
	return charges, nil
}

func (db *InMemory) CreateYield(user, conversation, message, parent, amount int64, created time.Time) (int64, error) {
	db.Lock()
	defer db.Unlock()
	id := database.NextId()
	db.YieldId[id] = true
	db.YieldUser[id] = user
	db.YieldConversation[id] = conversation
	db.YieldMessage[id] = message
	db.YieldParent[id] = parent
	db.YieldAmount[id] = amount
	db.YieldCreated[id] = created
	return id, nil
}

func (db *InMemory) SelectYields(user int64) (int64, error) {
	db.Lock()
	defer db.Unlock()
	var yields int64
	for yid := range db.YieldId {
		if db.YieldUser[yid] != user {
			continue
		}
		yields += db.YieldAmount[yid]
	}
	return yields, nil
}

func (db *InMemory) CreatePurchase(user int64, stripeSession, stripeCustomer, stripePaymentIntent, stripeCurrency string, stripeAmount, bundle_size int64, created time.Time) (int64, error) {
	db.Lock()
	defer db.Unlock()
	id := database.NextId()
	db.PurchaseId[id] = true
	db.PurchaseUser[id] = user
	db.PurchaseStripeSession[id] = stripeSession
	db.PurchaseStripeCustomer[id] = stripeCustomer
	db.PurchaseStripePaymentIntent[id] = stripePaymentIntent
	db.PurchaseStripeCurrency[id] = stripeCurrency
	db.PurchaseStripeAmount[id] = stripeAmount
	db.PurchaseBundleSize[id] = bundle_size
	db.PurchaseCreated[id] = created
	return id, nil
}

func (db *InMemory) SelectPurchases(user int64) (int64, error) {
	db.Lock()
	defer db.Unlock()
	var purchases int64
	for pid := range db.PurchaseId {
		if db.PurchaseUser[pid] != user {
			continue
		}
		purchases += db.PurchaseBundleSize[pid]
	}
	return purchases, nil
}

func (db *InMemory) UpdateNotificationPreferences(id, user int64, responses, mentions, digests bool) (int64, error) {
	db.NotificationPreferencesId[id] = true
	db.NotificationPreferencesUser[id] = user
	db.NotificationPreferencesResponses[id] = responses
	db.NotificationPreferencesMentions[id] = mentions
	db.NotificationPreferencesDigests[id] = digests
	return 1, nil
}

func (db *InMemory) SelectNotificationPreferences(user int64) (int64, bool, bool, bool, error) {
	var id int64
	responses := true
	mentions := true
	digests := true
	for i := range db.NotificationPreferencesId {
		if db.NotificationPreferencesUser[i] == user {
			id = i
			responses = db.NotificationPreferencesResponses[i]
			mentions = db.NotificationPreferencesMentions[i]
			digests = db.NotificationPreferencesDigests[i]
		}
	}
	return id, responses, mentions, digests, nil
}

func (db *InMemory) SelectAwards(user int64) (int64, error) {
	db.Lock()
	defer db.Unlock()
	var awards int64
	for pid := range db.AwardId {
		if db.AwardUser[pid] != user {
			continue
		}
		awards += db.AwardAmount[pid]
	}
	return awards, nil
}

func (db *InMemory) username(id int64) string {
	for k, v := range db.AccountId {
		if v == id {
			return k
		}
	}
	return ""
}

func (db *InMemory) cost(id int64) (cost int64) {
	for cid := range db.ChargeId {
		if db.ChargeMessage[cid] != id {
			continue
		}
		cost += db.ChargeAmount[cid]
	}
	return
}

func (db *InMemory) yield(id int64) (yield int64) {
	for cid := range db.YieldId {
		if db.YieldMessage[cid] != id {
			continue
		}
		yield += db.YieldAmount[cid]
	}
	return
}
