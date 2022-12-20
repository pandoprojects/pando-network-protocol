package rpc

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"

	"net/rpc"

	"github.com/gorilla/mux"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/pandoprojects/pando/common"
	"github.com/pandoprojects/pando/common/util"
	"github.com/pandoprojects/pando/rpc/lib/rpc-codec/jsonrpc2"
	wl "github.com/pandoprojects/pando/wallet"
	wt "github.com/pandoprojects/pando/wallet/types"
	"golang.org/x/net/netutil"
	"golang.org/x/net/websocket"
)

var logger *log.Entry

type PandoCliRPCService struct {
	wallet wt.Wallet

	// Life cycle
	wg      *sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
	stopped bool
}

// PandoCliRPCServer is an instance of the CLI RPC service.
type PandoCliRPCServer struct {
	*PandoCliRPCService
	port string

	server   *http.Server
	handler  *rpc.Server
	router   *mux.Router
	listener net.Listener
}

// NewPandoCliRPCServer creates a new instance of PandoRPCServer.
func NewPandoCliRPCServer(cfgPath, port string) (*PandoCliRPCServer, error) {
	wallet, err := wl.OpenWallet(cfgPath, wt.WalletTypeSoft, true)
	if err != nil {
		fmt.Printf("Failed to open wallet: %v\n", err)
		return nil, err
	}

	t := &PandoCliRPCServer{
		PandoCliRPCService: &PandoCliRPCService{
			wallet: wallet,
			wg:     &sync.WaitGroup{},
		},
		port: port,
	}

	s := rpc.NewServer()
	s.RegisterName("pandocli", t.PandoCliRPCService)

	t.handler = s

	t.router = mux.NewRouter()
	t.router.Handle("/rpc", jsonrpc2.HTTPHandler(s))
	t.router.Handle("/ws", websocket.Handler(func(ws *websocket.Conn) {
		s.ServeCodec(jsonrpc2.NewServerCodec(ws, s))
	}))

	t.server = &http.Server{
		Handler: t.router,
	}

	logger = util.GetLoggerForModule("rpc")

	return t, nil
}

// Start creates the main goroutine.
func (t *PandoCliRPCServer) Start(ctx context.Context) {
	c, cancel := context.WithCancel(ctx)
	t.ctx = c
	t.cancel = cancel

	t.wg.Add(1)
	go t.mainLoop()
}

func (t *PandoCliRPCServer) mainLoop() {
	defer t.wg.Done()

	go t.serve()

	<-t.ctx.Done()
	t.stopped = true
	t.server.Shutdown(t.ctx)
}

func (t *PandoCliRPCServer) serve() {
	l, err := net.Listen("tcp", ":"+t.port)
	if err != nil {
		logger.WithFields(log.Fields{"error": err}).Fatal("Failed to create listener")
	} else {
		logger.WithFields(log.Fields{"port": t.port}).Info("RPC server started")
	}
	defer l.Close()

	ll := netutil.LimitListener(l, viper.GetInt(common.CfgRPCMaxConnections))
	t.listener = ll

	logger.Fatal(t.server.Serve(ll))
}

// Stop notifies all goroutines to stop without blocking.
func (t *PandoCliRPCServer) Stop() {
	t.cancel()
}

// Wait blocks until all goroutines stop.
func (t *PandoCliRPCServer) Wait() {
	t.wg.Wait()
}
