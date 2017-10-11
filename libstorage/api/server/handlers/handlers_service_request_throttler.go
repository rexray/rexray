package handlers

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"time"

	gofigCore "github.com/akutz/gofig"
	gofig "github.com/akutz/gofig/types"

	apictx "github.com/codedellemc/rexray/libstorage/api/context"
	"github.com/codedellemc/rexray/libstorage/api/types"

	// Load the etcd client package.
	etcd "github.com/coreos/etcd/clientv3"
	etcdsync "github.com/coreos/etcd/clientv3/concurrency"
)

func init() {
	r := gofigCore.NewRegistration("etcd")
	r.Key(gofig.String, "", "",
		"A list of etcd endpoints", "etcd.endpoints")
	gofigCore.Register(r)
}

const (
	ttlBurst     = 60    // 1m
	ttlSustained = 86400 // 1d

	maxBurst     = 100
	maxSustained = 1000
)

// svcReqThrottler is an HTTP filter for throttling incoming service
// requests based on allowed burst and sustained request rates.
type svcReqThrottler struct {
	handler types.APIFunc
	config  gofig.Config
}

// NewServiceRequestThrottler returns a new filter for for throttling
// incoming service requests based on allowed burst and sustained request
// rates.
func NewServiceRequestThrottler(config gofig.Config) types.Middleware {
	return &svcReqThrottler{config: config}
}

func (h *svcReqThrottler) Name() string {
	return "service-request-throttler"
}

func (h *svcReqThrottler) Handler(m types.APIFunc) types.APIFunc {
	return (&svcReqThrottler{config: h.config, handler: m}).Handle
}

// Handle is the type's Handler function.
func (h *svcReqThrottler) Handle(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	svc, ok := apictx.Service(ctx)
	if !ok {
		return types.ErrMissingStorageService
	}

	// Scope the config to the request's service.
	config := h.config.Scope(
		fmt.Sprintf("libstorage.server.services.%s", svc.Name()))

	// Get the etcd connection information.
	etcdEndpoints := config.GetStringSlice("etcd.endpoints")

	// If there is no etcd connection info then quit this handler.
	if len(etcdEndpoints) == 0 {
		ctx.Debug("no etcd endpoints detected; disabling throttling")
		return h.handler(ctx, w, req, store)
	}

	// Attempt to connect to etcd.
	etcdClient, err := etcd.New(etcd.Config{
		Endpoints: etcdEndpoints,
	})
	if err != nil {
		ctx.Errorf("etcd: connection failed: %v", err)
		return err
	}
	defer func() {
		if err := etcdClient.Close(); err != nil {
			ctx.Errorf("etcd: close connection failed: %v", err)
		}
	}()
	ctx.WithField("endpoints", etcdEndpoints).Debug("connected to etcd")

	// Get the service's throttling domain.
	tdomain := config.GetString("throttling.domain")
	if tdomain == "" {
		tdomain = config.GetString("rexray.throttling.domain")
		if tdomain == "" {
			tdomain = svc.Name()
		}
	}

	var (
		// Define the throttling keys that store the throttling domain's
		// mutex and burst and sustained limits in etcd.
		tkeyDomain = path.Join("/rexray/throttling", tdomain)
		tkeyMutex  = path.Join(tkeyDomain, "mutex")

		tkeyBurst      = path.Join(tkeyDomain, "burst")
		tkeyBurstLease = path.Join(tkeyBurst, "lease")
		tkeyBurstCnxns = path.Join(tkeyBurst, "cnxns")

		tkeySust      = path.Join(tkeyDomain, "sustained")
		tkeySustLease = path.Join(tkeySust, "lease")
		tkeySustCnxns = path.Join(tkeySust, "cnxns")

		// burstLeaseID and sustLeaseID are the lease IDs for the
		// leases that control the number of allowed connections during
		// a burst and sustained period of time.
		burstLeaseID etcd.LeaseID
		sustLeaseID  etcd.LeaseID
	)

	// Check to see if this request is allowed per the number of
	// active sustained and burst connections.
	chkSustCnxns := func() error {
		sustCnxns, err := etcdClient.Get(
			ctx,
			tkeySustCnxns,
			etcd.WithPrefix(),
			etcd.WithCountOnly())
		if err != nil {
			ctx.Errorf(
				"etcd: get sustained cnxn count failed: %s: %v",
				tkeySustCnxns, err)
			return err
		}
		if sustCnxns.Count > maxSustained {
			return fmt.Errorf(
				"throttled: sustained count exceeded: %d", maxSustained)
		}
		return nil
	}
	if err := chkSustCnxns(); err != nil {
		ctx.Error(err)
		return err
	}

	chkBurstCnxns := func() error {
		burstCnxns, err := etcdClient.Get(
			ctx,
			tkeyBurstCnxns,
			etcd.WithPrefix(),
			etcd.WithCountOnly())
		if err != nil {
			ctx.Errorf(
				"etcd: get burst cnxn count failed: %s: %v",
				tkeyBurstCnxns, err)
			return err
		}
		if burstCnxns.Count > maxBurst {
			return fmt.Errorf(
				"throttled: burst count exceeded: %d", maxBurst)
		}
		return nil
	}
	if err := chkBurstCnxns(); err != nil {
		ctx.Error(err)
		return err
	}

	// Create a cancellation context used to create the etcd
	// concurrency session and obtain the lock.
	lockCtx, cancel := context.WithDeadline(
		ctx, time.Now().Add(time.Second*120))
	defer cancel()

	// Create the etcd key API instance used to access keys.
	etcdSession, err := etcdsync.NewSession(
		etcdClient, etcdsync.WithContext(lockCtx))
	if err != nil {
		ctx.Errorf("etcd: create concurrency session failed: %v", err)
		return err
	}
	defer func() {
		if err := etcdSession.Close(); err != nil {
			ctx.Errorf("etcd: close concurrency session failed: %v", err)
		}
	}()

	// Create a mutex for the throttling domain and attempt to obtain
	// its lock.
	lock := etcdsync.NewMutex(etcdSession, tkeyMutex)
	if err := lock.Lock(lockCtx); err != nil {
		ctx.Errorf("etcd: lock attempt failed: %v", err)
		return err
	}
	defer func() {
		if err := lock.Unlock(ctx); err != nil {
			ctx.Errorf("etcd: unlock attempt failed: %v", err)
		}
	}()

	// Get the ID of the burst lease.
	burstLeaseNew := false
	burstLeaseRes, err := etcdClient.Get(ctx, tkeyBurstLease)
	if err != nil {
		ctx.Errorf("etcd: get burst lease ID failed: %v", err)
		return err
	}

	// If there are no keys then create a lease for burst tracking.
	if burstLeaseRes.Count == 0 {
		grantBurstLeaseResp, err := etcdClient.Grant(ctx, ttlBurst)
		if err != nil {
			ctx.Errorf("etcd: grant burst lease failed: %v", err)
			return err
		}
		burstLeaseID = grantBurstLeaseResp.ID
		burstLeaseNew = true

		// Store the lease ID with a TTL referencing the epoynymous
		// lease. This way the lease ID expires when the lease does,
		// creating a natural workflow for when to create a new lease.
		if _, err := etcdClient.Put(
			ctx, tkeyBurstLease,
			fmt.Sprintf("%d", burstLeaseID),
			etcd.WithLease(burstLeaseID)); err != nil {

			ctx.Errorf("etcd: put burst lease id failed: %v", err)
			return err
		}
	} else {
		szID := string(burstLeaseRes.Kvs[0].Value)
		i, err := strconv.ParseInt(szID, 10, 64)
		if err != nil {
			ctx.Errorf("etcd: invalid burst lease id: %s: %v", szID, err)
			return err
		}
		burstLeaseID = etcd.LeaseID(i)
	}

	// Get the ID of the sustained lease.
	sustLeaseRes, err := etcdClient.Get(ctx, tkeySustLease)
	sustLeaseNew := false
	if err != nil {
		ctx.Errorf("etcd: get sustained lease ID failed: %v", err)
		return err
	}

	// If there are no keys then create a lease for sustained tracking.
	if sustLeaseRes.Count == 0 {
		grantSustLeaseResp, err := etcdClient.Grant(ctx, ttlSustained)
		if err != nil {
			ctx.Errorf("etcd: grant sustained lease failed: %v", err)
			return err
		}
		sustLeaseID = grantSustLeaseResp.ID
		sustLeaseNew = true

		// Store the lease ID with a TTL referencing the epoynymous
		// lease. This way the lease ID expires when the lease does,
		// creating a natural workflow for when to create a new lease.
		if _, err := etcdClient.Put(
			ctx, tkeySustLease,
			fmt.Sprintf("%d", sustLeaseID),
			etcd.WithLease(sustLeaseID)); err != nil {

			ctx.Errorf("etcd: put sustained lease id failed: %v", err)
			return err
		}
	} else {
		szID := string(sustLeaseRes.Kvs[0].Value)
		i, err := strconv.ParseInt(szID, 10, 64)
		if err != nil {
			ctx.Errorf("etcd: invalid sustained lease id: %s: %v", szID, err)
			return err
		}
		sustLeaseID = etcd.LeaseID(i)
	}

	var (
		// txID is this request's transaction ID.
		txID = apictx.MustTransaction(ctx).ID.String()

		// epoch is the current epoch.
		epoch = fmt.Sprintf("%d", time.Now().Unix())

		// Define the keys used to record this request as an active connection.
		tkeyBurstCnxnsTX    = path.Join(tkeyBurstCnxns, txID)
		tkeyBurstCnxnsTXReq = path.Join(tkeyBurstCnxnsTX, epoch)
		tkeySustCnxnsTX     = path.Join(tkeySustCnxns, txID)
		tkeySustCnxnsTXReq  = path.Join(tkeySustCnxnsTX, epoch)
	)

	// Check the sustained and burst connection count again.
	if err := chkSustCnxns(); err != nil {
		ctx.Error(err)
		return err
	}
	if err := chkBurstCnxns(); err != nil {
		ctx.Error(err)
		return err
	}

	// Refresh the leases if they're new.
	if sustLeaseNew {
		if _, err := etcdClient.KeepAliveOnce(ctx, sustLeaseID); err != nil {
			ctx.Errorf("etcd: keep sustained lease alive once failed: %d: %v",
				sustLeaseID, err)
			return err
		}
	}
	if burstLeaseNew {
		if _, err := etcdClient.KeepAliveOnce(ctx, burstLeaseID); err != nil {
			ctx.Errorf("etcd: keep burst lease alive once failed: %d: %v",
				burstLeaseID, err)
			return err
		}
	}

	// Record the connection as both a sustained and burst connection.
	// Store the lease ID with a TTL referencing the epoynymous
	// lease. This way the lease ID expires when the lease does,
	// creating a natural workflow for when to create a new lease.
	if _, err := etcdClient.Put(
		ctx, tkeySustCnxnsTXReq, "",
		etcd.WithLease(sustLeaseID)); err != nil {

		ctx.Errorf("etcd: record sustained connection failed: %s: %v",
			tkeySustCnxnsTXReq, err)
		return err
	}
	if _, err := etcdClient.Put(
		ctx, tkeyBurstCnxnsTXReq, "",
		etcd.WithLease(burstLeaseID)); err != nil {

		ctx.Errorf("etcd: record burst connection failed: %s: %v",
			tkeyBurstCnxnsTXReq, err)
		return err
	}

	return h.handler(ctx, w, req, store)
}
