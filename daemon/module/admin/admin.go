package admin

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	golog "log"
	"net/http"
	"os"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/gotil"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/emccode/rexray/daemon/module"
)

const (
	modName = "admin"
)

type mod struct {
	name string
	addr string
	desc string
}

type jsonError struct {
	Message string `json:"message"`
	Error   error  `json:"error"`
}

func init() {
	module.RegisterModule(modName, newModule)
}

func newModule(c *module.Config) (module.Module, error) {
	return &mod{
		name: c.Name,
		desc: c.Description,
		addr: c.Address,
	}, nil
}

func loadAsset(path, defaultValue string) string {

	devPath := fmt.Sprintf(
		"%s/src/github.com/emccode/rexray/daemon/module/admin/html/%s",
		os.Getenv("GOPATH"),
		path)

	if gotil.FileExists(devPath) {
		v, _ := ioutil.ReadFile(devPath)
		log.Printf("Loaded %s from %s\n", path, devPath)
		return string(v)
	}

	exeDir, _, _ := gotil.GetThisPathParts()

	relPath := fmt.Sprintf(
		"%s/html/%s",
		exeDir,
		path)

	if gotil.FileExists(relPath) {
		v, _ := ioutil.ReadFile(devPath)
		log.Printf("Loaded %s from %s\n", path, relPath)
		return string(v)
	}

	return defaultValue
}

func writeContentLength(w http.ResponseWriter, content string) {
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
}

func indexHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	fmt.Fprint(w, loadAsset("index.html", htmlIndex))
}

func scriptsHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/javascript; charset=UTF-8")
	a := loadAsset("scripts/jquery-1.11.3.min.js", scriptJQuery)
	writeContentLength(w, a)
	fmt.Fprint(w, a)
}

func stylesHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=UTF-8")
	a := loadAsset("styles/main.css", styleMain)
	writeContentLength(w, a)
	fmt.Fprint(w, a)
}

func imagesHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml; charset=UTF-8")
	a := loadAsset("images/rexray-banner-logo.svg", imageRexRayBannerLogo)
	writeContentLength(w, a)
	fmt.Fprint(w, a)
}

func moduleTypeHandler(w http.ResponseWriter, req *http.Request) {
	var mods []*module.Type
	for m := range module.Types() {
		mods = append(mods, m)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	jsonBuf, jsonBufErr := json.MarshalIndent(mods, "", "  ")
	if jsonBufErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Error servicing request ERR: %v", jsonBufErr)
		return
	}

	_, writeErr := w.Write(jsonBuf)
	if writeErr != nil {
		log.Printf("Error writing json buffer ERR: %v", writeErr)
	}
}

func moduleInstGetHandler(w http.ResponseWriter, req *http.Request) {
	var mods []*module.Instance
	for m := range module.Instances() {
		mods = append(mods, m)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	jsonBuf, jsonBufErr := json.MarshalIndent(mods, "", "  ")
	if jsonBufErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Error servicing request ERR: %v", jsonBufErr)
		return
	}

	_, writeErr := w.Write(jsonBuf)
	if writeErr != nil {
		log.Printf("Error writing json buffer ERR: %v", writeErr)
	}
}

func moduleInstPostHandler(w http.ResponseWriter, req *http.Request) {
	name := req.FormValue("name")
	typeName := req.FormValue("typeName")
	address := req.FormValue("address")
	cfgJSON := req.FormValue("config")
	start := req.FormValue("start")

	log.WithFields(log.Fields{
		"name":     name,
		"typeName": typeName,
		"address":  address,
		"start":    start,
		"config":   cfgJSON,
	}).Debug("received module instance post request")

	cfg, cfgErr := gofig.FromJSON(cfgJSON)
	if cfgErr != nil {
		w.Write(getJSONError("Error unmarshalling config json", nil))
		log.Printf("Error unmarshalling config json\n")
		return
	}

	modConfig := &module.Config{
		Name:    name,
		Type:    typeName,
		Address: address,
		Config:  cfg,
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if typeName == "" || address == "" {
		w.Write(getJSONError("Fields typeName and address are required", nil))
		log.Printf("Fields typeName and address are required\n")
		return
	}

	modInst, initErr := module.InitializeModule(modConfig)
	if initErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Error initializing module ERR: %v\n", initErr)
		return
	}

	jsonBuf, jsonBufErr := json.MarshalIndent(modInst, "", "  ")
	if jsonBufErr != nil {
		w.Write(getJSONError("Error marshalling object to json", jsonBufErr))
		log.Printf("Error marshalling object to json ERR: %v\n", jsonBufErr)
		return
	}

	startBool, startBoolErr := strconv.ParseBool(start)
	if startBoolErr != nil {
		startBool = false
	}

	if startBool {
		startErr := module.StartModule(modInst.Name)
		if startErr != nil {
			w.Write(getJSONError("Error starting module", startErr))
			log.Printf("Error starting module ERR: %v\n", startErr)
			return
		}

		jsonBufErr = nil
		jsonBuf, jsonBufErr = json.MarshalIndent(modInst, "", "  ")
		if jsonBufErr != nil {
			w.Write(getJSONError("Error marshalling object to json", jsonBufErr))
			log.Printf("Error marshalling object to json ERR: %v\n", jsonBufErr)
			return
		}
		w.Write(jsonBuf)
	} else {
		w.Write(jsonBuf)
	}
}

func moduleInstHandler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		moduleInstGetHandler(w, req)
	case "POST":
		moduleInstPostHandler(w, req)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func moduleInstStartHandler(w http.ResponseWriter, req *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(req)
	name := vars["name"]
	if name == "" {
		w.Write(getJSONError("The URL should include the module instance name", nil))
		log.Printf("The URL should include the module instance name\n")
		return
	}

	modInst, modInstErr := module.GetModuleInstance(name)
	if modInstErr != nil {
		w.Write(getJSONError("Unknown module name", modInstErr))
		log.Printf("Unknown module name ERR: %v\n", modInstErr)
		return
	}

	jsonBuf, jsonBufErr := json.MarshalIndent(modInst, "", "  ")
	if jsonBufErr != nil {
		w.Write(getJSONError("Error marshalling object to json", jsonBufErr))
		log.Printf("Error marshalling object to json ERR: %v\n", jsonBufErr)
		return
	}

	if modInst.IsStarted {
		w.Write(jsonBuf)
		return
	}

	startErr := module.StartModule(name)

	if startErr != nil {
		w.Write(getJSONError("Error starting moudle", startErr))
		log.Printf("Error starting module ERR: %v\n", startErr)
		return
	}

	jsonBufErr = nil
	jsonBuf, jsonBufErr = json.MarshalIndent(modInst, "", "  ")
	if jsonBufErr != nil {
		w.Write(getJSONError("Error marshalling object to json", jsonBufErr))
		log.Printf("Error marshalling object to json ERR: %v\n", jsonBufErr)
		return
	}

	w.Write(jsonBuf)
}

func getJSONError(msg string, err error) []byte {
	buf, marshalErr := json.MarshalIndent(
		&jsonError{
			Message: msg,
			Error:   err,
		}, "", "  ")
	if marshalErr != nil {
		panic(marshalErr)
	}
	return buf
}

func (m *mod) Start() error {
	stdOut := log.StandardLogger().Writer()
	stdErr := log.StandardLogger().Writer()

	r := mux.NewRouter()

	r.Handle("/r/module/instances",
		handlers.LoggingHandler(stdOut, http.HandlerFunc(moduleInstHandler)))
	r.Handle("/r/module/instances/{name}/start",
		handlers.LoggingHandler(stdOut, http.HandlerFunc(moduleInstStartHandler)))
	r.Handle("/r/module/types",
		handlers.LoggingHandler(stdOut, http.HandlerFunc(moduleTypeHandler)))

	r.Handle("/images/rexray-banner-logo.svg",
		handlers.LoggingHandler(stdOut, http.HandlerFunc(imagesHandler)))
	r.Handle("/scripts/jquery-1.11.3.min.js",
		handlers.LoggingHandler(stdOut, http.HandlerFunc(scriptsHandler)))
	r.Handle("/styles/main.css",
		handlers.LoggingHandler(stdOut, http.HandlerFunc(stylesHandler)))

	r.Handle("/",
		handlers.LoggingHandler(stdOut, http.HandlerFunc(indexHandler)))

	_, addr, parseAddrErr := gotil.ParseAddress(m.Address())
	if parseAddrErr != nil {
		return parseAddrErr
	}

	s := &http.Server{
		Addr:           addr,
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		ErrorLog:       golog.New(stdErr, "", 0),
	}

	go func() {
		defer stdOut.Close()
		defer stdErr.Close()

		sErr := s.ListenAndServe()
		if sErr != nil {
			panic(sErr)
		}
	}()

	return nil
}

func (m *mod) Stop() error {
	return nil
}

func (m *mod) Name() string {
	return m.name
}

func (m *mod) Description() string {
	return m.desc
}

func (m *mod) Address() string {
	return m.addr
}
