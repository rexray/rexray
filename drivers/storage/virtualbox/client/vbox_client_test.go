package client

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	uname    = ""
	password = ""
)
const xmlEnvelope = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope
	xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/"
	xmlns:SOAP-ENC="http://schemas.xmlsoap.org/soap/encoding/"
	xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
	xmlns:xsd="http://www.w3.org/2001/XMLSchema"
	xmlns:vbox="http://www.virtualbox.org/">
	<SOAP-ENV:Body>%s</SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

type testTag struct {
	Value string `xml:"val"`
}

func TestNewVirtualBox(t *testing.T) {
	vb := NewVirtualBox("uname", "password", "http://test/")
	if vb.username != "uname" {
		t.Fatal("Username not set")
	}
	if vb.password != "password" {
		t.Fatal("Password not set")
	}
	if vb.vbURL != "http://test/" {
		t.Fatal("URL not set")
	}
}

func TestSend_NotOK(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusForbidden)
		}),
	)
	defer server.Close()
	vb := NewVirtualBox(uname, password, server.URL)
	resp := new(string)
	if err := vb.send("test", resp); err == nil {
		t.Fatal("Expected failure")
	}
}
func TestSend_OK(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			payload, err := xml.Marshal(&testTag{Value: "Test"})
			if err != nil {
				t.Fatal(err)
			}
			env := new(envelope)
			env.Body.Payload = payload
			xml.NewEncoder(resp).Encode(env)
		}),
	)
	defer server.Close()

	vb := NewVirtualBox(uname, password, server.URL)
	resp := new(testTag)
	if err := vb.send("test", resp); err != nil {
		t.Fatal("Unexpected failure:", err)
	}
	if resp.Value != "Test" {
		t.Fatal("Failed to process xml response properly")
	}
}

func TestLogon(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			// unmarshal request
			env := new(envelope)
			logon := new(logonRequest)
			if err := xml.NewDecoder(req.Body).Decode(env); err != nil {
				t.Fatal("Error decoding logonRequest", err)
			}
			if err := xml.Unmarshal(env.Body.Payload, logon); err != nil {
				t.Fatal("Error unmarshaling payload: ", err)
			}
			if logon.Username != uname {
				t.Fatal("Unexpected data from logonRequest")
			}
			// return response
			resp.WriteHeader(http.StatusOK)
			payload := `<vbox:IWebsessionManager_logonResponse>
			<returnval>000-test-000</returnval>
			</vbox:IWebsessionManager_logonResponse>`
			xmlResp := fmt.Sprintf(xmlEnvelope, payload)
			resp.Write([]byte(xmlResp))
		}),
	)
	defer server.Close()

	vb := NewVirtualBox(uname, password, server.URL)
	if err := vb.Logon(); err != nil {
		t.Fatal("Logon failed:", err)
	}
	if vb.mobref != "000-test-000" {
		t.Fatal("Failed to get session id from logon")
	}
}

func TestFindMachine(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			// unmarshal request
			env := new(envelope)
			machine := new(findMachineRequest)
			if err := xml.NewDecoder(req.Body).Decode(env); err != nil {
				t.Fatal("Error decoding findMachineREquest", err)
			}
			if err := xml.Unmarshal(env.Body.Payload, machine); err != nil {
				t.Fatal("Error unmarshaling payload: ", err)
			}
			if machine.NameOrID != "000-machine-000" {
				t.Fatal("Unexpected data from findMachineREquest")
			}
			// return response
			resp.WriteHeader(http.StatusOK)
			payload := `<vbox:IVirtualBox_findMachineResponse>
			<returnval>000-machine-000</returnval>
			</vbox:IVirtualBox_findMachineResponse>`
			xmlResp := fmt.Sprintf(xmlEnvelope, payload)
			resp.Write([]byte(xmlResp))
		}),
	)
	defer server.Close()

	vb := NewVirtualBox(uname, password, server.URL)
	vb.mobref = "000-test-000" // simulated logon
	m, err := vb.FindMachine("000-machine-000")
	if err != nil {
		t.Fatal(err)
	}
	if m == nil {
		t.Fatal("Machine should not be nil")
	}
	if m.mobref != "000-machine-000" {
		t.Fatal("Machine id not set properly")
	}
	if m.vb != vb {
		t.Fatal("Machine's vb reference not set")
	}
}
