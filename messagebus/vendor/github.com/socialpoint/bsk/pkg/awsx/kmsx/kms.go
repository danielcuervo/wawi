package kmsx

import (
	"sync"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/kms"
)

// AliasKeyMap represents a mapping between alias and keys.
// It can be fetched from a given AWS session
type AliasKeyMap struct {
	keys  map[string]string
	mutex sync.RWMutex
}

// NewAliasKeyMap creates a empty AliasKeyMap
func NewAliasKeyMap() *AliasKeyMap {
	return &AliasKeyMap{keys: make(map[string]string)}
}

// Fetch fetches all aliases with paging and creates a map alias->key
// Be aware that keys disabled or scheduled for deletion are also returned
func (am *AliasKeyMap) Fetch(p client.ConfigProvider) error {
	svc := kms.New(p)

	am.mutex.Lock()
	am.keys = make(map[string]string)
	am.mutex.Unlock()

	const (
		maxPages = 50 //so that we realize in case we don't destroy the keys created by integration tests
		pageSize = 100
	)
	pageNum := 0
	err := svc.ListAliasesPages(&kms.ListAliasesInput{Limit: aws.Int64(pageSize)},
		func(page *kms.ListAliasesOutput, lastPage bool) bool {
			pageNum++
			if pageNum > maxPages {
				return false
			}
			am.mutex.Lock()
			defer am.mutex.Unlock()

			for _, a := range page.Aliases {
				am.keys[aws.StringValue(a.AliasName)] = aws.StringValue(a.TargetKeyId)
			}
			return true
		})
	if pageNum >= maxPages {
		return fmt.Errorf("too many keys found (>%d)", maxPages*pageSize)
	}
	return err
}

// Get returns the kid with the given alias
func (am *AliasKeyMap) Get(alias string) string {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	return am.keys[alias]
}

// Exists return whether a key with the give alias exists or not
func (am *AliasKeyMap) Exists(alias string) bool {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	_, ok := am.keys[alias]

	return ok
}

// CreateKeyWithAlias creates a new keys and associates it with the given alias.
// If a key with the given alias already exists, it returns the existing key id.
func CreateKeyWithAlias(p client.ConfigProvider, alias string, desc string) (string, error) {
	keys := NewAliasKeyMap()
	err := keys.Fetch(p)
	if err != nil {
		return "", err
	}

	if keys.Exists(alias) {
		return keys.Get(alias), nil
	}

	svc := kms.New(p)
	req := &kms.CreateKeyInput{
		Description: aws.String(desc),
		KeyUsage:    aws.String("ENCRYPT_DECRYPT"),
	}

	res, err := svc.CreateKey(req)
	if err != nil {
		return "", err
	}

	kid := aws.StringValue(res.KeyMetadata.KeyId)

	_, err = svc.CreateAlias(&kms.CreateAliasInput{
		TargetKeyId: res.KeyMetadata.KeyId,
		AliasName:   aws.String(alias),
	})

	return kid, err
}

// ScheduleKeyDeletion schedules the key to be deleted in 7 days
// Will fail is key not found. It does not update keys to be consistent with Fetch
func ScheduleKeyDeletion(p client.ConfigProvider, keyID string) error {
	svc := kms.New(p)
	req := &kms.ScheduleKeyDeletionInput{
		KeyId:               aws.String(keyID),
		PendingWindowInDays: aws.Int64(7)}
	_, err := svc.ScheduleKeyDeletion(req)
	return err
}
