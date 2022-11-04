/*
Package cmd is the entrypoint the for cli
*/
package cmd

import (
	"fmt"
	"log"
	"tidalwave/internal/google"

	"github.com/kyokomi/emoji/v2"
	"github.com/spf13/viper"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
	containerpb "google.golang.org/genproto/googleapis/container/v1"
)

func googleDefaults() {
	viper.SetDefault("spec.region", "us-central1")
	viper.SetDefault("spec.cidrs.nodes", "10.0.0.0/24")
	viper.SetDefault("spec.cidrs.pods", "10.1.0.0/16")
	viper.SetDefault("spec.cidrs.services", "10.2.0.0/20")
	viper.SetDefault("spec.cluster.machineType", "n2-standard-4")
	viper.SetDefault("spec.cluster.minNodeCount", 1)
	viper.SetDefault("spec.cluster.maxNodeCount", 3)
	viper.SetDefault("spec.cluster.masterAuthBlock", []map[string]string{
		{
			"displayName": "public",
			"cidrBlock":   "0.0.0.0/0",
		},
	})
	viper.SetDefault("spec.cluster.masterCidrBlock", "172.16.0.0/28")
}

// CreateGoogleControlplane creates google.Controlplane from options form the config file
func CreateGoogleControlplane() (*google.Controlplane, error) {
	googleDefaults()
	name := viper.GetString("metadata.name")
	if name == "" {
		log.Fatalln("metadata.name cannot be nil")
	}
	projectID := viper.GetString("spec.projectID")
	if projectID == "" {
		log.Fatalln("spec.projectID cannot be nil")
	}
	projectNumber, err := google.GetProjectNumber(projectID)
	if err != nil {
		log.Fatalf("project-id %s not found: %s\n", projectID, err)
	}
	emoji.Printf(":bullseye: Project Id: %s\n", projectID)
	emoji.Printf(":bullseye: Project Number: %s\n", *projectNumber)
	region := viper.GetString("spec.region")
	nodesCidr := viper.GetString("spec.cidrs.nodes")
	podCidr := viper.GetString("spec.cidrs.pods")
	serviceCidr := viper.GetString("spec.cidrs.services")
	machineType := viper.GetString("spec.cluster.machineType")
	diskSize := viper.GetInt32("spec.cluster.diskSize")
	minNodes := viper.GetInt32("spec.cluster.minNodeCount")
	maxNodes := viper.GetInt32("spec.cluster.maxNodeCount")
	masterAuthCidrBlocks := []*containerpb.MasterAuthorizedNetworksConfig_CidrBlock{}
	if err := viper.UnmarshalKey("spec.cluster.masterAuthBlock", &masterAuthCidrBlocks); err != nil {
		return nil, err
	}
	masterIpv4CidrBlock := viper.GetString("spec.cluster.masterCidrBlock")
	cp := google.Controlplane{
		ServiceAccount: google.ServiceAccount{
			Name:        "config-connector",
			ProjectID:   projectID,
			DisplayName: "Config Connector",
			Description: "Config Connector controller",
		},
		Apis: google.RequiredApis.Services,
		Vpc: google.Vpc{
			Name:      name,
			ProjectID: projectID,
		},
		Subnetwork: google.Subnetwork{
			Name:         name,
			ProjectID:    projectID,
			Region:       region,
			NodesCidr:    nodesCidr,
			PodsCidr:     podCidr,
			ServicesCidr: serviceCidr,
		},
		Router: google.Router{
			Name:      name,
			ProjectID: projectID,
			Region:    region,
		},
		Keyring: google.Keyring{
			Name:      name,
			ProjectID: projectID,
			Region:    region,
		},
		CryptoKey: google.CryptoKey{
			Name:          name,
			ProjectID:     projectID,
			ProjectNumber: *projectNumber,
		},
		Cluster: google.Cluster{
			Name:                 name,
			ProjectID:            projectID,
			Region:               region,
			Network:              name,
			Subnetwork:           name,
			MachineType:          machineType,
			DiskSizeGb:           diskSize,
			MinNodeCount:         minNodes,
			MaxNodeCount:         maxNodes,
			MasterAuthCidrBlocks: masterAuthCidrBlocks,
			MasterIpv4CidrBlock:  masterIpv4CidrBlock,
		},
		Firewalls: []google.Firewall{
			{
				Name:      fmt.Sprintf("%s-intra-cluster-egress", name),
				ProjectID: projectID,
				Allowed: []*computepb.Allowed{
					{
						IPProtocol: google.StrPtr("tcp"),
					},
					{
						IPProtocol: google.StrPtr("udp"),
					},
					{
						IPProtocol: google.StrPtr("icmp"),
					},
					{
						IPProtocol: google.StrPtr("sctp"),
					},
					{
						IPProtocol: google.StrPtr("esp"),
					},
					{
						IPProtocol: google.StrPtr("ah"),
					},
				},
				Direction: "EGRESS",
				DestinationRanges: []string{
					masterIpv4CidrBlock,
					nodesCidr,
					podCidr,
				},
				TargetTags: []string{
					"default-pool",
				},
			},
			{
				Name:      fmt.Sprintf("%s-webhooks", name),
				ProjectID: projectID,
				Allowed: []*computepb.Allowed{
					{
						IPProtocol: google.StrPtr("tcp"),
						Ports: []string{
							"8443",
							"9443",
							"15017",
						},
					},
				},
				Direction: "INGRESS",
				SourceRanges: []string{
					masterIpv4CidrBlock,
				},
				TargetTags: []string{
					"default-pool",
				},
			},
		},
	}
	return &cp, nil
}
