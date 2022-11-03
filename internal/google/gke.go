package google

import (
	container "cloud.google.com/go/container/apiv1"
	"context"
	"errors"
	"fmt"
	"github.com/kyokomi/emoji/v2"
	containerpb "google.golang.org/genproto/googleapis/container/v1"
	"log"
	"os/exec"
	"time"
)

// counter is used to check status
var counter int

// Cluster represents a GKE cluster
type Cluster struct {
	Name                 string
	ProjectID            string
	Region               string
	CryptoKeyName        string
	Network              string
	Subnetwork           string
	MachineType          string
	DiskSizeGb           int32
	MinNodeCount         int32
	MaxNodeCount         int32
	MasterAuthCidrBlocks []*containerpb.MasterAuthorizedNetworksConfig_CidrBlock
	MasterIpv4CidrBlock  string
}

// Create GKE cluster
func (c *Cluster) create(ctx context.Context, client *container.ClusterManagerClient) (*containerpb.Cluster, error) {
	if c.exists(ctx, client) {
		if err := c.connect(); err != nil {
			//just log error and continue
			log.Println(err)
		} else {
			emoji.Println(":check_mark_button: Controlplane cluster added to kubeconfig")
		}
		return c.get(ctx, client)
	}

	req := &containerpb.CreateClusterRequest{
		Cluster: &containerpb.Cluster{
			Name:    c.Name,
			Network: c.Network,
			AddonsConfig: &containerpb.AddonsConfig{
				HttpLoadBalancing: &containerpb.HttpLoadBalancing{
					Disabled: false,
				},
				HorizontalPodAutoscaling: &containerpb.HorizontalPodAutoscaling{
					Disabled: false,
				},
				ConfigConnectorConfig: &containerpb.ConfigConnectorConfig{
					Enabled: true,
				},
				GcePersistentDiskCsiDriverConfig: &containerpb.GcePersistentDiskCsiDriverConfig{
					Enabled: true,
				},
				GcpFilestoreCsiDriverConfig: &containerpb.GcpFilestoreCsiDriverConfig{
					Enabled: true,
				},
			},
			DatabaseEncryption: &containerpb.DatabaseEncryption{
				State:   containerpb.DatabaseEncryption_ENCRYPTED,
				KeyName: c.CryptoKeyName,
			},
			Subnetwork: c.Subnetwork,
			NodePools: []*containerpb.NodePool{
				{
					Name: "default-pool",
					Config: &containerpb.NodeConfig{
						MachineType: c.MachineType,
						DiskSizeGb:  c.DiskSizeGb,
						OauthScopes: []string{
							"https://www.googleapis.com/auth/devstorage.read_only",
							"https://www.googleapis.com/auth/logging.write",
							"https://www.googleapis.com/auth/monitoring",
							"https://www.googleapis.com/auth/servicecontrol",
							"https://www.googleapis.com/auth/service.management.readonly",
							"https://www.googleapis.com/auth/trace.append",
							"https://www.googleapis.com/auth/cloud-platform",
						},
						Tags: []string{
							"default-pool",
						},
						DiskType: "pd-ssd",
						WorkloadMetadataConfig: &containerpb.WorkloadMetadataConfig{
							Mode: 2,
						},
						ShieldedInstanceConfig: &containerpb.ShieldedInstanceConfig{
							EnableSecureBoot: true,
						},
					},
					InitialNodeCount: 1,
					Autoscaling: &containerpb.NodePoolAutoscaling{
						Enabled:      true,
						MinNodeCount: c.MinNodeCount,
						MaxNodeCount: c.MaxNodeCount,
					},
					Management: &containerpb.NodeManagement{
						AutoUpgrade: true,
						AutoRepair:  true,
					},
					UpgradeSettings: &containerpb.NodePool_UpgradeSettings{
						MaxSurge:       1,
						MaxUnavailable: 1,
					},
				},
			},
			IpAllocationPolicy: &containerpb.IPAllocationPolicy{
				UseIpAliases:               true,
				ClusterSecondaryRangeName:  "pods",
				ServicesSecondaryRangeName: "services",
			},
			MasterAuthorizedNetworksConfig: &containerpb.MasterAuthorizedNetworksConfig{
				Enabled:    true,
				CidrBlocks: c.MasterAuthCidrBlocks,
			},
			MaintenancePolicy: &containerpb.MaintenancePolicy{
				Window: &containerpb.MaintenanceWindow{
					Policy: &containerpb.MaintenanceWindow_DailyMaintenanceWindow{
						DailyMaintenanceWindow: &containerpb.DailyMaintenanceWindow{
							StartTime: "06:00",
						},
					},
				},
			},
			BinaryAuthorization: &containerpb.BinaryAuthorization{
				Enabled: true,
			},
			NetworkConfig: &containerpb.NetworkConfig{
				EnableIntraNodeVisibility: true,
				DatapathProvider:          0,
			},
			PrivateClusterConfig: &containerpb.PrivateClusterConfig{
				EnablePrivateNodes:  true,
				MasterIpv4CidrBlock: c.MasterIpv4CidrBlock,
			},
			ShieldedNodes: &containerpb.ShieldedNodes{
				Enabled: true,
			},
			ReleaseChannel: &containerpb.ReleaseChannel{
				Channel: 1,
			},
			WorkloadIdentityConfig: &containerpb.WorkloadIdentityConfig{
				WorkloadPool: fmt.Sprintf("%s.svc.id.goog", c.ProjectID),
			},
		},
		Parent: fmt.Sprintf("projects/%s/locations/%s", c.ProjectID, c.Region),
	}

	op, err := client.CreateCluster(ctx, req)

	if err != nil {
		return nil, err
	}

status:
	for {
		if counter%10 == 0 {
			emoji.Println(":beer: Cluster is being created")
			time.Sleep(time.Second * 30)
		}
		s, err := client.GetOperation(ctx, &containerpb.GetOperationRequest{
			Name: fmt.Sprintf("projects/%s/locations/%s/operations/%s", c.ProjectID, c.Region, op.GetName()),
		})
		if err != nil {
			return nil, err
		}
		switch s.GetStatus().Number() {
		case 3:
			break status
		case 4:
			return nil, errors.New(s.GetError().Message)
		}
		counter++
	}
	if err := c.connect(); err != nil {
		//just log error and continue
		log.Println(err)
	} else {
		emoji.Println(":check_mark_button: Controlplane cluster added to kubeconfig")
	}
	return c.get(ctx, client)
}

func (c *Cluster) connect() error {
	p, err := exec.LookPath("gcloud")
	if err != nil {
		return errors.New("no gcloud in path... need to interface with API in future")
	}
	cmd := exec.Command(p, "container", "clusters", "get-credentials", c.Name, "--region", c.Region, "--project", c.ProjectID)
	if err := cmd.Run(); err != nil {
		return errors.New("could not connect")
	}
	return nil
}

// Get GKE cluster
func (c *Cluster) get(ctx context.Context, client *container.ClusterManagerClient) (*containerpb.Cluster, error) {
	req := &containerpb.GetClusterRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/clusters/%s", c.ProjectID, c.Region, c.Name),
	}

	resp, err := client.GetCluster(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Check if GKE cluster exists
func (c *Cluster) exists(ctx context.Context, client *container.ClusterManagerClient) bool {
	_, err := c.get(ctx, client)
	return err == nil
}

// Delete GKE cluster
func (c *Cluster) delete(ctx context.Context, client *container.ClusterManagerClient) error {
	if c.exists(ctx, client) {
		req := &containerpb.DeleteClusterRequest{
			Name: fmt.Sprintf("projects/%s/locations/%s/clusters/%s", c.ProjectID, c.Region, c.Name),
		}
		op, err := client.DeleteCluster(ctx, req)
		if err != nil {
			return err
		}
	status:
		for {
			if counter%10 == 0 {
				emoji.Println(":beer: Cluster is being deleted")
				time.Sleep(time.Second * 30)
			}
			s, err := client.GetOperation(ctx, &containerpb.GetOperationRequest{
				Name: fmt.Sprintf("projects/%s/locations/%s/operations/%s", c.ProjectID, c.Region, op.GetName()),
			})
			if err != nil {
				return err
			}
			switch s.GetStatus().Number() {
			case 3:
				break status
			case 4:
				return errors.New(s.GetError().Message)
			}
			counter++
		}
	}
	return nil
}

// Update GKE cluster
func (c *Cluster) update(ctx context.Context, client *container.ClusterManagerClient) (*containerpb.Cluster, error) {
	if c.exists(ctx, client) {
		return c.get(ctx, client)
	}

	creq := &containerpb.UpdateClusterRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/clusters/%s", c.ProjectID, c.Region, c.Name),
		Update: &containerpb.ClusterUpdate{
			DesiredAddonsConfig: &containerpb.AddonsConfig{
				HttpLoadBalancing: &containerpb.HttpLoadBalancing{
					Disabled: false,
				},
				HorizontalPodAutoscaling: &containerpb.HorizontalPodAutoscaling{
					Disabled: false,
				},
				ConfigConnectorConfig: &containerpb.ConfigConnectorConfig{
					Enabled: true,
				},
				GcePersistentDiskCsiDriverConfig: &containerpb.GcePersistentDiskCsiDriverConfig{
					Enabled: true,
				},
				GcpFilestoreCsiDriverConfig: &containerpb.GcpFilestoreCsiDriverConfig{
					Enabled: true,
				},
			},
			DesiredNodePoolId: "default-pool",
			DesiredNodePoolAutoscaling: &containerpb.NodePoolAutoscaling{
				Enabled:      true,
				MinNodeCount: c.MinNodeCount,
				MaxNodeCount: c.MaxNodeCount,
			},
			DesiredMasterAuthorizedNetworksConfig: &containerpb.MasterAuthorizedNetworksConfig{
				Enabled:    true,
				CidrBlocks: c.MasterAuthCidrBlocks,
			},
			DesiredBinaryAuthorization: &containerpb.BinaryAuthorization{
				Enabled: true,
			},
			DesiredIntraNodeVisibilityConfig: &containerpb.IntraNodeVisibilityConfig{
				Enabled: true,
			},
			DesiredReleaseChannel: &containerpb.ReleaseChannel{
				Channel: 1,
			},
			DesiredPrivateClusterConfig: &containerpb.PrivateClusterConfig{
				EnablePrivateNodes:  true,
				MasterIpv4CidrBlock: c.MasterIpv4CidrBlock,
			},
			DesiredDatapathProvider: 0,
			DesiredShieldedNodes: &containerpb.ShieldedNodes{
				Enabled: true,
			},
		},
	}

	op, err := client.UpdateCluster(ctx, creq)

	if err != nil {
		return nil, err
	}

cstatus:
	for {
		s, err := client.GetOperation(ctx, &containerpb.GetOperationRequest{
			Name: fmt.Sprintf("projects/%s/locations/%s/operations/%s", c.ProjectID, c.Region, op.GetName()),
		})
		if err != nil {
			return nil, err
		}
		switch s.GetStatus().Number() {
		case 3:
			break cstatus
		case 4:
			return nil, errors.New(s.GetError().Message)
		}
	}

	cluster, _ := c.get(ctx, client)

	nreq := &containerpb.UpdateNodePoolRequest{
		Name:        fmt.Sprintf("projects/%s/locations/%s/clusters/%s/nodePools/default-pool", c.ProjectID, c.Region, c.Name),
		NodeVersion: "-",
		ImageType:   cluster.NodePools[0].Config.ImageType,
		WorkloadMetadataConfig: &containerpb.WorkloadMetadataConfig{
			Mode: 2,
		},
		UpgradeSettings: &containerpb.NodePool_UpgradeSettings{
			MaxSurge:       1,
			MaxUnavailable: 1,
		},
		Tags: &containerpb.NetworkTags{
			Tags: []string{
				"default-pool",
			},
		},
	}

	op, err = client.UpdateNodePool(ctx, nreq)

	if err != nil {
		return nil, err
	}

nstatus:
	for {
		s, err := client.GetOperation(ctx, &containerpb.GetOperationRequest{
			Name: fmt.Sprintf("projects/%s/locations/%s/operations/%s", c.ProjectID, c.Region, op.GetName()),
		})
		if err != nil {
			return nil, err
		}
		switch s.GetStatus().Number() {
		case 3:
			break nstatus
		case 4:
			return nil, errors.New(s.GetError().Message)
		}
	}

	return c.get(ctx, client)
}
