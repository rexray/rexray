package testing

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/dns/v2/zones"
	th "github.com/gophercloud/gophercloud/testhelper"
	"github.com/gophercloud/gophercloud/testhelper/client"
)

// List Output is a sample response to a List call.
const ListOutput = `
{
    "links": {
      "self": "http://example.com:9001/v2/zones"
    },
    "metadata": {
      "total_count": 2
    },
    "zones": [
        {
            "id": "a86dba58-0043-4cc6-a1bb-69d5e86f3ca3",
            "pool_id": "572ba08c-d929-4c70-8e42-03824bb24ca2",
            "project_id": "4335d1f0-f793-11e2-b778-0800200c9a66",
            "name": "example.org.",
            "email": "joe@example.org",
            "ttl": 7200,
            "serial": 1404757531,
            "status": "ACTIVE",
            "action": "CREATE",
            "description": "This is an example zone.",
            "masters": [],
            "type": "PRIMARY",
            "transferred_at": null,
            "version": 1,
            "created_at": "2014-07-07T18:25:31.275934",
            "updated_at": null,
            "links": {
              "self": "https://127.0.0.1:9001/v2/zones/a86dba58-0043-4cc6-a1bb-69d5e86f3ca3"
            }
        },
        {
            "id": "34c4561c-9205-4386-9df5-167436f5a222",
            "pool_id": "572ba08c-d929-4c70-8e42-03824bb24ca2",
            "project_id": "4335d1f0-f793-11e2-b778-0800200c9a66",
            "name": "foo.example.com.",
            "email": "joe@foo.example.com",
            "ttl": 7200,
            "serial": 1488053571,
            "status": "ACTIVE",
            "action": "CREATE",
            "description": "This is another example zone.",
            "masters": ["example.com."],
            "type": "PRIMARY",
            "transferred_at": null,
            "version": 1,
            "created_at": "2014-07-07T18:25:31.275934",
            "updated_at": "2015-02-25T20:23:01.234567",
            "links": {
              "self": "https://127.0.0.1:9001/v2/zones/34c4561c-9205-4386-9df5-167436f5a222"
            }
        }
    ]
}
`

// GetOutput is a sample response to a Get call.
const GetOutput = `
{
    "id": "a86dba58-0043-4cc6-a1bb-69d5e86f3ca3",
    "pool_id": "572ba08c-d929-4c70-8e42-03824bb24ca2",
    "project_id": "4335d1f0-f793-11e2-b778-0800200c9a66",
    "name": "example.org.",
    "email": "joe@example.org",
    "ttl": 7200,
    "serial": 1404757531,
    "status": "ACTIVE",
    "action": "CREATE",
    "description": "This is an example zone.",
    "masters": [],
    "type": "PRIMARY",
    "transferred_at": null,
    "version": 1,
    "created_at": "2014-07-07T18:25:31.275934",
    "updated_at": null,
    "links": {
      "self": "https://127.0.0.1:9001/v2/zones/a86dba58-0043-4cc6-a1bb-69d5e86f3ca3"
    }
}
`

// FirstZone is the first result in ListOutput
var FirstZoneCreatedAt, _ = time.Parse(gophercloud.RFC3339MilliNoZ, "2014-07-07T18:25:31.275934")
var FirstZone = zones.Zone{
	ID:          "a86dba58-0043-4cc6-a1bb-69d5e86f3ca3",
	PoolID:      "572ba08c-d929-4c70-8e42-03824bb24ca2",
	ProjectID:   "4335d1f0-f793-11e2-b778-0800200c9a66",
	Name:        "example.org.",
	Email:       "joe@example.org",
	TTL:         7200,
	Serial:      1404757531,
	Status:      "ACTIVE",
	Action:      "CREATE",
	Description: "This is an example zone.",
	Masters:     []string{},
	Type:        "PRIMARY",
	Version:     1,
	CreatedAt:   FirstZoneCreatedAt,
	Links: map[string]interface{}{
		"self": "https://127.0.0.1:9001/v2/zones/a86dba58-0043-4cc6-a1bb-69d5e86f3ca3",
	},
}

var SecondZoneCreatedAt, _ = time.Parse(gophercloud.RFC3339MilliNoZ, "2014-07-07T18:25:31.275934")
var SecondZoneUpdatedAt, _ = time.Parse(gophercloud.RFC3339MilliNoZ, "2015-02-25T20:23:01.234567")
var SecondZone = zones.Zone{
	ID:          "34c4561c-9205-4386-9df5-167436f5a222",
	PoolID:      "572ba08c-d929-4c70-8e42-03824bb24ca2",
	ProjectID:   "4335d1f0-f793-11e2-b778-0800200c9a66",
	Name:        "foo.example.com.",
	Email:       "joe@foo.example.com",
	TTL:         7200,
	Serial:      1488053571,
	Status:      "ACTIVE",
	Action:      "CREATE",
	Description: "This is another example zone.",
	Masters:     []string{"example.com."},
	Type:        "PRIMARY",
	Version:     1,
	CreatedAt:   SecondZoneCreatedAt,
	UpdatedAt:   SecondZoneUpdatedAt,
	Links: map[string]interface{}{
		"self": "https://127.0.0.1:9001/v2/zones/34c4561c-9205-4386-9df5-167436f5a222",
	},
}

// ExpectedZonesSlice is the slice of results that should be parsed
// from ListOutput, in the expected order.
var ExpectedZonesSlice = []zones.Zone{FirstZone, SecondZone}

// HandleListSuccessfully configures the test server to respond to a List request.
func HandleListSuccessfully(t *testing.T) {
	th.Mux.HandleFunc("/zones", func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "GET")
		th.TestHeader(t, r, "X-Auth-Token", client.TokenID)

		w.Header().Add("Content-Type", "application/json")
		fmt.Fprintf(w, ListOutput)
	})
}

// HandleGetSuccessfully configures the test server to respond to a List request.
func HandleGetSuccessfully(t *testing.T) {
	th.Mux.HandleFunc("/zones/a86dba58-0043-4cc6-a1bb-69d5e86f3ca3", func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "GET")
		th.TestHeader(t, r, "X-Auth-Token", client.TokenID)

		w.Header().Add("Content-Type", "application/json")
		fmt.Fprintf(w, GetOutput)
	})
}
