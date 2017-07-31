package v2

import (
	"testing"

	"github.com/gophercloud/gophercloud/acceptance/clients"
	"github.com/gophercloud/gophercloud/openstack/sharedfilesystems/v2/securityservices"
)

func TestSecurityServiceCreateDelete(t *testing.T) {
	client, err := clients.NewSharedFileSystemV2Client()
	if err != nil {
		t.Fatalf("Unable to create shared file system client: %v", err)
	}

	securityService, err := CreateSecurityService(t, client)
	if err != nil {
		t.Fatalf("Unable to create security service: %v", err)
	}

	PrintSecurityService(t, securityService)

	defer DeleteSecurityService(t, client, securityService)
}

func TestSecurityServiceList(t *testing.T) {
	client, err := clients.NewSharedFileSystemV2Client()
	if err != nil {
		t.Fatalf("Unable to create a shared file system client: %v", err)
	}

	allPages, err := securityservices.List(client, securityservices.ListOpts{}).AllPages()
	if err != nil {
		t.Fatalf("Unable to retrieve security services: %v", err)
	}

	allSecurityServices, err := securityservices.ExtractSecurityServices(allPages)
	if err != nil {
		t.Fatalf("Unable to extract security services: %v", err)
	}

	for _, securityService := range allSecurityServices {
		PrintSecurityService(t, &securityService)
	}
}

// The test creates 2 security services and verifies that only the one(s) with
// a particular name are being listed
func TestSecurityServiceListFiltering(t *testing.T) {
	client, err := clients.NewSharedFileSystemV2Client()
	if err != nil {
		t.Fatalf("Unable to create a shared file system client: %v", err)
	}

	securityService, err := CreateSecurityService(t, client)
	if err != nil {
		t.Fatalf("Unable to create security service: %v", err)
	}
	defer DeleteSecurityService(t, client, securityService)

	securityService, err = CreateSecurityService(t, client)
	if err != nil {
		t.Fatalf("Unable to create security service: %v", err)
	}
	defer DeleteSecurityService(t, client, securityService)

	options := securityservices.ListOpts{
		Name: securityService.Name,
	}

	allPages, err := securityservices.List(client, options).AllPages()
	if err != nil {
		t.Fatalf("Unable to retrieve security services: %v", err)
	}

	allSecurityServices, err := securityservices.ExtractSecurityServices(allPages)
	if err != nil {
		t.Fatalf("Unable to extract security services: %v", err)
	}

	for _, listedSecurityService := range allSecurityServices {
		if listedSecurityService.Name != securityService.Name {
			t.Fatalf("The name of the security service was expected to be %s", securityService.Name)
		}
		PrintSecurityService(t, &listedSecurityService)
	}
}
