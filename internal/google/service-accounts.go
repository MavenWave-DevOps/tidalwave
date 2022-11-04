package google

import (
	"context"
	"fmt"

	admin "cloud.google.com/go/iam/admin/apiv1"
	adminpb "google.golang.org/genproto/googleapis/iam/admin/v1"
)

// Service Account represents a IAM service account
type ServiceAccount struct {
	Name        string
	ProjectID   string
	DisplayName string
	Description string
}

// Create Service Account
func (s *ServiceAccount) create(ctx context.Context, client *admin.IamClient) (*adminpb.ServiceAccount, error) {
	if s.exists(ctx, client) {
		return s.get(ctx, client)
	}
	req := &adminpb.CreateServiceAccountRequest{
		Name:      fmt.Sprintf("projects/%s", s.ProjectID),
		AccountId: s.Name,
		ServiceAccount: &adminpb.ServiceAccount{
			DisplayName: s.DisplayName,
			Description: s.Description,
		},
	}
	resp, err := client.CreateServiceAccount(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Get Service Account
func (s *ServiceAccount) get(ctx context.Context, client *admin.IamClient) (*adminpb.ServiceAccount, error) {
	req := &adminpb.GetServiceAccountRequest{
		Name: fmt.Sprintf("projects/%s/serviceAccounts/%s", s.ProjectID, s.Name),
	}
	resp, err := client.GetServiceAccount(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Check if Service Account exists
func (s *ServiceAccount) exists(ctx context.Context, client *admin.IamClient) bool {
	_, err := s.get(ctx, client)
	return err == nil
}

// Delete Service Account
func (s *ServiceAccount) delete(ctx context.Context, client *admin.IamClient) error {
	if s.exists(ctx, client) {
		req := &adminpb.DeleteServiceAccountRequest{
			Name: fmt.Sprintf("projects/%s/serviceAccounts/%s", s.ProjectID, s.Name),
		}
		if err := client.DeleteServiceAccount(ctx, req); err != nil {
			return err
		}
	}
	return nil
}
