package utils

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testAddrInput struct {
	address []string
	ip      net.IP
}

var (
	ipZeroDotTwo    = net.ParseIP("192.168.0.2")
	ipOneDotTwo     = net.ParseIP("192.168.1.2")
	ipZeroDotTen    = net.ParseIP("192.168.0.10")
	ipZeroDotTwenty = net.ParseIP("192.168.0.20")
	ipZeroDotThirty = net.ParseIP("192.168.0.30")
	ipv6            = net.ParseIP("2001:db8:85a3::8a2e:370:7334")
	ipv6Local       = net.ParseIP("::1")
)

var testAddrs = []testAddrInput{
	{
		address: []string{"192.168.0.2"},
		ip:      ipZeroDotTwo,
	},
	{
		address: []string{
			"192.168.0.2",
			" 192.168.1.2"},
		ip: ipZeroDotTwo,
	},
	{
		address: []string{"192.168.0.2:6789"},
		ip:      ipZeroDotTwo,
	},
	{
		address: []string{
			"192.168.0.2:6789",
			" 192.168.1.2:6789"},
		ip: ipZeroDotTwo,
	},
	{
		address: []string{"[2001:db8:85a3::8a2e:370:7334]"},
		ip:      ipv6,
	},
	{
		address: []string{
			"[2001:db8:85a3::8a2e:370:7334]",
			" [2001:db8:85a3::8a2e:370:7335] "},
		ip: ipv6,
	},
	{
		address: []string{"[2001:db8:85a3::8a2e:370:7334]:6789"},
		ip:      ipv6,
	},
	{
		address: []string{"[::1]"},
		ip:      ipv6Local,
	},
}

func TestParseMonitorAddresses(t *testing.T) {
	for _, test := range testAddrs {
		ip, err := ParseMonitorAddresses(test.address)

		assert.NoError(t, err, "failed with %s", test.address[0])
		if err != nil {
			t.Error("failed TestParseMonitorAddresses")
			t.FailNow()
		}
		assert.True(t, test.ip.Equal(ip[0]))
	}
}

func TestArgsClient(t *testing.T) {
	tests := []struct {
		args    string
		present bool
	}{
		{
			args:    "--id myuser",
			present: true,
		},
		{
			args:    "--user myuser",
			present: true,
		},
		{
			args:    "-n client.myuser",
			present: true,
		},
		{
			args:    "--name client.myuser",
			present: true,
		},
		{
			args:    "--cluster cluster2",
			present: false,
		},
		{
			args:    "--cluster cluster2 --id ceph",
			present: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run("", func(st *testing.T) {
			st.Parallel()
			b := StrContainsClient(tt.args)
			if b != tt.present {
				t.Errorf("detection of client in cephArgs: %s incorrect, got: %v want: %v",
					tt.args, b, tt.present)
			}
		})
	}
}
