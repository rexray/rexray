package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	golog "log"
	"net"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/goof"
	"github.com/akutz/gotil"
	gcontext "github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc"
	gjson "github.com/gorilla/rpc/json"

	"github.com/emccode/libstorage/api"
	"github.com/emccode/libstorage/service/server/handlers"
)

var (
	noInstanceIDMethods = []string{
		"libStorage.GetServiceInfo",
		"libStorage.GetClientTool",
	}
)

// ServiceInfo is information used to serve a service.
type ServiceInfo struct {
	Name    string
	Service api.ServiceEndpoint
	Config  gofig.Config
	server  *rpc.Server
}

// Serve serves one or more services via HTTP/JSON.
func Serve(
	serviceInfo map[string]*ServiceInfo,
	config gofig.Config) (err error) {

	if err = initServers(serviceInfo, config); err != nil {
		return
	}

	var l net.Listener
	var host, laddr string
	if host, laddr, l, err = getNetInfo(config); err != nil {
		return
	}
	log.WithField("host", host).Debug("ready to listen")

	r := mux.NewRouter()
	httpHandler := http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			vars := mux.Vars(req)
			name := vars["name"]
			serviceInfo[name].server.ServeHTTP(w, req)
		})
	loggingHandler := handlers.NewLoggingHandler(httpHandler, config)
	r.Handle("/libStorage/{name}", gcontext.ClearHandler(loggingHandler))

	hs := &http.Server{
		Addr:           laddr,
		Handler:        r,
		ReadTimeout:    getReadTimeout(config) * time.Second,
		WriteTimeout:   getWriteTimeout(config) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if loggingHandler.Enabled {
		hs.ErrorLog = golog.New(loggingHandler.StdErr, "", 0)
	}

	go func() {
		defer func() {
			if err := loggingHandler.Close(); err != nil {
				log.Error(err)
			}

			r := recover()
			switch tr := r.(type) {
			case error:
				log.Panic(
					"unhandled exception when serving libStorage service", tr)
			}
		}()

		if err := hs.Serve(l); err != nil {
			log.Panic("error serving libStorage service", err)
		}
	}()

	updatedHost := fmt.Sprintf("%s://%s",
		l.Addr().Network(),
		l.Addr().String())

	if updatedHost != host {
		host = updatedHost
		config.Set("libstorage.host", host)
	}
	log.WithField("host", host).Debug("listening")

	return nil
}

func initServers(
	serviceInfo map[string]*ServiceInfo,
	config gofig.Config) error {

	for _, si := range serviceInfo {
		s := rpc.NewServer()
		s.RegisterBeforeFunc(func(i *rpc.RequestInfo) {
			if !gotil.StringInSlice(i.Method, noInstanceIDMethods) {
				initInstanceID(i.Request)
			}
		})

		codec := gjson.NewCodec()
		s.RegisterCodec(codec, "application/json")
		s.RegisterCodec(codec, "application/json;charset=UTF-8")

		if err := s.RegisterService(si.Service, "libStorage"); err != nil {
			return err
		}

		si.server = s
	}

	return nil
}

func getNetInfo(config gofig.Config) (
	host, laddr string,
	l net.Listener, err error) {

	host = config.GetString("libstorage.host")
	if host == "" {
		host = "tcp://127.0.0.1:0"
	}

	var netProto string
	if netProto, laddr, err = gotil.ParseAddress(host); err != nil {
		return
	}

	if l, err = net.Listen(netProto, laddr); err != nil {
		return
	}

	return
}

func initInstanceID(req *http.Request) {
	iidb64 := req.Header.Get("libStorage-InstanceID")
	if iidb64 == "" {
		panic(goof.New("instanceID required"))
	}

	iidJSON, err := base64.URLEncoding.DecodeString(iidb64)
	if err != nil {
		panic(err)
	}

	var iid *api.InstanceID
	err = json.Unmarshal(iidJSON, &iid)
	if err != nil {
		panic(err)
	}

	log.WithField("instanceID", iid).Debug("request's instance ID")

	gcontext.Set(req, "instanceID", iid)
}

func getReadTimeout(config gofig.Config) time.Duration {
	t := config.GetInt("libstorage.service.readtimeout")
	if t == 0 {
		return 60
	}
	return time.Duration(t)
}

func getWriteTimeout(config gofig.Config) time.Duration {
	t := config.GetInt("libstorage.service.writetimeout")
	if t == 0 {
		return 60
	}
	return time.Duration(t)
}
