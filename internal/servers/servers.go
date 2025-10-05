package servers

import (
	"context"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
	"github.com/sxwebdev/sentinel/internal/alertresolver"
	"github.com/sxwebdev/sentinel/internal/apiserver"
	"github.com/sxwebdev/sentinel/internal/apiserver/interceptor"
	"github.com/sxwebdev/sentinel/internal/hub/hubserver"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/services/baseservices"
	"github.com/sxwebdev/sentinel/pkg/locker"
	"github.com/tkcrm/mx/logger"
	"go.akshayshah.org/connectproto"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/protobuf/encoding/protojson"
)

type Servers struct {
	httpServer *http.Server
}

func New(
	gCtx context.Context,
	l logger.Logger,
	addr string,
	bs *baseservices.BaseServices,
	ar *alertresolver.AlertResolver,
	systemInfo *models.SystemInfo,
	availableUpdateData *locker.Locker[models.AvailableUpdate],
) *Servers {
	as := apiserver.New(gCtx, l, bs, systemInfo, availableUpdateData)
	hs := hubserver.New(gCtx, l, bs, ar)

	mux := http.NewServeMux()

	opts := []connect.HandlerOption{
		connectproto.WithJSON(
			protojson.MarshalOptions{EmitUnpopulated: true},
			protojson.UnmarshalOptions{DiscardUnknown: true},
		),
	}

	var rpcBases []string

	// register hubserver handler
	hubServerPath, hubServerhandler := hs.RegisterHandler(opts...)
	rpcBases = append(rpcBases, hubServerPath)
	mux.Handle(
		"/api"+hubServerPath,
		withCORS(
			http.StripPrefix("/api", hubserver.
				NewInterceptor(l, bs).
				ConnectRPCAuthMiddleware().
				Wrap(hubServerhandler),
			),
		),
	)

	apiInterceptor := interceptor.New(l, bs)

	// register all apiserver handlers
	for _, srv := range as.AllServers() {
		path, h := srv.RegisterHandler(opts...)
		rpcBases = append(rpcBases, path)
		mux.Handle(
			"/api"+path,
			withCORS(http.StripPrefix("/api", apiInterceptor.ConnectRPCAuthMiddleware().Wrap(h))),
		)
	}

	// init gRPC reflection
	allServerNames := append(as.AllServerNames(), hs.Name())
	ref := grpcreflect.NewStaticReflector(allServerNames...)

	// gRPC reflection handlers
	v1Path, v1Handler := grpcreflect.NewHandlerV1(ref)
	v1aPath, v1aHandler := grpcreflect.NewHandlerV1Alpha(ref)
	mux.Handle("/api"+v1Path, http.StripPrefix("/api", v1Handler))
	mux.Handle("/api"+v1aPath, http.StripPrefix("/api", v1aHandler))

	// register spa server
	mux.Handle("/", spaFileServer("index.html"))

	final := &pathRewriter{
		next:       mux,
		rpcBases:   rpcBases,
		reflection: []string{v1Path, v1aPath},
	}

	return &Servers{
		httpServer: &http.Server{
			Addr:              addr,
			Handler:           h2c.NewHandler(final, &http2.Server{}),
			ReadHeaderTimeout: time.Second * 10,
		},
	}
}

// Name returns the name of the servers
func (s *Servers) Name() string { return "servers" }

// Start starts the servers
func (s *Servers) Start(ctx context.Context) error {
	errChan := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
	}

	return nil
}

// Stop stops the servers
func (s *Servers) Stop(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}
