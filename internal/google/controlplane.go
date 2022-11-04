package google

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	container "cloud.google.com/go/container/apiv1"
	admin "cloud.google.com/go/iam/admin/apiv1"
	kms "cloud.google.com/go/kms/apiv1"
	"github.com/kyokomi/emoji/v2"
	crmgr "google.golang.org/api/cloudresourcemanager/v1"
)

// Controlplane contains values for a GKE clutser and its dependencies
type Controlplane struct {
	Apis []string
	Vpc
	Subnetwork
	Router
	Cluster
	Firewalls []Firewall
	Keyring
	CryptoKey
	ServiceAccount
}

// BoolPtr convertes a bool to *bool
func BoolPtr(b bool) *bool {
	return &b
}

// StrPtr convertes a string to *string
func StrPtr(s string) *string {
	return &s
}

// Create controlplane
func (c *Controlplane) Create() error {
	ctx := context.Background()

	adminClient, err := admin.NewIamClient(ctx)
	if err != nil {
		return err
	}
	defer adminClient.Close()
	sa, err := c.ServiceAccount.create(ctx, adminClient)
	if err != nil {
		return err
	}

	crmgrClient, err := crmgr.NewService(ctx)
	if err != nil {
		return err
	}
	policy, err := GetProjectPolicy(crmgrClient, c.ServiceAccount.ProjectID)
	if err != nil {
		return err
	}
	policy = AddProjectBinding(policy, c.ServiceAccount.ProjectID, c.ServiceAccount.ProjectRole, fmt.Sprintf("serviceAccount:%s", sa.Email))
	if err := SetProjectIam(crmgrClient, c.ServiceAccount.ProjectID, policy); err != nil {
		return err
	}
	emoji.Println(":check_mark_button: Controlplane service account created")

	// vpcClient, err := compute.NewNetworksRESTClient(ctx)
	// if err != nil {
	// 	return err
	// }
	// defer vpcClient.Close()
	// network, err := c.Vpc.create(ctx, vpcClient)
	// if err != nil {
	// 	return err
	// }
	// emoji.Println(":check_mark_button: Controlplane VPC created")

	// c.Subnetwork.Network = network.GetSelfLink()
	// subnetClient, err := compute.NewSubnetworksRESTClient(ctx)
	// if err != nil {
	// 	return err
	// }
	// defer subnetClient.Close()
	// _, err = c.Subnetwork.create(ctx, subnetClient)
	// if err != nil {
	// 	return err
	// }
	// emoji.Println(":check_mark_button: Controlplane subnetwork created")

	// c.Router.Network = network.GetSelfLink()
	// routerClient, err := compute.NewRoutersRESTClient(ctx)
	// if err != nil {
	// 	return err
	// }
	// defer routerClient.Close()
	// _, err = c.Router.create(ctx, routerClient)
	// if err != nil {
	// 	return err
	// }
	// emoji.Println(":check_mark_button: Controlplane router created")

	// kmsClient, err := kms.NewKeyManagementClient(ctx)
	// if err != nil {
	// 	return err
	// }
	// defer kmsClient.Close()

	// keyring, err := c.Keyring.create(ctx, kmsClient)
	// if err != nil {
	// 	return err
	// }
	// emoji.Println(":check_mark_button: Controlplane KMS Keyring created")

	// c.CryptoKey.Keyring = keyring.GetName()
	// cryptoKey, err := c.CryptoKey.create(ctx, kmsClient)
	// if err != nil {
	// 	return err
	// }
	// emoji.Println(":check_mark_button: Controlplane KMS Crypto Key created")

	// c.Cluster.CryptoKeyName = cryptoKey.GetName()
	// clusterClient, err := container.NewClusterManagerClient(ctx)
	// if err != nil {
	// 	return err
	// }
	// defer clusterClient.Close()
	// _, err = c.Cluster.create(ctx, clusterClient)
	// if err != nil {
	// 	return err
	// }
	// emoji.Println(":check_mark_button: Controlplane cluster created")

	// firewallClient, err := compute.NewFirewallsRESTClient(ctx)
	// if err != nil {
	// 	return err
	// }
	// defer firewallClient.Close()
	// for i := range c.Firewalls {
	// 	c.Firewalls[i].Network = network.GetSelfLink()
	// 	_, err = c.Firewalls[i].create(ctx, firewallClient)
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	// emoji.Println(":check_mark_button: Controlplane firewall rules created")

	return nil
}

// Delete controlplane
func (c *Controlplane) Delete() error {
	ctx := context.Background()

	// firewallClient, err := compute.NewFirewallsRESTClient(ctx)
	// if err != nil {
	// 	return err
	// }
	// defer firewallClient.Close()
	// for i := range c.Firewalls {
	// 	err = c.Firewalls[i].delete(ctx, firewallClient)
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	// emoji.Println(":cross_mark_button: Controlplane firewall rules deleted")

	// clusterClient, err := container.NewClusterManagerClient(ctx)
	// if err != nil {
	// 	return err
	// }
	// defer clusterClient.Close()
	// err = c.Cluster.delete(ctx, clusterClient)
	// if err != nil {
	// 	return err
	// }
	// emoji.Println(":cross_mark_button: Controlplane cluster destroyed")

	// kmsClient, err := kms.NewKeyManagementClient(ctx)
	// if err != nil {
	// 	return err
	// }
	// defer kmsClient.Close()
	// keyring, err := c.Keyring.get(ctx, kmsClient)
	// if err != nil {
	// 	return err
	// }
	// c.CryptoKey.Keyring = keyring.GetName()
	// err = c.CryptoKey.delete(ctx, kmsClient)
	// if err != nil {
	// 	return err
	// }
	// emoji.Println(":cross_mark_button: Controlplane KMS Crypto Key IAM permissions deleted")

	// routerClient, err := compute.NewRoutersRESTClient(ctx)
	// if err != nil {
	// 	return err
	// }
	// defer routerClient.Close()
	// err = c.Router.delete(ctx, routerClient)
	// if err != nil {
	// 	return err
	// }
	// emoji.Println(":cross_mark_button: Controlplane router destroyed")

	// subnetClient, err := compute.NewSubnetworksRESTClient(ctx)
	// if err != nil {
	// 	return err
	// }
	// defer subnetClient.Close()
	// err = c.Subnetwork.delete(ctx, subnetClient)
	// if err != nil {
	// 	return err
	// }
	// emoji.Println(":cross_mark_button: Controlplane subnetwork destroyed")

	// vpcClient, err := compute.NewNetworksRESTClient(ctx)
	// if err != nil {
	// 	return err
	// }
	// defer vpcClient.Close()
	// err = c.Vpc.delete(ctx, vpcClient)
	// if err != nil {
	// 	return err
	// }
	// emoji.Println(":cross_mark_button: Controlplane VPC destroyed")

	adminClient, err := admin.NewIamClient(ctx)
	if err != nil {
		return err
	}
	defer adminClient.Close()

	sa, err := c.ServiceAccount.get(ctx, adminClient)
	if err != nil {
		return err
	}

	crmgrClient, err := crmgr.NewService(ctx)
	if err != nil {
		return err
	}
	policy, err := GetProjectPolicy(crmgrClient, c.ServiceAccount.ProjectID)
	if err != nil {
		return err
	}
	policy = RemoveProjectBinding(policy, c.ServiceAccount.ProjectID, c.ServiceAccount.ProjectRole, fmt.Sprintf("serviceAccount:%s", sa.Email))
	if err := SetProjectIam(crmgrClient, c.ServiceAccount.ProjectID, policy); err != nil {
		return err
	}

	_, err = c.ServiceAccount.delete(ctx, adminClient)
	if err != nil {
		return err
	}
	emoji.Println(":cross_mark_button: Controlplane service account destroyed")

	return nil
}

// Update controlplane
func (c *Controlplane) Update() error {
	ctx := context.Background()

	vpcClient, err := compute.NewNetworksRESTClient(ctx)
	if err != nil {
		return err
	}
	defer vpcClient.Close()
	network, err := c.Vpc.update(ctx, vpcClient)
	if err != nil {
		return err
	}
	emoji.Println(":check_mark_button: Controlplane VPC updated")

	c.Subnetwork.Network = network.GetSelfLink()
	subnetClient, err := compute.NewSubnetworksRESTClient(ctx)
	if err != nil {
		return err
	}
	defer subnetClient.Close()
	_, err = c.Subnetwork.update(ctx, subnetClient)
	if err != nil {
		return err
	}
	emoji.Println(":check_mark_button: Controlplane subnetwork updated")

	c.Router.Network = network.GetSelfLink()
	routerClient, err := compute.NewRoutersRESTClient(ctx)
	if err != nil {
		return err
	}
	defer routerClient.Close()
	_, err = c.Router.update(ctx, routerClient)
	if err != nil {
		return err
	}
	emoji.Println(":check_mark_button: Controlplane router updated")

	kmsClient, err := kms.NewKeyManagementClient(ctx)
	if err != nil {
		return err
	}
	defer kmsClient.Close()
	keyring, err := c.Keyring.get(ctx, kmsClient)
	if err != nil {
		return err
	}
	c.CryptoKey.Keyring = keyring.GetName()
	cryptoKey, err := c.CryptoKey.update(ctx, kmsClient)
	if err != nil {
		return err
	}
	emoji.Println(":check_mark_button: Controlplane KMS Crypto Key updated")

	c.Cluster.CryptoKeyName = cryptoKey.GetName()
	clusterClient, err := container.NewClusterManagerClient(ctx)
	if err != nil {
		return err
	}
	defer clusterClient.Close()
	_, err = c.Cluster.update(ctx, clusterClient)
	if err != nil {
		return err
	}
	emoji.Println(":check_mark_button: Controlplane cluster updated")

	firewallClient, err := compute.NewFirewallsRESTClient(ctx)
	if err != nil {
		return err
	}
	defer firewallClient.Close()
	for i := range c.Firewalls {
		c.Firewalls[i].Network = network.GetSelfLink()
		_, err = c.Firewalls[i].update(ctx, firewallClient)
		if err != nil {
			return err
		}
	}
	emoji.Println(":check_mark_button: Controlplane firewall rules updated")

	return nil
}
