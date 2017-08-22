package agent

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/jonbonazza/huton/command"
	"github.com/jonbonazza/huton/lib"
	"github.com/mitchellh/cli"
)

// Command is a CLI command use to start a huton agent.
type Command struct {
	UI       cli.Ui
	instance huton.Instance
}

// Run is called by the CLI to execute the command.
func (c *Command) Run(args []string) int {
	name, opts, peers, err := c.readConfig()
	if err != nil {
		return 1
	}
	c.instance, err = huton.NewInstance(name, opts...)
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}
	_, err = c.instance.Join(peers)
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}
	return c.handleSignals()
}

// Synopsis is used by the CLI to provide a synopsis of the command.
func (c *Command) Synopsis() string {
	return ""
}

// Help is used by the CLI to provide help text for the command.
func (c *Command) Help() string {
	return ""
}

func (c *Command) readConfig() (string, []huton.Option, []string, error) {
	var name string
	var bindAddr string
	var bindPort int
	var bootstrap bool
	var bootstrapExpect int
	peers := []string{}
	flags := flag.NewFlagSet("agent", flag.ContinueOnError)
	flags.Usage = func() {
		c.UI.Output(c.Help())
	}
	flags.StringVar(&name, "name", "", "unique instance name")
	flags.StringVar(&bindAddr, "bindAddr", "", "address to bind serf to")
	flags.IntVar(&bindPort, "bindPort", -1, "port to bind serf to")
	flags.BoolVar(&bootstrap, "bootstrap", false, "bootstrap mode")
	flags.IntVar(&bootstrapExpect, "expect", -1, "bootstrap expect")
	flags.Var((*command.AppendSliceValue)(&peers), "peers", "peer list")
	if err := flags.Parse(os.Args[2:]); err != nil {
		return "", nil, nil, err
	}
	var opts []huton.Option
	opts = append(opts, huton.Bootstrap(bootstrap))
	if bindAddr != "" {
		opts = append(opts, huton.BindAddr(bindAddr))
	}
	if bindPort >= 0 {
		opts = append(opts, huton.BindPort(bindPort))
	}
	if bootstrapExpect >= 0 {
		opts = append(opts, huton.BootstrapExpect(bootstrapExpect))
	}
	return name, opts, peers, nil
}

func (c *Command) handleSignals() int {
	signalCh := make(chan os.Signal, 3)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	select {
	case <-signalCh:
		c.instance.Leave()
		c.instance.Shutdown()
		return 0
	}
}
