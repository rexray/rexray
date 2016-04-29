package handlers

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"
)

// instanceIDHandler is a global HTTP filter for grokking the InstanceIDs
// from the headers
type instanceIDHandler struct {
	handler types.APIFunc
}

// NewInstanceIDHandler returns a new global HTTP filter for grokking the
// InstanceIDs from the headers
func NewInstanceIDHandler() types.Middleware {
	return &instanceIDHandler{}
}

func (h *instanceIDHandler) Name() string {
	return "instanceIDs-handler"
}

func (h *instanceIDHandler) Handler(m types.APIFunc) types.APIFunc {
	return (&instanceIDHandler{m}).Handle
}

// Handle is the type's Handler function.
func (h *instanceIDHandler) Handle(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	valMap := map[string]*types.InstanceID{}

	if err := parseInstanceIDHeaders(
		ctx,
		types.InstanceIDHeader,
		utils.GetHeader(req.Header, types.InstanceIDHeader),
		valMap); err != nil {
		return err
	}

	if err := parseInstanceIDHeaders(
		ctx,
		types.InstanceID64Header,
		utils.GetHeader(req.Header, types.InstanceID64Header),
		valMap); err != nil {
		return err
	}

	if len(valMap) > 0 {
		ctx = ctx.WithInstanceIDsByService(valMap)
	}

	return h.handler(ctx, w, req, store)
}

func parseInstanceIDHeaders(
	ctx types.Context,
	name string,
	headers []string,
	instanceIDs map[string]*types.InstanceID) error {

	ctx.WithField(name, headers).Debug("http header")

	for _, v := range headers {
		iidParts := strings.SplitN(v, "=", 2)
		iidDriver := strings.ToLower(iidParts[0])
		iidValue := iidParts[1]

		iid := &types.InstanceID{}

		if name == types.InstanceIDHeader {
			iid.ID = iidValue
		} else {
			buf, err := base64.StdEncoding.DecodeString(iidValue)
			if err != nil {
				return err
			}
			if err := json.Unmarshal(buf, iid); err != nil {
				return err
			}
		}

		instanceIDs[iidDriver] = iid
		ctx.WithFields(log.Fields{
			"driver":     iidDriver,
			"instanceID": iid.ID,
		}).Debug("set instanceID")
	}

	return nil
}
