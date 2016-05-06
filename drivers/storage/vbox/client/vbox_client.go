package client

import (
	"bytes"
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"net"
	"net/http"
	"time"
)

// VirtualBox Represents a virtualbox sesion
type VirtualBox struct {
	username     string
	password     string
	vbURL        string
	client       *http.Client
	useBasicAuth bool
	mobref       string
}

// NewVirtualBox returns a reference to a VirtualBox value.
func NewVirtualBox(uname, pwd, url string) *VirtualBox {
	vb := &VirtualBox{
		username: uname,
		password: pwd,
		vbURL:    url,
		client: &http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
			Dial: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   15 * time.Second,
			ExpectContinueTimeout: 3 * time.Second,
		},
		},
	}
	return vb
}

// WithTimeout sets connection timeout
func (vb *VirtualBox) WithTimeout(dur time.Duration) *VirtualBox {
	vb.client.Transport.(*http.Transport).Dial = (&net.Dialer{
		Timeout:   dur,
		KeepAlive: 30 * time.Second,
	}).Dial
	return vb
}

// UseBasicAuth Sets the use of basic-auth as true or false
func (vb *VirtualBox) UseBasicAuth(flag bool) *VirtualBox {
	vb.useBasicAuth = true
	return vb
}

func (vb *VirtualBox) send(request, response interface{}) error {
	// encode request
	payload, err := xml.Marshal(request)

	if err != nil {
		return err
	}

	reqData := new(bytes.Buffer)
	env := new(envelope)
	env.Body.Payload = payload
	err = xml.NewEncoder(reqData).Encode(env)

	if err != nil {
		return err
	}

	// send req as http
	httpReq, err := http.NewRequest("POST", vb.vbURL, reqData)
	if err != nil {
		return err
	}
	httpReq.Header.Add("Content-Type", "text/xml; charset=\"utf-8\"")
	httpReq.Header.Set("User-Agent", "libstorage/0.1")
	if vb.useBasicAuth {
		httpReq.SetBasicAuth(vb.username, vb.password)
	}

	httpResp, err := vb.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("Unable to do request: %s", err)
	}

	switch code := httpResp.StatusCode; {
	case code == http.StatusInternalServerError:
		if httpResp.ContentLength > 0 { // check for soap error
			env = new(envelope)
			if err = xml.NewDecoder(httpResp.Body).Decode(env); err != nil {
				return fmt.Errorf("Failed to decode SOAP fault: %s", err)
			}
			if f := env.Body.Fault; f != nil {
				return fmt.Errorf("SOAP Fault: %s: %s", f.String, f.Detail)
			}
		}
		return fmt.Errorf("Unexpected http status: %d", httpResp.StatusCode)

	}
	defer httpResp.Body.Close()

	// decode response
	env = new(envelope)
	err = xml.NewDecoder(httpResp.Body).Decode(env)
	if err != nil {
		return fmt.Errorf("Error decoding response: %s", err)
	}
	if f := env.Body.Fault; f != nil {
		return fmt.Errorf("Failed to decode response: %s: %s", f.String, f.Detail)
	}

	err = xml.Unmarshal(env.Body.Payload, response)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal payload: %s", err)
	}
	return nil
}

// Logon logs into the soap server.
func (vb *VirtualBox) Logon() error {
	request := logonRequest{
		Username: vb.username,
		Password: vb.password,
	}
	response := new(logonResponse)
	if err := vb.send(request, response); err != nil {
		return err
	}
	vb.mobref = response.Returnval
	return nil
}

// FindMachine finds a machine based on its name or machine id.
func (vb *VirtualBox) FindMachine(nameOrID string) (*Machine, error) {
	if err := vb.assertMobRef(); err != nil {
		return nil, err
	}

	request := findMachineRequest{VbID: vb.mobref, NameOrID: nameOrID}
	response := new(findMachineResponse)
	err := vb.send(request, response)
	if err != nil {
		return nil, err
	}

	m := &Machine{mobref: response.Returnval, vb: vb}
	return m, nil
}

//PopulateMachineInfo loads additional descriptive information for machine
func (vb *VirtualBox) PopulateMachineInfo(machine *Machine) error {
	if machine.mobref == "" {
		return fmt.Errorf("Machine missing object reference id")
	}
	// get machine id
	req1 := getMachineIDRequest{Mobref: machine.mobref}
	rsp1 := new(getMachineIDResponse)
	err := vb.send(req1, rsp1)
	if err != nil {
		return err
	}
	machine.id = rsp1.Returnval

	// get machine name
	req2 := getMachineNameRequest{Mobref: machine.mobref}
	rsp2 := new(getMachineNameResponse)
	err = vb.send(req2, rsp2)
	if err != nil {
		return err
	}
	machine.name = rsp2.Returnval

	return nil
}

// GetMachines returns all registered machines for the virtualbox
func (vb *VirtualBox) GetMachines() ([]*Machine, error) {
	if err := vb.assertMobRef(); err != nil {
		return nil, err
	}

	request := getMachinesRequest{MobRef: vb.mobref}
	response := new(getMachinesResponse)
	err := vb.send(request, response)
	if err != nil {
		return nil, err
	}
	machines := make([]*Machine, len(response.Returnval))
	for i, machineID := range response.Returnval {
		machines[i] = NewMachine(vb, machineID)
	}
	return machines, nil
}

func (vb *VirtualBox) assertMobRef() error {
	if vb.mobref == "" {
		return fmt.Errorf("Missing VirtualBox managed object id")
	}
	return nil
}
