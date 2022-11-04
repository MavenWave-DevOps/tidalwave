package google

import (
	"context"
	"fmt"
	"strings"

	resource "cloud.google.com/go/resourcemanager/apiv3"
	crmgr "google.golang.org/api/cloudresourcemanager/v1"
	resourcemanagerpb "google.golang.org/genproto/googleapis/cloud/resourcemanager/v3"
)

// GetProjectNumber returns a GCP project number from a GCP project id
func GetProjectNumber(id string) (*string, error) {
	ctx := context.Background()
	client, err := resource.NewProjectsClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	req := &resourcemanagerpb.GetProjectRequest{
		Name: fmt.Sprintf("projects/%s", id),
	}
	resp, err := client.GetProject(ctx, req)
	if err != nil {
		return nil, err
	}
	p := strings.Trim(resp.GetName(), "projects/")
	ptr := &p
	return ptr, nil
}

func SetProjectIam(client *crmgr.Service, projectID string, policy *crmgr.Policy) error {
	req := &crmgr.SetIamPolicyRequest{
		Policy: policy,
	}
	op := client.Projects.SetIamPolicy(projectID, req)
	_, err := op.Do()
	if err != nil {
		return err
	}
	return nil
}

func GetProjectPolicy(client *crmgr.Service, projectID string) (*crmgr.Policy, error) {
	op := client.Projects.GetIamPolicy(projectID, &crmgr.GetIamPolicyRequest{})
	policies, err := op.Do()
	if err != nil {
		return nil, err
	}
	return policies, nil
}

func AddProjectBinding(policy *crmgr.Policy, projectID, role, member string) *crmgr.Policy {
	var binding *crmgr.Binding
	for _, v := range policy.Bindings {
		if v.Role == role {
			binding = v
			break
		}
	}

	if binding != nil {
		binding.Members = append(binding.Members, member)
	} else {
		binding = &crmgr.Binding{
			Role:    role,
			Members: []string{member},
		}
		policy.Bindings = append(policy.Bindings, binding)
	}
	return policy
}

func RemoveProjectBinding(policy *crmgr.Policy, projectID, role, member string) *crmgr.Policy {
	var binding *crmgr.Binding
	var bindingIndex int
	for i, v := range policy.Bindings {
		if v.Role == role {
			binding = v
			bindingIndex = i
			break
		}
	}

	if len(binding.Members) == 1 {
		last := len(policy.Bindings) - 1
		policy.Bindings[bindingIndex] = policy.Bindings[last]
		policy.Bindings = policy.Bindings[:last]
	} else {
		var memberIndex int
		for i, mm := range binding.Members {
			if mm == member {
				memberIndex = i
			}
		}
		last := len(policy.Bindings[bindingIndex].Members) - 1
		binding.Members[memberIndex] = binding.Members[last]
		binding.Members = binding.Members[:last]
	}
	return policy
}
