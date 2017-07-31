package goscaleio

import (
	. "gopkg.in/check.v1"
)

func (s *S) Test_FindCatalogItem(c *C) {

	// // Get the Org populated
	// testServer.Response(200, nil, orgExample)
	// org, err := s.vdc.GetVDCOrg()
	// _ = testServer.WaitRequest()
	// testServer.Flush()
	// c.Assert(err, IsNil)
	//
	// // Populate Catalog
	// testServer.Response(200, nil, catalogExample)
	// cat, err := org.FindCatalog("Public Catalog")
	// _ = testServer.WaitRequest()
	// testServer.Flush()
	//
	// // Find Catalog Item
	// testServer.Response(200, nil, catalogitemExample)
	// catitem, err := cat.FindCatalogItem("CentOS64-32bit")
	// _ = testServer.WaitRequest()
	// testServer.Flush()
	//
	// c.Assert(err, IsNil)
	// c.Assert(catitem.CatalogItem.HREF, Equals, "http://localhost:4444/api/catalogItem/1176e485-8858-4e15-94e5-ae4face605ae")
	// c.Assert(catitem.CatalogItem.Description, Equals, "id: cts-6.4-32bit")
	//
	// // Test non-existant catalog item
	// catitem, err = cat.FindCatalogItem("INVALID")
	// c.Assert(err, NotNil)

}

var systemResponse = `
[
  {"mdmMode":"Cluster",
  "mdmClusterState":"ClusteredNormal",
  "secondaryMdmActorIpList":["192.168.50.13"],
  "installId":"3500e7df2592947e",
  "primaryMdmActorIpList":["192.168.50.12"],
  "systemVersionName":"EMC ScaleIO Version: R1_31.1277.3",
  "capacityAlertHighThresholdPercent":80,
  "capacityAlertCriticalThresholdPercent":90,
  "remoteReadOnlyLimitState":false,
  "primaryMdmActorPort":9011,
  "secondaryMdmActorPort":9011,
  "tiebreakerMdmActorPort":9011,
  "mdmManagementPort":6611,
  "tiebreakerMdmIpList":["192.168.50.11"],
  "mdmManagementIpList":["192.168.50.12"],
  "defaultIsVolumeObfuscated":false,
  "restrictedSdcModeEnabled":false,
  "swid":"",
  "daysInstalled":1,
  "maxCapacityInGb":"Unlimited",
  "capacityTimeLeftInDays":"29",
  "enterpriseFeaturesEnabled":true,
  "isInitialLicense":true,
  "name":"cluster1",
  "id":"38a6603e69c6b8b1",
  "links":[
    {"rel":"self","href":"/api/instances/System::38a6603e69c6b8b1"},
    {"rel":"/api/System/relationship/Statistics","href":"/api/instances/System::38a6603e69c6b8b1/relationships/Statistics"},
    {"rel":"/api/System/relationship/User","href":"/api/instances/System::38a6603e69c6b8b1/relationships/User"},
    {"rel":"/api/System/relationship/ScsiInitiator","href":"/api/instances/System::38a6603e69c6b8b1/relationships/ScsiInitiator"},
    {"rel":"/api/System/relationship/ProtectionDomain","href":"/api/instances/System::38a6603e69c6b8b1/relationships/ProtectionDomain"},
    {"rel":"/api/System/relationship/Sdc","href":"/api/instances/System::38a6603e69c6b8b1/relationships/Sdc"}
  ]}
]
`
