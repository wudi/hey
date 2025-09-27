package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v3"
	"github.com/wudi/hey/pkg/fpm/master"
	"github.com/wudi/hey/pkg/fpm/pool"
	"github.com/wudi/hey/version"
)

func main() {
	app := &cli.Command{
		Name:    "hey-fpm",
		Usage:   "FastCGI Process Manager for Hey PHP Interpreter",
		Version: version.FullVersion(),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "fpm-config",
				Aliases: []string{"y"},
				Usage:   "Specify alternative path to FastCGI process manager config file",
			},
			&cli.StringFlag{
				Name:  "listen",
				Usage: "Listen address (e.g., 127.0.0.1:9000 or /var/run/hey-fpm.sock)",
				Value: "127.0.0.1:9000",
			},
			&cli.BoolFlag{
				Name:  "nodaemonize",
				Usage: "Run in foreground (do not daemonize)",
				Value: true,
			},
			&cli.StringFlag{
				Name:  "pid",
				Usage: "Path to PID file",
				Value: "/var/run/hey-fpm.pid",
			},
			&cli.StringFlag{
				Name:  "pm",
				Usage: "Process management mode (static, dynamic, ondemand)",
				Value: "dynamic",
			},
			&cli.IntFlag{
				Name:  "pm-max-children",
				Usage: "Maximum number of child processes",
				Value: 50,
			},
			&cli.IntFlag{
				Name:  "pm-start-servers",
				Usage: "Number of child processes to start (dynamic mode)",
				Value: 5,
			},
			&cli.IntFlag{
				Name:  "pm-min-spare-servers",
				Usage: "Minimum number of idle processes (dynamic mode)",
				Value: 5,
			},
			&cli.IntFlag{
				Name:  "pm-max-spare-servers",
				Usage: "Maximum number of idle processes (dynamic mode)",
				Value: 35,
			},
			&cli.IntFlag{
				Name:  "pm-max-requests",
				Usage: "Number of requests each worker handles before respawning",
				Value: 500,
			},
			&cli.BoolFlag{
				Name:    "test",
				Aliases: []string{"t"},
				Usage:   "Test configuration and exit",
			},
		},
		Action: runFPM,
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatalf("Error: %v\n", err)
	}
}

func runFPM(ctx context.Context, cmd *cli.Command) error {
	if cmd.Bool("test") {
		fmt.Println("Configuration test successful")
		return nil
	}

	pm := pool.ProcessManagement(cmd.String("pm"))
	if pm != pool.PMStatic && pm != pool.PMDynamic && pm != pool.PMOndemand {
		return fmt.Errorf("invalid process management mode: %s (must be static, dynamic, or ondemand)", pm)
	}

	poolConfig := &pool.PoolConfig{
		Name:              "www",
		ProcessManagement: pm,
		MaxChildren:       cmd.Int("pm-max-children"),
		StartServers:      cmd.Int("pm-start-servers"),
		MinSpareServers:   cmd.Int("pm-min-spare-servers"),
		MaxSpareServers:   cmd.Int("pm-max-spare-servers"),
		MaxRequests:       cmd.Int("pm-max-requests"),
	}

	masterConfig := &master.MasterConfig{
		Listen:     cmd.String("listen"),
		PIDFile:    cmd.String("pid"),
		ErrorLog:   "/var/log/hey-fpm.log",
		LogLevel:   "notice",
		PoolConfig: poolConfig,
	}

	m := master.NewMaster(masterConfig)

	if err := m.Start(); err != nil {
		return fmt.Errorf("failed to start FPM: %v", err)
	}

	log.Printf("Hey-FPM started successfully")
	log.Printf("Listening on: %s", masterConfig.Listen)
	log.Printf("Process management: %s", poolConfig.ProcessManagement)
	log.Printf("Max children: %d", poolConfig.MaxChildren)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	sig := <-sigChan
	log.Printf("Received signal %v, shutting down gracefully", sig)

	m.GracefulShutdown()
	m.Wait()

	log.Printf("Hey-FPM shutdown complete")
	return nil
}