package v2

import (
	"testing"

	"github.com/gophercloud/gophercloud/acceptance/clients"
	"github.com/gophercloud/gophercloud/openstack/sharedfilesystems/v2/sharetypes"
)

func TestShareTypeCreateDestroy(t *testing.T) {
	client, err := clients.NewSharedFileSystemV2Client()
	if err != nil {
		t.Fatalf("Unable to create shared file system client: %v", err)
	}

	shareType, err := CreateShareType(t, client)
	if err != nil {
		t.Fatalf("Unable to create share type: %v", err)
	}

	PrintShareType(t, shareType)

	defer DeleteShareType(t, client, shareType)
}

func TestShareTypeList(t *testing.T) {
	client, err := clients.NewSharedFileSystemV2Client()
	if err != nil {
		t.Fatalf("Unable to create a shared file system client: %v", err)
	}

	allPages, err := sharetypes.List(client, sharetypes.ListOpts{}).AllPages()
	if err != nil {
		t.Fatalf("Unable to retrieve share types: %v", err)
	}

	allShareTypes, err := sharetypes.ExtractShareTypes(allPages)
	if err != nil {
		t.Fatalf("Unable to extract share types: %v", err)
	}

	for _, shareType := range allShareTypes {
		PrintShareType(t, &shareType)
	}
}

func TestShareTypeGetDefault(t *testing.T) {
	client, err := clients.NewSharedFileSystemV2Client()
	if err != nil {
		t.Fatalf("Unable to create a shared file system client: %v", err)
	}

	shareType, err := sharetypes.GetDefault(client).Extract()
	if err != nil {
		t.Fatalf("Unable to retrieve the default share type: %v", err)
	}

	if shareType.Name != "default" {
		t.Fatal("Share type name was expected to be: default")
	}

	PrintShareType(t, shareType)
}

func TestShareTypeExtraSpecs(t *testing.T) {
	client, err := clients.NewSharedFileSystemV2Client()
	if err != nil {
		t.Fatalf("Unable to create shared file system client: %v", err)
	}

	shareType, err := CreateShareType(t, client)
	if err != nil {
		t.Fatalf("Unable to create share type: %v", err)
	}

	options := sharetypes.SetExtraSpecsOpts{
		Specs: map[string]interface{}{"my_new_key": "my_value"},
	}

	_, err = sharetypes.SetExtraSpecs(client, shareType.ID, options).Extract()
	if err != nil {
		t.Fatalf("Unable to set extra specs for Share type: %s", shareType.Name)
	}

	extraSpecs, err := sharetypes.GetExtraSpecs(client, shareType.ID).Extract()
	if err != nil {
		t.Fatalf("Unable to retrieve share type: %s", shareType.Name)
	}

	if extraSpecs["driver_handles_share_servers"] != "True" {
		t.Fatal("driver_handles_share_servers was expected to be true")
	}

	if extraSpecs["my_new_key"] != "my_value" {
		t.Fatal("my_new_key was expected to be equal to my_value")
	}

	err = sharetypes.UnsetExtraSpecs(client, shareType.ID, "my_new_key").ExtractErr()
	if err != nil {
		t.Fatalf("Unable to unset extra specs for Share type: %s", shareType.Name)
	}

	extraSpecs, err = sharetypes.GetExtraSpecs(client, shareType.ID).Extract()
	if err != nil {
		t.Fatalf("Unable to retrieve share type: %s", shareType.Name)
	}

	if _, ok := extraSpecs["my_new_key"]; ok {
		t.Fatalf("my_new_key was expected to be unset for Share type: %s", shareType.Name)
	}

	PrintShareType(t, shareType)

	defer DeleteShareType(t, client, shareType)
}

func TestShareTypeShowAccess(t *testing.T) {
	client, err := clients.NewSharedFileSystemV2Client()
	if err != nil {
		t.Fatalf("Unable to create shared file system client: %v", err)
	}

	shareType, err := CreateShareType(t, client)
	if err != nil {
		t.Fatalf("Unable to create share type: %v", err)
	}

	_, err = sharetypes.ShowAccess(client, shareType.ID).Extract()
	if err != nil {
		t.Fatalf("Unable to retrieve the access details for a share type: %v", err)
	}

	PrintShareType(t, shareType)

	defer DeleteShareType(t, client, shareType)

}
