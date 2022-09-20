package google

import (
	"context"

	compute "cloud.google.com/go/compute/apiv1"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

type Subnetwork struct {
	Name         string
	ProjectID    string
	Region       string
	NodesCidr    string
	PodsCidr     string
	ServicesCidr string
}

func (s *Subnetwork) create(ctx context.Context, client *compute.SubnetworksClient, network *string) (*computepb.Subnetwork, error) {
	if s.exists(ctx, client) {
		return s.get(ctx, client)
	}
	req := &computepb.InsertSubnetworkRequest{
		SubnetworkResource: &computepb.Subnetwork{
			IpCidrRange:           &s.NodesCidr,
			Name:                  &s.Name,
			Region:                &s.Region,
			Network:               network,
			PrivateIpGoogleAccess: boolPtr(true),
			SecondaryIpRanges: []*computepb.SubnetworkSecondaryRange{
				{
					IpCidrRange: &s.PodsCidr,
					RangeName:   strPtr("pods"),
				},
				{
					IpCidrRange: &s.ServicesCidr,
					RangeName:   strPtr("services"),
				},
			},
		},
		Project: s.ProjectID,
		Region:  s.Region,
	}
	op, err := client.Insert(ctx, req)
	if err != nil {
		return nil, err
	}
	err = op.Wait(ctx)
	if err != nil {
		return nil, err
	}
	return s.get(ctx, client)
}

func (s *Subnetwork) get(ctx context.Context, client *compute.SubnetworksClient) (*computepb.Subnetwork, error) {
	req := &computepb.GetSubnetworkRequest{
		Project:    s.ProjectID,
		Subnetwork: s.Name,
		Region:     s.Region,
	}
	resp, err := client.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *Subnetwork) exists(ctx context.Context, client *compute.SubnetworksClient) bool {
	_, err := s.get(ctx, client)
	return err == nil
}
