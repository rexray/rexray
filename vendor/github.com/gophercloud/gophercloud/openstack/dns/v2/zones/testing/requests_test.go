package testing

import (
	"testing"

	"github.com/gophercloud/gophercloud/openstack/dns/v2/zones"
	"github.com/gophercloud/gophercloud/pagination"
	th "github.com/gophercloud/gophercloud/testhelper"
	"github.com/gophercloud/gophercloud/testhelper/client"
)

func TestList(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleListSuccessfully(t)

	count := 0
	err := zones.List(client.ServiceClient(), nil).EachPage(func(page pagination.Page) (bool, error) {
		count++
		actual, err := zones.ExtractZones(page)
		th.AssertNoErr(t, err)
		th.CheckDeepEquals(t, ExpectedZonesSlice, actual)

		return true, nil
	})
	th.AssertNoErr(t, err)
	th.CheckEquals(t, 1, count)
}

func TestListAllPages(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleListSuccessfully(t)

	allPages, err := zones.List(client.ServiceClient(), nil).AllPages()
	th.AssertNoErr(t, err)
	allZones, err := zones.ExtractZones(allPages)
	th.AssertNoErr(t, err)
	th.CheckEquals(t, 2, len(allZones))
}

func TestGet(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleGetSuccessfully(t)

	actual, err := zones.Get(client.ServiceClient(), "a86dba58-0043-4cc6-a1bb-69d5e86f3ca3").Extract()
	th.AssertNoErr(t, err)
	th.CheckDeepEquals(t, &FirstZone, actual)
}
