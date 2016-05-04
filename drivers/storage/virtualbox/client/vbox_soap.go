package client

import "encoding/xml"

type envelope struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope,omniempty"`
	Body    struct {
		XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body,omniempty"`
		Payload []byte   `xml:",innerxml"`
		Fault   *fault   `xml:",omitempty"`
	}
}

type fault struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault"`

	Code   string `xml:"faultcode,omitempty"`
	String string `xml:"faultstring,omitempty"`
	Actor  string `xml:"faultactor,omitempty"`
	Detail string `xml:"detail,omitempty"`
}

type logonRequest struct {
	XMLName  xml.Name `xml:"http://www.virtualbox.org/ IWebsessionManager_logon"`
	Username string   `xml:"username,omitempty"`
	Password string   `xml:"password,omitempty"`
}

type logonResponse struct {
	XMLName   xml.Name `xml:"IWebsessionManager_logonResponse"`
	Returnval string   `xml:"returnval,omitempty"`
}

type findMachineRequest struct {
	XMLName  xml.Name `xml:"http://www.virtualbox.org/ IVirtualBox_findMachine"`
	VbID     string   `xml:"_this,omitempty"`
	NameOrID string   `xml:"nameOrId,omitempty"`
}

type findMachineResponse struct {
	XMLName   xml.Name `xml:"IVirtualBox_findMachineResponse"`
	Returnval string   `xml:"returnval,omitempty"`
}

type getMachineIDRequest struct {
	XMLName xml.Name `xml:"http://www.virtualbox.org/ IMachine_getId"`
	Mobref  string   `xml:"_this,omitempty"`
}

type getMachineIDResponse struct {
	XMLName   xml.Name `xml:"IMachine_getIdResponse"`
	Returnval string   `xml:"returnval,omitempty"`
}

type getMachineNameRequest struct {
	XMLName xml.Name `xml:"http://www.virtualbox.org/ IMachine_getName"`
	Mobref  string   `xml:"_this,omitempty"`
}

type getMachineNameResponse struct {
	XMLName   xml.Name `xml:"IMachine_getNameResponse"`
	Returnval string   `xml:"returnval,omitempty"`
}

type getMachinesRequest struct {
	XMLName xml.Name `xml:"http://www.virtualbox.org/ IVirtualBox_getMachines"`
	MobRef  string   `xml:"_this,omitempty"`
}

type getMachinesResponse struct {
	XMLName   xml.Name `xml:"http://www.virtualbox.org/ IVirtualBox_getMachinesResponse"`
	Returnval []string `xml:"returnval,omitempty"`
}

type mediumAttachment struct {
	XMLName        xml.Name `xml:"http://www.virtualbox.org/ IMediumAttachment"`
	Medium         string   `xml:"medium,omitempty"`
	Controller     string   `xml:"controller,omitempty"`
	Port           int32    `xml:"port,omitempty"`
	Device         int32    `xml:"device,omitempty"`
	Type           string   `xml:"type,omitempty"`
	Passthrough    bool     `xml:"passthrough,omitempty"`
	TemporaryEject bool     `xml:"temporaryEject,omitempty"`
	IsEjected      bool     `xml:"isEjected,omitempty"`
	NonRotational  bool     `xml:"nonRotational,omitempty"`
	Discard        bool     `xml:"discard,omitempty"`
	HotPluggable   bool     `xml:"hotPluggable,omitempty"`
	BandwidthGroup string   `xml:"bandwidthGroup,omitempty"`
}

type getMediumAttachmentsRequest struct {
	XMLName xml.Name `xml:"http://www.virtualbox.org/ IMachine_getMediumAttachments"`
	Mobref  string   `xml:"_this,omitempty"`
}

type getMediumAttachmentsResponse struct {
	XMLName   xml.Name            `xml:"IMachine_getMediumAttachmentsResponse"`
	Returnval []*mediumAttachment `xml:"returnval,omitempty"`
}
