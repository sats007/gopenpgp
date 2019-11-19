package crypto

import (
	"encoding/base64"
	"testing"

	"github.com/ProtonMail/gopenpgp/constants"
	"github.com/stretchr/testify/assert"
)

var decodedSymmetricKey, _ = base64.StdEncoding.DecodeString("ExXmnSiQ2QCey20YLH6qlLhkY3xnIBC1AwlIXwK/HvY=")

var testSymmetricKey = &SymmetricKey{
	Key:  decodedSymmetricKey,
	Algo: constants.AES256,
}

var testWrongSymmetricKey = &SymmetricKey{
	Key:  []byte("WrongPass"),
	Algo: constants.AES256,
}

// Corresponding key in testdata/keyring_privateKey
var testMailboxPassword = []byte("apple")

// Corresponding key in testdata/keyring_privateKeyLegacy
// const testMailboxPasswordLegacy = [][]byte{ []byte("123") }

var (
	keyRingTestPrivate   *KeyRing
	keyRingTestPublic    *KeyRing
	keyRingTestMultiple  *KeyRing
)

var testIdentity = &Identity{
	Name:  "UserID",
	Email: "",
}

func initKeyRings() {
	var err error

	privateKey, err := NewKeyFromArmored(readTestFile("keyring_privateKey", false))
	if err != nil {
		panic("Expected no error while unarmoring private key, got:" + err.Error())
	}

	keyRingTestPrivate, err = NewKeyRing(privateKey)
	if err == nil {
		panic("Able to create a keyring with a locked key")
	}

	unlockedKey, err := privateKey.Unlock(testMailboxPassword)
	if err != nil {
		panic("Expected no error while unlocking private key, got:" + err.Error())
	}

	keyRingTestPrivate, err = NewKeyRing(unlockedKey)
	if err != nil {
		panic("Expected no error while building private keyring, got:" + err.Error())
	}

	publicKey, err := NewKeyFromArmored(readTestFile("keyring_publicKey", false))
	if err != nil {
		panic("Expected no error while unarmoring public key, got:" + err.Error())
	}

	keyRingTestPublic, err = NewKeyRing(publicKey)
	if err != nil {
		panic("Expected no error while building public keyring, got:" + err.Error())
	}

	keyRingTestMultiple, err = NewKeyRing(nil)
	if err != nil {
		panic("Expected no error while building empty keyring, got:" + err.Error())
	}

	err = keyRingTestMultiple.AddKey(keyTestRSA)
	if err != nil {
		panic("Expected no error while adding RSA key to keyring, got:" + err.Error())
	}

	err = keyRingTestMultiple.AddKey(keyTestEC)
	if err != nil {
		panic("Expected no error while adding EC key to keyring, got:" + err.Error())
	}

	err = keyRingTestMultiple.AddKey(unlockedKey)
	if err != nil {
		panic("Expected no error while adding unlocked key to keyring, got:" + err.Error())
	}
}


func TestIdentities(t *testing.T) {
	identities := keyRingTestPrivate.GetIdentities()
	assert.Len(t, identities, 1)
	assert.Exactly(t, identities[0], testIdentity)
}

func TestFilterExpiredKeys(t *testing.T) {
	expiredKey, err := NewKeyFromArmored(readTestFile("key_expiredKey", false))
	if err != nil {
		t.Fatal("Cannot unarmor expired key:", err)
	}

	expiredKeyRing, err := NewKeyRing(expiredKey)
	if err != nil {
		t.Fatal("Cannot create keyring with expired key:", err)
	}

	keys := []*KeyRing{keyRingTestPrivate, expiredKeyRing}
	unexpired, err := FilterExpiredKeys(keys)

	if err != nil {
		t.Fatal("Expected no error while filtering expired keyrings, got:", err)
	}

	assert.Len(t, unexpired, 1)
	assert.Exactly(t, unexpired[0].KeyIds(), keyRingTestPrivate.KeyIds())
}

func TestKeyIds(t *testing.T) {
	keyIDs := keyRingTestPrivate.KeyIds()
	var assertKeyIDs = []uint64{4518840640391470884}
	assert.Exactly(t, assertKeyIDs, keyIDs)
}

func TestMultipleKeyRing(t *testing.T) {
	assert.Exactly(t, 3, len(keyRingTestMultiple.entities))
	assert.Exactly(t, 3, keyRingTestMultiple.CountEntities())
	assert.Exactly(t, 3, keyRingTestMultiple.CountDecryptionEntities())

	singleKeyRing, err := keyRingTestMultiple.FirstKey()
	if err != nil {
		t.Fatal("Expected no error while filtering the first key, got:", err)
	}
	assert.Exactly(t, 1, len(singleKeyRing.entities))
	assert.Exactly(t, 1, singleKeyRing.CountEntities())
	assert.Exactly(t, 1, singleKeyRing.CountDecryptionEntities())
}
