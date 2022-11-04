package google

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/pkg/errors"

	kms "cloud.google.com/go/kms/apiv1"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// CryptoKey represents a KMS Crypto Key
type CryptoKey struct {
	Name          string
	Keyring       string
	ProjectID     string
	ProjectNumber string
}

func setKmsIam(ctx context.Context, client *kms.KeyManagementClient, id string, key *kmspb.CryptoKey) error {
	member := fmt.Sprintf("serviceAccount:service-%s@container-engine-robot.iam.gserviceaccount.com", id)
	handle := client.ResourceIAM(key.GetName())
	policy, err := handle.Policy(ctx)
	if err != nil {
		return err
	}
	policy.Add(member, "roles/cloudkms.cryptoKeyDecrypter")
	policy.Add(member, "roles/cloudkms.cryptoKeyEncrypter")
	return handle.SetPolicy(ctx, policy)
}

func removeKmsIam(ctx context.Context, client *kms.KeyManagementClient, id string, key *kmspb.CryptoKey) error {
	member := fmt.Sprintf("serviceAccount:service-%s@container-engine-robot.iam.gserviceaccount.com", id)
	handle := client.ResourceIAM(key.GetName())
	policy, err := handle.Policy(ctx)
	if err != nil {
		return err
	}
	policy.Remove(member, "roles/cloudkms.cryptoKeyDecrypter")
	policy.Remove(member, "roles/cloudkms.cryptoKeyEncrypter")
	return handle.SetPolicy(ctx, policy)
}

// Create KMS Crypto Key
func (c *CryptoKey) create(ctx context.Context, client *kms.KeyManagementClient) (*kmspb.CryptoKey, error) {
	k, ok := c.exists(ctx, client)
	if ok {
		if err := setKmsIam(ctx, client, c.ProjectNumber, k); err != nil {
			return nil, err
		}
		_, err := c.checkVersion(ctx, client, k)
		if err != nil {
			return nil, err
		}
		return c.get(ctx, client)
	}
	cryptoKey := &kmspb.CryptoKey{
		Purpose: kmspb.CryptoKey_ENCRYPT_DECRYPT,
	}
	req := &kmspb.CreateCryptoKeyRequest{
		Parent:                     c.Keyring,
		CryptoKeyId:                c.Name,
		CryptoKey:                  cryptoKey,
		SkipInitialVersionCreation: false,
	}
	resp, err := client.CreateCryptoKey(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := setKmsIam(ctx, client, c.ProjectNumber, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Get KMS Crypto Key
func (c *CryptoKey) get(ctx context.Context, client *kms.KeyManagementClient) (*kmspb.CryptoKey, error) {
	req := &kmspb.GetCryptoKeyRequest{
		Name: fmt.Sprintf("%s/cryptoKeys/%s", c.Keyring, c.Name),
	}
	resp, err := client.GetCryptoKey(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Check if KMS Crypto Key exists
func (c *CryptoKey) exists(ctx context.Context, client *kms.KeyManagementClient) (*kmspb.CryptoKey, bool) {
	k, err := c.get(ctx, client)
	if err != nil {
		return nil, false
	}
	return k, err == nil
}

func (c *CryptoKey) checkVersion(ctx context.Context, client *kms.KeyManagementClient, k *kmspb.CryptoKey) (*kmspb.CryptoKeyVersion, error) {
	kp := k.GetPrimary()
	switch kp.State {
	case kmspb.CryptoKeyVersion_DISABLED:
		log.Printf("cryptokey version %s is disabled\n", kp.GetName())
		kv, err := enableKeyVersion(ctx, client, kp)
		if err != nil {
			return kv, err
		}
		return kv, nil
	case kmspb.CryptoKeyVersion_DESTROYED:
		log.Printf("cryptokey %s is destroyed\n", kp.GetName())
		kv, err := createKeyVersion(ctx, client, k)
		if err != nil {
			return kv, err
		}
		return kv, nil
	case kmspb.CryptoKeyVersion_DESTROY_SCHEDULED:
		log.Printf("cryptokey %s is scheduled to be destroyed\n", k.GetName())
		kv, err := restoreKeyVersion(ctx, client, kp)
		if err != nil {
			return kv, err
		}
		return kv, nil
	}
	return kp, nil
}

// Update KMS Crypto Key
func (c *CryptoKey) update(ctx context.Context, client *kms.KeyManagementClient) (*kmspb.CryptoKey, error) {
	key, err := c.get(ctx, client)
	if err != nil {
		return nil, err
	}
	if err := setKmsIam(ctx, client, c.ProjectNumber, key); err != nil {
		return nil, err
	}
	return key, nil
}

func (c *CryptoKey) delete(ctx context.Context, client *kms.KeyManagementClient) error {
	key, err := c.get(ctx, client)
	if err != nil {
		return err
	}
	return removeKmsIam(ctx, client, c.ProjectNumber, key)
}

func waitKeyVersion(ctx context.Context, client *kms.KeyManagementClient, kv *kmspb.CryptoKeyVersion, state kmspb.CryptoKeyVersion_CryptoKeyVersionState) (*kmspb.CryptoKeyVersion, error) {
	for {
		kv, err := client.GetCryptoKeyVersion(ctx, &kmspb.GetCryptoKeyVersionRequest{
			Name: kv.GetName(),
		})
		if err != nil {
			return nil, err
		}
		if kv.GetState() == state {
			break
		}
	}
	return kv, nil
}

func waitPrimaryKey(ctx context.Context, client *kms.KeyManagementClient, k *kmspb.CryptoKey, kv *kmspb.CryptoKeyVersion) (*kmspb.CryptoKeyVersion, error) {
	for {
		k, err := client.GetCryptoKey(ctx, &kmspb.GetCryptoKeyRequest{
			Name: k.GetName(),
		})
		if err != nil {
			return nil, err
		}
		if kv.GetName() == k.GetPrimary().GetName() {
			break
		}
	}
	return kv, nil
}

func enableKeyVersion(ctx context.Context, client *kms.KeyManagementClient, kv *kmspb.CryptoKeyVersion) (*kmspb.CryptoKeyVersion, error) {
	req := &kmspb.UpdateCryptoKeyVersionRequest{
		CryptoKeyVersion: &kmspb.CryptoKeyVersion{
			Name:  kv.GetName(),
			State: kmspb.CryptoKeyVersion_ENABLED,
		},
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{"state"},
		},
	}
	kv, err := client.UpdateCryptoKeyVersion(ctx, req)
	if err != nil {
		return nil, err
	}
	return waitKeyVersion(ctx, client, kv, kmspb.CryptoKeyVersion_ENABLED)
}

func createKeyVersion(ctx context.Context, client *kms.KeyManagementClient, k *kmspb.CryptoKey) (*kmspb.CryptoKeyVersion, error) {
	req := &kmspb.CreateCryptoKeyVersionRequest{
		Parent: k.GetName(),
	}
	kv, err := client.CreateCryptoKeyVersion(ctx, req)
	if err != nil {
		return nil, err
	}
	kv, err = waitKeyVersion(ctx, client, kv, kmspb.CryptoKeyVersion_ENABLED)
	if err != nil {
		return nil, err
	}
	kid := strings.Split(kv.GetName(), "/")
	updateReq := &kmspb.UpdateCryptoKeyPrimaryVersionRequest{
		Name:               k.GetName(),
		CryptoKeyVersionId: kid[len(kid)-1],
	}
	k, err = client.UpdateCryptoKeyPrimaryVersion(ctx, updateReq)
	if err != nil {
		return nil, errors.Wrap(err, "unable to set key version to primary")
	}
	return waitPrimaryKey(ctx, client, k, kv)
}

func restoreKeyVersion(ctx context.Context, client *kms.KeyManagementClient, k *kmspb.CryptoKeyVersion) (*kmspb.CryptoKeyVersion, error) {
	req := &kmspb.RestoreCryptoKeyVersionRequest{
		Name: k.GetName(),
	}
	kv, err := client.RestoreCryptoKeyVersion(ctx, req)
	if err != nil {
		return nil, err
	}
	return waitKeyVersion(ctx, client, kv, kmspb.CryptoKeyVersion_DISABLED)
}
