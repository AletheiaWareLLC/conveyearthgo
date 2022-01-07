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
		ConversationDeleted:              make(map[int64]time.Time),
		MessageId:                        make(map[int64]bool),
		MessageUser:                      make(map[int64]int64),
		MessageConversation:              make(map[int64]int64),
		MessageParent:                    make(map[int64]int64),
		MessageCreated:                   make(map[int64]time.Time),
		MessageDeleted:                   make(map[int64]time.Time),
		FileId:                           make(map[int64]bool),
		FileMessage:                      make(map[int64]int64),
		FileHash:                         make(map[int64]string),
		FileMime:                         make(map[int64]string),
		FileCreated:                      make(map[int64]time.Time),
		FileDeleted:                      make(map[int64]time.Time),
		ChargeId:                         make(map[int64]bool),
		ChargeUser:                       make(map[int64]int64),
		ChargeConversation:               make(map[int64]int64),
		ChargeMessage:                    make(map[int64]int64),
		ChargeAmount:                     make(map[int64]int64),
		ChargeCreated:                    make(map[int64]time.Time),
		ChargeDeleted:                    make(map[int64]time.Time),
		YieldId:                          make(map[int64]bool),
		YieldUser:                        make(map[int64]int64),
		YieldConversation:                make(map[int64]int64),
		YieldMessage:                     make(map[int64]int64),
		YieldParent:                      make(map[int64]int64),
		YieldAmount:                      make(map[int64]int64),
		YieldCreated:                     make(map[int64]time.Time),
		YieldDeleted:                     make(map[int64]time.Time),
		PurchaseId:                       make(map[int64]bool),
		PurchaseUser:                     make(map[int64]int64),
		PurchaseStripeSession:            make(map[int64]string),
		PurchaseStripeCustomer:           make(map[int64]string),
		PurchaseStripePaymentIntent:      make(map[int64]string),
		PurchaseStripeCurrency:           make(map[int64]string),
		PurchaseStripeAmount:             make(map[int64]int64),
		PurchaseBundleSize:               make(map[int64]int64),
		PurchaseCreated:                  make(map[int64]time.Time),
		PurchaseDeleted:                  make(map[int64]time.Time),
		NotificationPreferencesId:        make(map[int64]bool),
		NotificationPreferencesUser:      make(map[int64]int64),
		NotificationPreferencesResponses: make(map[int64]bool),
		NotificationPreferencesMentions:  make(map[int64]bool),
		NotificationPreferencesGifts:     make(map[int64]bool),
		NotificationPreferencesDigests:   make(map[int64]bool),
		AwardId:                          make(map[int64]bool),
		AwardUser:                        make(map[int64]int64),
		AwardAmount:                      make(map[int64]int64),
		AwardCreated:                     make(map[int64]time.Time),
		AwardDeleted:                     make(map[int64]time.Time),
		StripeAccountId:                  make(map[int64]bool),
		StripeAccountUser:                make(map[int64]int64),
		StripeAccountIdentity:            make(map[int64]string),
		StripeAccountCreated:             make(map[int64]time.Time),
		StripeAccountDeleted:             make(map[int64]time.Time),
		GiftId:                           make(map[int64]bool),
		GiftUser:                         make(map[int64]int64),
		GiftConversation:                 make(map[int64]int64),
		GiftMessage:                      make(map[int64]int64),
		GiftAmount:                       make(map[int64]int64),
		GiftCreated:                      make(map[int64]time.Time),
		GiftDeleted:                      make(map[int64]time.Time),
	}
}

type InMemory struct {
	sync.RWMutex
	*database.InMemory
	ConversationId                   map[int64]bool
	ConversationUser                 map[int64]int64
	ConversationTopic                map[int64]string
	ConversationCreated              map[int64]time.Time
	ConversationDeleted              map[int64]time.Time
	MessageId                        map[int64]bool
	MessageUser                      map[int64]int64
	MessageConversation              map[int64]int64
	MessageParent                    map[int64]int64
	MessageCreated                   map[int64]time.Time
	MessageDeleted                   map[int64]time.Time
	FileId                           map[int64]bool
	FileMessage                      map[int64]int64
	FileHash                         map[int64]string
	FileMime                         map[int64]string
	FileCreated                      map[int64]time.Time
	FileDeleted                      map[int64]time.Time
	ChargeId                         map[int64]bool
	ChargeUser                       map[int64]int64
	ChargeConversation               map[int64]int64
	ChargeMessage                    map[int64]int64
	ChargeAmount                     map[int64]int64
	ChargeCreated                    map[int64]time.Time
	ChargeDeleted                    map[int64]time.Time
	YieldId                          map[int64]bool
	YieldUser                        map[int64]int64
	YieldConversation                map[int64]int64
	YieldMessage                     map[int64]int64
	YieldParent                      map[int64]int64
	YieldAmount                      map[int64]int64
	YieldCreated                     map[int64]time.Time
	YieldDeleted                     map[int64]time.Time
	PurchaseId                       map[int64]bool
	PurchaseUser                     map[int64]int64
	PurchaseStripeSession            map[int64]string
	PurchaseStripeCustomer           map[int64]string
	PurchaseStripePaymentIntent      map[int64]string
	PurchaseStripeCurrency           map[int64]string
	PurchaseStripeAmount             map[int64]int64
	PurchaseBundleSize               map[int64]int64
	PurchaseCreated                  map[int64]time.Time
	PurchaseDeleted                  map[int64]time.Time
	NotificationPreferencesId        map[int64]bool
	NotificationPreferencesUser      map[int64]int64
	NotificationPreferencesResponses map[int64]bool
	NotificationPreferencesMentions  map[int64]bool
	NotificationPreferencesGifts     map[int64]bool
	NotificationPreferencesDigests   map[int64]bool
	AwardId                          map[int64]bool
	AwardUser                        map[int64]int64
	AwardAmount                      map[int64]int64
	AwardCreated                     map[int64]time.Time
	AwardDeleted                     map[int64]time.Time
	StripeAccountId                  map[int64]bool
	StripeAccountUser                map[int64]int64
	StripeAccountIdentity            map[int64]string
	StripeAccountCreated             map[int64]time.Time
	StripeAccountDeleted             map[int64]time.Time
	GiftId                           map[int64]bool
	GiftUser                         map[int64]int64
	GiftConversation                 map[int64]int64
	GiftMessage                      map[int64]int64
	GiftAmount                       map[int64]int64
	GiftCreated                      map[int64]time.Time
	GiftDeleted                      map[int64]time.Time
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

func (db *InMemory) DeleteConversation(user, id int64, deleted time.Time) (int64, error) {
	if db.ConversationUser[id] != user {
		return 0, nil
	}
	db.ConversationDeleted[id] = deleted
	return 1, nil
}

func (db *InMemory) SelectConversation(id int64) (*authgo.Account, string, time.Time, error) {
	db.Lock()
	defer db.Unlock()
	if _, ok := db.ConversationId[id]; !ok {
		return nil, "", time.Time{}, database.ErrNoSuchRecord
	}
	if _, ok := db.ConversationDeleted[id]; ok {
		return nil, "", time.Time{}, database.ErrNoSuchRecord
	}
	user := db.ConversationUser[id]
	username := db.username(user)
	if _, ok := db.AccountDeleted[username]; ok {
		return nil, "", time.Time{}, database.ErrNoSuchRecord
	}
	email := db.AccountEmail[username]
	joined := db.AccountCreated[username]
	topic := db.ConversationTopic[id]
	created := db.ConversationCreated[id]
	return &authgo.Account{
		ID:       user,
		Username: username,
		Email:    email,
		Created:  joined,
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
		if _, ok := db.ConversationDeleted[cid]; ok {
			continue
		}
		for mid := range db.MessageId {
			if db.MessageConversation[mid] != cid || db.MessageParent[mid] != 0 {
				continue
			}
			if _, ok := db.MessageDeleted[mid]; ok {
				continue
			}
			costs[cid] = db.cost(mid)
			yields[cid] = db.yield(mid)
		}
		if y, ok := yields[cid]; ok && y > 0 {
			results = append(results, cid)
		}
	}
	// Sort results by decending yields
	sort.Slice(results, func(a, b int) bool {
		return yields[results[a]] > yields[results[b]]
	})
	count := int64(len(results))
	for i := int64(0); i < limit && i < count; i++ {
		cid := results[i]
		user := db.ConversationUser[cid]
		username := db.username(user)
		if _, ok := db.AccountDeleted[username]; ok {
			// This can result in fewer than limit results returned
			continue
		}
		email := db.AccountEmail[username]
		joined := db.AccountCreated[username]
		topic := db.ConversationTopic[cid]
		created := db.ConversationCreated[cid]
		cost := costs[cid]
		yield := yields[cid]
		if err := callback(cid, &authgo.Account{
			ID:       user,
			Username: username,
			Email:    email,
			Created:  joined,
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
		if _, ok := db.ConversationDeleted[cid]; ok {
			continue
		}
		results = append(results, cid)
		for mid := range db.MessageId {
			if db.MessageConversation[mid] != cid || db.MessageParent[mid] != 0 {
				continue
			}
			if _, ok := db.MessageDeleted[mid]; ok {
				continue
			}
			costs[cid] = db.cost(mid)
			yields[cid] = db.yield(mid)
		}
	}
	// Sort results by decending creation time
	sort.Slice(results, func(a, b int) bool {
		return db.ConversationCreated[results[a]].After(db.ConversationCreated[results[b]])
	})
	count := int64(len(results))
	for i := int64(0); i < limit && i < count; i++ {
		cid := results[i]
		user := db.ConversationUser[cid]
		username := db.username(user)
		if _, ok := db.AccountDeleted[username]; ok {
			// This can result in fewer than limit results returned
			continue
		}
		email := db.AccountEmail[username]
		joined := db.AccountCreated[username]
		topic := db.ConversationTopic[cid]
		created := db.ConversationCreated[cid]
		cost := costs[cid]
		yield := yields[cid]
		if err := callback(cid, &authgo.Account{
			ID:       user,
			Username: username,
			Email:    email,
			Created:  joined,
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

func (db *InMemory) DeleteMessage(user, id int64, deleted time.Time) (int64, error) {
	if db.MessageUser[id] != user {
		return 0, nil
	}

	// Don't delete message if it has received any replies
	for mid := range db.MessageId {
		if db.MessageParent[mid] != id {
			continue
		}
		if _, ok := db.MessageDeleted[mid]; ok {
			continue
		}
		return 0, nil
	}

	// Don't delete message if it has received any gifts
	for gid := range db.GiftId {
		if db.GiftMessage[gid] != id {
			continue
		}
		if _, ok := db.GiftDeleted[gid]; ok {
			continue
		}
		return 0, nil
	}

	db.MessageDeleted[id] = deleted
	for f := range db.FileId {
		if db.FileMessage[f] == id {
			db.FileDeleted[f] = deleted
		}
	}
	for c := range db.ChargeId {
		if db.ChargeMessage[c] == id {
			db.ChargeDeleted[c] = deleted
		}
	}
	for y := range db.YieldId {
		if db.YieldMessage[y] == id {
			db.YieldDeleted[y] = deleted
		}
	}
	return 1, nil
}

func (db *InMemory) SelectMessage(id int64) (*authgo.Account, int64, int64, time.Time, int64, int64, error) {
	db.Lock()
	defer db.Unlock()
	if _, ok := db.MessageId[id]; !ok {
		return nil, 0, 0, time.Time{}, 0, 0, database.ErrNoSuchRecord
	}
	if _, ok := db.MessageDeleted[id]; ok {
		return nil, 0, 0, time.Time{}, 0, 0, database.ErrNoSuchRecord
	}
	user := db.MessageUser[id]
	username := db.username(user)
	if _, ok := db.AccountDeleted[username]; ok {
		return nil, 0, 0, time.Time{}, 0, 0, database.ErrNoSuchRecord
	}
	email := db.AccountEmail[username]
	joined := db.AccountCreated[username]
	conversation := db.MessageConversation[id]
	parent := db.MessageParent[id]
	created := db.MessageCreated[id]
	cost := db.cost(id)
	yield := db.yield(id)
	return &authgo.Account{
		ID:       user,
		Username: username,
		Email:    email,
		Created:  joined,
	}, conversation, parent, created, cost, yield, nil
}

func (db *InMemory) SelectMessages(conversation int64, callback func(int64, *authgo.Account, int64, time.Time, int64, int64) error) error {
	db.Lock()
	defer db.Unlock()
	for id := range db.MessageId {
		if db.MessageConversation[id] != conversation {
			continue
		}
		if _, ok := db.MessageDeleted[id]; ok {
			continue
		}
		user := db.MessageUser[id]
		username := db.username(user)
		if _, ok := db.AccountDeleted[username]; ok {
			continue
		}
		email := db.AccountEmail[username]
		joined := db.AccountCreated[username]
		parent := db.MessageParent[id]
		created := db.MessageCreated[id]
		cost := db.cost(id)
		yield := db.yield(id)
		if err := callback(id, &authgo.Account{
			ID:       user,
			Username: username,
			Email:    email,
			Created:  joined,
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
	if _, ok := db.MessageDeleted[id]; ok {
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
	if _, ok := db.FileDeleted[id]; ok {
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
		if _, ok := db.FileDeleted[id]; ok {
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

func (db *InMemory) SelectChargesForUser(user int64) (int64, error) {
	db.Lock()
	defer db.Unlock()
	var charges int64
	for cid := range db.ChargeId {
		if db.ChargeUser[cid] != user {
			continue
		}
		if _, ok := db.ChargeDeleted[cid]; ok {
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

func (db *InMemory) SelectYieldsForUser(user int64) (int64, error) {
	db.Lock()
	defer db.Unlock()
	var yields int64
	for mid := range db.MessageId {
		if db.MessageUser[mid] != user {
			continue
		}
		if _, ok := db.MessageDeleted[mid]; ok {
			continue
		}
		for yid := range db.YieldId {
			if db.YieldParent[yid] != mid {
				continue
			}
			if _, ok := db.YieldDeleted[yid]; ok {
				continue
			}
			yields += db.YieldAmount[yid]
		}
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

func (db *InMemory) SelectPurchasesForUser(user int64) (int64, error) {
	db.Lock()
	defer db.Unlock()
	var purchases int64
	for pid := range db.PurchaseId {
		if db.PurchaseUser[pid] != user {
			continue
		}
		if _, ok := db.PurchaseDeleted[pid]; ok {
			continue
		}
		purchases += db.PurchaseBundleSize[pid]
	}
	return purchases, nil
}

func (db *InMemory) UpdateNotificationPreferences(id, user int64, responses, mentions, gifts, digests bool) (int64, error) {
	db.NotificationPreferencesId[id] = true
	db.NotificationPreferencesUser[id] = user
	db.NotificationPreferencesResponses[id] = responses
	db.NotificationPreferencesMentions[id] = mentions
	db.NotificationPreferencesGifts[id] = gifts
	db.NotificationPreferencesDigests[id] = digests
	return 1, nil
}

func (db *InMemory) SelectNotificationPreferences(user int64) (int64, bool, bool, bool, bool, error) {
	var id int64
	responses := true
	mentions := true
	gifts := true
	digests := true
	for i := range db.NotificationPreferencesId {
		if db.NotificationPreferencesUser[i] == user {
			id = i
			responses = db.NotificationPreferencesResponses[i]
			mentions = db.NotificationPreferencesMentions[i]
			gifts = db.NotificationPreferencesGifts[i]
			digests = db.NotificationPreferencesDigests[i]
		}
	}
	return id, responses, mentions, gifts, digests, nil
}

func (db *InMemory) SelectAwardsForUser(user int64) (int64, error) {
	db.Lock()
	defer db.Unlock()
	var awards int64
	for aid := range db.AwardId {
		if db.AwardUser[aid] != user {
			continue
		}
		if _, ok := db.AwardDeleted[aid]; ok {
			continue
		}
		awards += db.AwardAmount[aid]
	}
	return awards, nil
}

func (db *InMemory) CreateStripeAccount(user int64, identity string, created time.Time) (int64, error) {
	db.Lock()
	defer db.Unlock()
	id := database.NextId()
	db.StripeAccountId[id] = true
	db.StripeAccountUser[id] = user
	db.StripeAccountIdentity[id] = identity
	db.StripeAccountCreated[id] = created
	return id, nil
}

func (db *InMemory) SelectStripeAccount(user int64) (string, time.Time, error) {
	db.Lock()
	defer db.Unlock()
	for sid := range db.StripeAccountId {
		if _, ok := db.StripeAccountDeleted[sid]; ok {
			continue
		}
		if db.StripeAccountUser[sid] == user {
			return db.StripeAccountIdentity[sid], db.StripeAccountCreated[sid], nil
		}
	}
	return "", time.Time{}, nil
}

func (db *InMemory) CreateGift(user, conversation, message, amount int64, created time.Time) (int64, error) {
	db.Lock()
	defer db.Unlock()
	id := database.NextId()
	db.GiftId[id] = true
	db.GiftUser[id] = user
	db.GiftConversation[id] = conversation
	db.GiftMessage[id] = message
	db.GiftAmount[id] = amount
	db.GiftCreated[id] = created
	return id, nil
}

func (db *InMemory) DeleteGift(user, id int64, deleted time.Time) (int64, error) {
	if db.GiftUser[id] != user {
		return 0, nil
	}
	db.GiftDeleted[id] = deleted
	return 1, nil
}

func (db *InMemory) SelectGift(id int64) (int64, int64, *authgo.Account, int64, time.Time, error) {
	db.Lock()
	defer db.Unlock()
	if _, ok := db.GiftId[id]; !ok {
		return 0, 0, nil, 0, time.Time{}, database.ErrNoSuchRecord
	}
	if _, ok := db.GiftDeleted[id]; ok {
		return 0, 0, nil, 0, time.Time{}, database.ErrNoSuchRecord
	}
	conversation := db.GiftConversation[id]
	message := db.GiftMessage[id]
	user := db.GiftUser[id]
	username := db.username(user)
	if _, ok := db.AccountDeleted[username]; ok {
		return 0, 0, nil, 0, time.Time{}, database.ErrNoSuchRecord
	}
	email := db.AccountEmail[username]
	joined := db.AccountCreated[username]
	amount := db.GiftAmount[id]
	created := db.GiftCreated[id]
	return conversation, message, &authgo.Account{
		ID:       user,
		Username: username,
		Email:    email,
		Created:  joined,
	}, amount, created, nil
}

func (db *InMemory) SelectGifts(conversation, message int64, callback func(int64, int64, int64, *authgo.Account, int64, time.Time) error) error {
	db.Lock()
	defer db.Unlock()
	for id := range db.GiftId {
		if conversation > 0 && db.GiftConversation[id] != conversation {
			continue
		}
		if message > 0 && db.GiftMessage[id] != message {
			continue
		}
		if _, ok := db.GiftDeleted[id]; ok {
			continue
		}
		conversation = db.GiftConversation[id]
		message = db.GiftMessage[id]
		user := db.GiftUser[id]
		username := db.username(user)
		if _, ok := db.AccountDeleted[username]; ok {
			continue
		}
		email := db.AccountEmail[username]
		joined := db.AccountCreated[username]
		amount := db.GiftAmount[id]
		created := db.GiftCreated[id]
		if err := callback(id, conversation, message, &authgo.Account{
			ID:       user,
			Username: username,
			Email:    email,
			Created:  joined,
		}, amount, created); err != nil {
			return err
		}
	}
	return nil
}

func (db *InMemory) SelectGiftsForUser(user int64) (int64, error) {
	db.Lock()
	defer db.Unlock()
	var gifts int64
	for mid := range db.MessageId {
		if db.MessageUser[mid] != user {
			continue
		}
		if _, ok := db.MessageDeleted[mid]; ok {
			continue
		}
		for gid := range db.GiftId {
			if db.GiftMessage[gid] != mid {
				continue
			}
			if _, ok := db.GiftDeleted[gid]; ok {
				continue
			}
			gifts += db.GiftAmount[gid]
		}
	}
	return gifts, nil
}

func (db *InMemory) SelectGiftsFromUser(user int64) (int64, error) {
	db.Lock()
	defer db.Unlock()
	var gifts int64
	for gid := range db.GiftId {
		if db.GiftUser[gid] != user {
			continue
		}
		if _, ok := db.GiftDeleted[gid]; ok {
			continue
		}
		gifts += db.GiftAmount[gid]
	}
	return gifts, nil
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
		if _, ok := db.ChargeDeleted[cid]; ok {
			continue
		}
		cost += db.ChargeAmount[cid]
	}
	return
}

func (db *InMemory) yield(id int64) (yield int64) {
	for yid := range db.YieldId {
		if db.YieldParent[yid] != id {
			continue
		}
		if _, ok := db.YieldDeleted[yid]; ok {
			continue
		}
		yield += db.YieldAmount[yid]
	}
	return
}
