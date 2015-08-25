package admin

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/emccode/rexray/daemon/module"
	"github.com/emccode/rexray/util"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const MOD_PORT = 7979
const MOD_NAME = "AdminModule"
const MOD_DESC = "The REX-Ray admin module"

type Module struct {
	id   int32
	name string
	addr string
	desc string
}

type JsonError struct {
	msg string `json:"message"`
	err error  `json:"error"`
}

func init() {
	addr := fmt.Sprintf("tcp://:%d", MOD_PORT)
	module.RegisterModule(MOD_NAME, false, Init, []string{addr})
}

func Init(id int32, address string) (module.Module, error) {
	return &Module{
		id:   id,
		name: MOD_NAME,
		desc: MOD_DESC,
		addr: address,
	}, nil
}

func (mod *Module) Id() int32 {
	return mod.id
}

func loadAsset(path, defaultValue string) string {

	devPath := fmt.Sprintf(
		"%s/src/github.com/emccode/rexray/daemon/module/admin/html/%s",
		os.Getenv("GOPATH"),
		path)

	if util.FileExists(devPath) {
		v, _ := ioutil.ReadFile(devPath)
		log.Printf("Loaded %s from %s\n", path, devPath)
		return string(v)
	}

	exeDir, _, _ := util.GetThisPathParts()

	relPath := fmt.Sprintf(
		"%s/html/%s",
		exeDir,
		path)

	if util.FileExists(relPath) {
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
	fmt.Fprint(w, loadAsset("index.html", HtmlIndex))
}

func scriptsHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/javascript; charset=UTF-8")
	a := loadAsset("scripts/jquery-1.11.3.min.js", ScriptJQuery)
	writeContentLength(w, a)
	fmt.Fprint(w, a)
}

func stylesHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=UTF-8")
	a := loadAsset("styles/main.css", StyleMain)
	writeContentLength(w, a)
	fmt.Fprint(w, a)
}

func imagesHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml; charset=UTF-8")
	a := loadAsset("images/rexray-banner-logo.svg", ImageRexRayBannerLogo)
	writeContentLength(w, a)
	fmt.Fprint(w, a)
}

func moduleTypeHandler(w http.ResponseWriter, req *http.Request) {
	mods := make([]*module.ModuleType, 0)
	for m := range module.ModuleTypes() {
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
	mods := make([]*module.ModuleInstance, 0)
	for m := range module.ModuleInstances() {
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
	typeId := req.FormValue("typeId")
	address := req.FormValue("address")
	start := req.FormValue("start")

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if typeId == "" || address == "" {
		w.Write(jsonError("Fields typeId and address are required", nil))
		log.Printf("Fields typeId and address are required\n")
		return
	}

	typeIdInt, typeIdIntErr := strconv.ParseInt(typeId, 10, 32)
	if typeIdIntErr != nil {
		w.Write(jsonError("Error parsing typeId", typeIdIntErr))
		log.Printf("Error parsing typeId ERR: %v\n", typeIdIntErr)
		return
	}

	typeIdInt32 := int32(typeIdInt)

	modInst, initErr := module.InitializeModule(typeIdInt32, address)
	if initErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Error initializing module ERR: %v\n", initErr)
		return
	}

	jsonBuf, jsonBufErr := json.MarshalIndent(modInst, "", "  ")
	if jsonBufErr != nil {
		w.Write(jsonError("Error marshalling object to json", jsonBufErr))
		log.Printf("Error marshalling object to json ERR: %v\n", jsonBufErr)
		return
	}

	startBool, startBoolErr := strconv.ParseBool(start)
	if startBoolErr != nil {
		startBool = false
	}

	if startBool {
		startErr := module.StartModule(modInst.Id)
		if startErr != nil {
			w.Write(jsonError("Error starting module", startErr))
			log.Printf("Error starting module ERR: %v\n", startErr)
			return
		} else {

			jsonBufErr = nil
			jsonBuf, jsonBufErr = json.MarshalIndent(modInst, "", "  ")
			if jsonBufErr != nil {
				w.Write(jsonError("Error marshalling object to json", jsonBufErr))
				log.Printf("Error marshalling object to json ERR: %v\n", jsonBufErr)
				return
			}
			w.Write(jsonBuf)
		}
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
	id := vars["id"]
	if id == "" {
		w.Write(jsonError("The URL should include the module instance ID", nil))
		log.Printf("The URL should include the module instance ID\n")
		return
	}

	idInt, idIntErr := strconv.ParseInt(id, 10, 32)

	if idIntErr != nil {
		w.Write(jsonError("Error parsing id", idIntErr))
		log.Printf("Error parsing id ERR: %v\n", idIntErr)
		return
	}

	idInt32 := int32(idInt)

	modInst, modInstErr := module.GetModuleInstance(idInt32)
	if modInstErr != nil {
		w.Write(jsonError("Unknown module id", modInstErr))
		log.Printf("Unknown module id ERR: %v\n", modInstErr)
		return
	}

	jsonBuf, jsonBufErr := json.MarshalIndent(modInst, "", "  ")
	if jsonBufErr != nil {
		w.Write(jsonError("Error marshalling object to json", jsonBufErr))
		log.Printf("Error marshalling object to json ERR: %v\n", jsonBufErr)
		return
	}

	if modInst.IsStarted {
		w.Write(jsonBuf)
		return
	}

	startErr := module.StartModule(idInt32)

	if startErr != nil {
		w.Write(jsonError("Error starting moudle", startErr))
		log.Printf("Error starting module ERR: %v\n", startErr)
		return
	}

	jsonBufErr = nil
	jsonBuf, jsonBufErr = json.MarshalIndent(modInst, "", "  ")
	if jsonBufErr != nil {
		w.Write(jsonError("Error marshalling object to json", jsonBufErr))
		log.Printf("Error marshalling object to json ERR: %v\n", jsonBufErr)
		return
	}

	w.Write(jsonBuf)
}

func jsonError(msg string, err error) []byte {
	buf, marshalErr := json.MarshalIndent(
		&JsonError{
			msg: msg,
			err: err,
		}, "", "  ")
	if marshalErr != nil {
		panic(marshalErr)
	}
	return buf
}

func (mod *Module) Start() error {
	r := mux.NewRouter()

	r.Handle("/r/module/instances",
		handlers.LoggingHandler(os.Stdout, http.HandlerFunc(moduleInstHandler)))
	r.Handle("/r/module/instances/{id}/start",
		handlers.LoggingHandler(os.Stdout, http.HandlerFunc(moduleInstStartHandler)))
	r.Handle("/r/module/types",
		handlers.LoggingHandler(os.Stdout, http.HandlerFunc(moduleTypeHandler)))

	r.Handle("/images/rexray-banner-logo.svg",
		handlers.LoggingHandler(os.Stdout, http.HandlerFunc(imagesHandler)))
	r.Handle("/scripts/jquery-1.11.3.min.js",
		handlers.LoggingHandler(os.Stdout, http.HandlerFunc(scriptsHandler)))
	r.Handle("/styles/main.css",
		handlers.LoggingHandler(os.Stdout, http.HandlerFunc(stylesHandler)))

	r.Handle("/",
		handlers.LoggingHandler(os.Stdout, http.HandlerFunc(indexHandler)))

	_, addr, parseAddrErr := util.ParseAddress(mod.Address())
	if parseAddrErr != nil {
		return parseAddrErr
	}

	s := &http.Server{
		Addr:           addr,
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		sErr := s.ListenAndServe()
		if sErr != nil {
			panic(sErr)
		}
	}()

	return nil
}

func (mod *Module) Stop() error {
	return nil
}

func (mod *Module) Name() string {
	return mod.name
}

func (mod *Module) Description() string {
	return mod.desc
}

func (mod *Module) Address() string {
	return mod.addr
}
