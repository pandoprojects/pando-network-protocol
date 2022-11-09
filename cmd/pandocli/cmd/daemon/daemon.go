package daemon

import (
	"context"
	"log"
	"sync"

	"github.com/spf13/cobra"
	"github.com/pandoprojects/pando/cmd/pandocli/rpc"
)

// startDaemonCmd runs the pandocli daemon
// Example:
//		pandocli daemon start --port=16889
var startDaemonCmd = &cobra.Command{
	Use:     "start",
	Short:   "Run the thatacli daemon",
	Long:    `Run the thatacli daemon.`,
	Example: `pandocli daemon start --port=16889`,
	Run: func(cmd *cobra.Command, args []string) {
		cfgPath := cmd.Flag("config").Value.String()
		server, err := rpc.NewPandoCliRPCServer(cfgPath, portFlag)
		if err != nil {
			log.Fatalf("Failed to run the PandoCli Daemon: %v", err)
		}
		daemon := &PandoCliDaemon{
			RPC: server,
		}
		daemon.Start(context.Background())
		daemon.Wait()
	},
}

func init() {
	startDaemonCmd.Flags().StringVar(&portFlag, "port", "16889", "Port to run the PandoCli Daemon")
}

type PandoCliDaemon struct {
	RPC *rpc.PandoCliRPCServer

	// Life cycle
	wg      *sync.WaitGroup
	quit    chan struct{}
	ctx     context.Context
	cancel  context.CancelFunc
	stopped bool
}

func (d *PandoCliDaemon) Start(ctx context.Context) {
	c, cancel := context.WithCancel(ctx)
	d.ctx = c
	d.cancel = cancel

	if d.RPC != nil {
		d.RPC.Start(d.ctx)
	}
}

func (d *PandoCliDaemon) Stop() {
	d.cancel()
}

func (d *PandoCliDaemon) Wait() {
	if d.RPC != nil {
		d.RPC.Wait()
	}
}
