package tests

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
)

type handlerFunc func(http.ResponseWriter, *http.Request)
type mockVboxServer struct {
	server   *httptest.Server
	handlers map[string]handlerFunc
}

func newMockVBoxServer() *mockVboxServer {
	return &mockVboxServer{
		handlers: make(map[string]handlerFunc),
	}
}

func (vbox *mockVboxServer) withHandler(
	soapMessage string, handler handlerFunc) *mockVboxServer {
	vbox.handlers[soapMessage] = handler
	return vbox
}

func (vbox *mockVboxServer) start() {
	vbox.setup()
	vbox.server = httptest.NewUnstartedServer(
		http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			if req.ContentLength == 0 {
				resp.WriteHeader(http.StatusBadRequest)
				return
			}
			payload := new(bytes.Buffer)
			defer req.Body.Close()
			if _, err := io.Copy(payload, req.Body); err != nil {
				resp.WriteHeader(http.StatusInternalServerError)
				return
			}
			bodyStr := payload.String()
			for k, v := range vbox.handlers {
				if strings.Contains(bodyStr, k) {
					v(resp, req)
				}
			}
			//resp.WriteHeader(http.StatusNotFound)
		}),
	)
	vbox.server.Start()
}

func (vbox *mockVboxServer) stop() {
	vbox.server.Close()
}

func (vbox *mockVboxServer) url() string {
	return vbox.server.URL
}

func (vbox *mockVboxServer) setup() {
	var (
		envStr = `<SOAP-ENV:Envelope
		xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/"
		xmlns:SOAP-ENC="http://schemas.xmlsoap.org/soap/encoding/"
		xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
		xmlns:xsd="http://www.w3.org/2001/XMLSchema"
		xmlns:vbox="http://www.virtualbox.org/">
		<SOAP-ENV:Body>%s</SOAP-ENV:Body>
	    </SOAP-ENV:Envelope>`
	)

	vbox.withHandler(
		"IWebsessionManager_logon",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IWebsessionManager_logonResponse>
						<returnval>10d3b34eb3c6449f-0000000000000001</returnval>
					</vbox:IWebsessionManager_logonResponse>`)),
			)
		},
	).withHandler(
		"IVirtualBox_findMachine",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IVirtualBox_findMachineResponse>
					<returnval>714738d6a5c5f6f0-0000000000000002</returnval>
					</vbox:IVirtualBox_findMachineResponse>`)),
			)
		},
	).withHandler(
		"IMachine_getId",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IMachine_getIdResponse>
					<returnval>9f49850d-f617-4b43-a46d-272c380e7cc6</returnval>
				</vbox:IMachine_getIdResponse>`)),
			)
		},
	).withHandler(
		"IMachine_getName",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IMachine_getNameResponse>
						<returnval>default</returnval>
					</vbox:IMachine_getNameResponse>`)),
			)
		},
	).withHandler(
		"IVirtualBox_getMachines",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IVirtualBox_getMachinesResponse>
						<returnval>714738d6a5c5f6f0-0000000000000002</returnval>
						<returnval>714738d6a5c5f6f0-0000000000000003</returnval>
					</vbox:IVirtualBox_getMachinesResponse>`)),
			)
		},
	).withHandler(
		"IVirtualBox_getSystemProperties",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IVirtualBox_getSystemPropertiesResponse>
						<returnval>196c60f58ef29b42-0000000000000002</returnval>
					</vbox:IVirtualBox_getSystemPropertiesResponse>`)),
			)
		},
	).withHandler(
		"IManagedObjectRef_release",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IManagedObjectRef_releaseResponse/>`)),
			)
		},
	).withHandler(
		"IMachine_getChipsetType",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IMachine_getChipsetTypeResponse>
						<returnval>PIIX3</returnval>
					</vbox:IMachine_getChipsetTypeResponse>`)),
			)
		},
	).withHandler(
		"ISystemProperties_getMaxNetworkAdapters",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:ISystemProperties_getMaxNetworkAdaptersResponse>
						<returnval>0</returnval>
					</vbox:ISystemProperties_getMaxNetworkAdaptersResponse>`)),
			)
		},
	).withHandler(
		"IMachine_getNetworkAdapter",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IMachine_getNetworkAdapterResponse>
						<returnval>365a27c24ae0ca53-0000000000000005</returnval>
					</vbox:IMachine_getNetworkAdapterResponse>`)),
			)
		},
	).withHandler(
		"IVirtualBox_getHardDisks",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IVirtualBox_getHardDisksResponse>
			      		<returnval>329320f6f03e2476-0000000000000004</returnval>
			      		<returnval>329320f6f03e2476-0000000000000005</returnval>
			    	</vbox:IVirtualBox_getHardDisksResponse>`)),
			)
		},
	).withHandler(
		"IMedium_getName",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IMedium_getNameResponse>
		      			<returnval>disk.vmdk</returnval>
		    		</vbox:IMedium_getNameResponse>`)),
			)
		},
	).withHandler(
		"IMedium_getId",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IMedium_getIdResponse>
		      		<returnval>32a50c6d-ddcc-4e0a-a3c6-5c126a5f3f2c</returnval>
		    		</vbox:IMedium_getIdResponse>`)),
			)
		},
	).withHandler(
		"IMachine_getMediumAttachments",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IMachine_getMediumAttachmentsResponse>
				      <returnval>
				        <medium>3057518e9ce2b2ad-0000000000000004</medium>
				        <controller>SATA</controller>
				        <port>0</port>
				        <device>0</device>
				        <type>DVD</type>
				        <passthrough>false</passthrough>
				        <temporaryEject>false</temporaryEject>
				        <isEjected>false</isEjected>
				        <nonRotational>false</nonRotational>
				        <discard>false</discard>
				        <hotPluggable>false</hotPluggable>
				        <bandwidthGroup/>
				      </returnval>
				      <returnval>
				        <medium>3057518e9ce2b2ad-0000000000000005</medium>
				        <controller>SATA</controller>
				        <port>1</port>
				        <device>0</device>
				        <type>HardDisk</type>
				        <passthrough>false</passthrough>
				        <temporaryEject>false</temporaryEject>
				        <isEjected>false</isEjected>
				        <nonRotational>false</nonRotational>
				        <discard>false</discard>
				        <hotPluggable>false</hotPluggable>
				        <bandwidthGroup/>
				      </returnval>
		    		</vbox:IMachine_getMediumAttachmentsResponse>`)),
			)
		},
	).withHandler(
		"IMedium_getLocation",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IMedium_getLocationResponse>
		      			<returnval>/home/user/machines/disk.vmdk</returnval>
		    		</vbox:IMedium_getLocationResponse>`)),
			)
		},
	).withHandler(
		"IMedium_getDescription",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IMedium_getDescriptionResponse>
		      			<returnval/>
		    		</vbox:IMedium_getDescriptionResponse>`)),
			)
		},
	).withHandler(
		"IMedium_getLogicalSize",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IMedium_getLogicalSizeResponse>
      					<returnval>20971520000</returnval>
    				</vbox:IMedium_getLogicalSizeResponse>`),
			))
		},
	).withHandler(
		"IMedium_getDeviceType",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IMedium_getDeviceTypeResponse>
      					<returnval>HardDisk</returnval>
    				</vbox:IMedium_getDeviceTypeResponse>`),
			))
		},
	).withHandler(
		"IMedium_getFormat",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IMedium_getFormatResponse>
      				<returnval>VMDK</returnval>
    				</vbox:IMedium_getFormatResponse>`),
			))
		},
	).withHandler(
		"IMedium_getMediumFormat",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IMedium_getMediumFormatResponse>
      				<returnval>454bed3b394ff1e3-0000000000000004</returnval>
    				</vbox:IMedium_getMediumFormatResponse>`),
			))
		},
	).withHandler(
		"IMedium_getHostDrive",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IMedium_getHostDriveResponse>
      				<returnval>false</returnval>
    				</vbox:IMedium_getHostDriveResponse>`),
			))
		},
	).withHandler(
		"IMedium_getChildren",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IMedium_getChildrenResponse/>`),
			))
		},
	).withHandler(
		"IMedium_getParent",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IMedium_getParentResponse>
      				 <returnval/>
					 </vbox:IMedium_getParentResponse>`),
			))
		},
	).withHandler(
		"IMedium_getMachineIds",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IMedium_getMachineIdsResponse>
					<returnval>50b39738-53c7-4757-bc76-939d0182e001</returnval>
					</vbox:IMedium_getMachineIdsResponse>`),
			))
		},
	).withHandler(
		"IMedium_getSnapshotIds",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IMedium_getSnapshotIdsResponse/>`),
			))
		},
	).withHandler(
		"IVirtualBox_createMedium",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IVirtualBox_createMediumResponse>
					<returnval>e29b088928eebe30-0000000000000003</returnval>
					</vbox:IVirtualBox_createMediumResponse>`),
			))
		},
	).withHandler(
		"IMedium_createBaseStorage",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IMedium_createBaseStorageResponse>
					<returnval>e29b088928eebe30-0000000000000002</returnval>
					</vbox:IMedium_createBaseStorageResponse>`),
			))
		},
	).withHandler(
		"IProgress_waitForCompletion",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IProgress_waitForCompletionResponse/>`),
			))
		},
	).withHandler(
		"IProgress_getPercent",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IProgress_getPercentResponse>
					<returnval>100</returnval>
					</vbox:IProgress_getPercentResponse>`),
			))
		},
	).withHandler(
		"IMedium_deleteStorage",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IMedium_deleteStorageResponse>
					<returnval>454bed3b394ff1e3-0000000000000004</returnval>
					</vbox:IMedium_deleteStorageResponse>`),
			))
		},
	).withHandler(
		"IWebsessionManager_getSessionObject",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IWebsessionManager_getSessionObjectResponse>
					<returnval>c180b140c9b5ccc5-0000000000000003</returnval>
					</vbox:IWebsessionManager_getSessionObjectResponse>`),
			))
		},
	).withHandler(
		"IMachine_lockMachine",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IMachine_lockMachineResponse/>`),
			))
		},
	).withHandler(
		"ISession_unlockMachine",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:ISession_unlockMachineResponse/>`),
			))
		},
	).withHandler(
		"ISession_getMachine",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:ISession_getMachineResponse>
					<returnval>714738d6a5c5f6f0-0000000000000002</returnval>
					</vbox:ISession_getMachineResponse>`),
			))
		},
	).withHandler(
		"IMachine_getStorageControllers",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IMachine_getStorageControllersResponse>
					<returnval>144738d6a5c5f611-0000000000000002</returnval>
					<returnval>164738d6a5c5f6ab-0000000000000002</returnval>
					</vbox:IMachine_getStorageControllersResponse>`),
			))
		},
	).withHandler(
		"IStorageController_getName",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IStorageController_getNameResponse>
					<returnval>SATA</returnval>
					</vbox:IStorageController_getNameResponse>`),
			))
		},
	).withHandler(
		"IStorageController_getMaxPortCount",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IStorageController_getMaxPortCountResponse>
					<returnval>2</returnval>
					</vbox:IStorageController_getMaxPortCountResponse>`),
			))
		},
	).withHandler(
		"IMachine_getMediumAttachmentsOfController",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IMachine_getMediumAttachmentsOfControllerResponse>
					      <returnval>
					        <medium>37c53179330d22f5-0000000000000004</medium>
					        <controller>SATA</controller>
					        <port>0</port>
					        <device>0</device>
					        <type>DVD</type>
					        <passthrough>false</passthrough>
					        <temporaryEject>false</temporaryEject>
					        <isEjected>false</isEjected>
					        <nonRotational>false</nonRotational>
					        <discard>false</discard>
					        <hotPluggable>false</hotPluggable>
					        <bandwidthGroup/>
					      </returnval>
					      <returnval>
					        <medium>37c53179330d22f5-0000000000000005</medium>
					        <controller>SATA</controller>
					        <port>1</port>
					        <device>0</device>
					        <type>HardDisk</type>
					        <passthrough>false</passthrough>
					        <temporaryEject>false</temporaryEject>
					        <isEjected>false</isEjected>
					        <nonRotational>false</nonRotational>
					        <discard>false</discard>
					        <hotPluggable>false</hotPluggable>
					        <bandwidthGroup/>
					      </returnval>
    		  		</vbox:IMachine_getMediumAttachmentsOfControllerResponse>`),
			))
		},
	).withHandler(
		"IMachine_attachDevice",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IMachine_attachDeviceResponse/>`),
			))
		},
	).withHandler(
		"IMachine_detachDevice",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IMachine_detachDeviceResponse/>`),
			))
		},
	).withHandler(
		"IMachine_saveSettings",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IMachine_saveSettingsResponse/>`),
			))
		},
	).withHandler(
		"IMachine_discardSettings",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(
				fmt.Sprintf(envStr,
					`<vbox:IMachine_discardSettingsResponse/>`),
			))
		},
	)
}
