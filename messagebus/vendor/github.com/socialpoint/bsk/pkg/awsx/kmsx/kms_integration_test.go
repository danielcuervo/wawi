// +build integration

package kmsx_test

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/socialpoint/bsk/pkg/awsx/awstest"
	"github.com/socialpoint/bsk/pkg/awsx/kmsx"
	"github.com/socialpoint/bsk/pkg/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAliasKeyMap_Fetch(t *testing.T) {
	assert := assert.New(t)

	session := awstest.NewSession()
	am := kmsx.NewAliasKeyMap()
	alias := "alias/integration-test-" + uuid.New()

	kid, err := kmsx.CreateKeyWithAlias(session, alias, "testing alias")
	assert.NoError(err)
	assert.NotEmpty(kid)

	err = am.Fetch(session)
	assert.NoError(err)

	assertAliasEventuallyExists(t, session, am, alias, kid)

	assert.NoError(kmsx.ScheduleKeyDeletion(session, kid))
}

func TestAliasKeyMap_CreateKeyWithAlias(t *testing.T) {
	assert := assert.New(t)

	session := awstest.NewSession()
	am := kmsx.NewAliasKeyMap()
	alias := "alias/integration-test-" + uuid.New()

	kid, err := kmsx.CreateKeyWithAlias(session, alias, "testing alias")
	assert.NoError(err)
	assert.NotEmpty(kid)

	assertAliasEventuallyExists(t, session, am, alias, kid)

	assert.NoError(kmsx.ScheduleKeyDeletion(session, kid))
}

func assertAliasEventuallyExists(t *testing.T, session *session.Session, am *kmsx.AliasKeyMap, alias, kid string) {
	assert := assert.New(t)

	// Keys are not immediately available, so let's try for a while
	for i := 0; i < 5; i++ {

		err := am.Fetch(session)
		assert.NoError(err)

		if am.Exists(alias) {
			assert.Equal(kid, am.Get(alias))
			return
		}

		time.Sleep(time.Second)
	}

	t.Fatal("Alias " + alias + " not found after several tries")
}
