package main

import (
	"errors"
	"os"
	"strconv"

	"github.com/abiosoft/ishell"
	log "github.com/sirupsen/logrus"

	"aos_updatetester/grpcserver"
)

/*******************************************************************************
 * Types
 ******************************************************************************/

type updateTester struct {
	server  *grpcserver.Instance
	shell   *ishell.Shell
	clients []string
}

/*******************************************************************************
 * Variables
 ******************************************************************************/

var errWrongArgCount = errors.New("wrong argument count")

/*******************************************************************************
 * Init
 ******************************************************************************/

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: false,
		TimestampFormat:  "2006-01-02 15:04:05.000",
		FullTimestamp:    true})
	log.SetLevel(log.DebugLevel)
	log.SetOutput(os.Stdout)
}

/*******************************************************************************
 * Main
 ******************************************************************************/

func main() {
	url := ":9000"

	if len(os.Args) > 1 {
		url = os.Args[1]
	}

	tester, err := newUpdateTester(url)
	if err != nil {
		log.Fatalf("Can't create update tester: %s")
	}
	defer tester.close()

	tester.run()
}

/*******************************************************************************
 * Private
 ******************************************************************************/

func newUpdateTester(url string) (tester *updateTester, err error) {
	tester = &updateTester{}

	defer func(tester *updateTester) {
		if err != nil {
			tester.close()
		}
	}(tester)

	if tester.server, err = grpcserver.New(url, tester); err != nil {
		return nil, err
	}

	tester.shell = ishell.New()

	tester.shell.AddCmd(&ishell.Cmd{
		Name: "prepare",
		Help: "prepare <id> <path> <version>",
		Completer: func(args []string) []string {
			if len(args) == 0 {
				return tester.clients
			}

			return []string{}
		},
		Func: func(context *ishell.Context) {
			if len(context.Args) != 3 {
				context.Err(errWrongArgCount)
				return
			}

			version, err := strconv.ParseUint(context.Args[2], 10, 64)
			if err != nil {
				context.Err(err)
				return
			}

			if err = tester.server.PrepareUpdate(context.Args[0], context.Args[1], version); err != nil {
				context.Err(err)
				return
			}
		},
	})

	tester.shell.AddCmd(&ishell.Cmd{
		Name: "update",
		Help: "update <id>",
		Completer: func(args []string) []string {
			if len(args) == 0 {
				return tester.clients
			}

			return []string{}
		},
		Func: func(context *ishell.Context) {
			if len(context.Args) != 1 {
				context.Err(errWrongArgCount)
				return
			}

			if err = tester.server.StartUpdate(context.Args[0]); err != nil {
				context.Err(err)
				return
			}
		},
	})

	tester.shell.AddCmd(&ishell.Cmd{
		Name: "apply",
		Help: "apply <id>",
		Completer: func(args []string) []string {
			if len(args) == 0 {
				return tester.clients
			}

			return []string{}
		},
		Func: func(context *ishell.Context) {
			if len(context.Args) != 1 {
				context.Err(errWrongArgCount)
				return
			}

			if err = tester.server.ApplyUpdate(context.Args[0]); err != nil {
				context.Err(err)
				return
			}
		},
	})

	tester.shell.AddCmd(&ishell.Cmd{
		Name: "revert",
		Help: "revert <id>",
		Completer: func(args []string) []string {
			if len(args) == 0 {
				return tester.clients
			}

			return []string{}
		},
		Func: func(context *ishell.Context) {
			if len(context.Args) != 1 {
				context.Err(errWrongArgCount)
				return
			}

			if err = tester.server.RevertUpdate(context.Args[0]); err != nil {
				context.Err(err)
				return
			}
		},
	})

	tester.shell.Printf("Start server, url: %s\n", url)

	return tester, nil
}

func (tester *updateTester) close() {
	if tester.server != nil {
		tester.server.Close()
	}
}

func (tester *updateTester) run() {
	tester.shell.Run()
}

func (tester *updateTester) Registered() {
	tester.shell.Println("Client registered")
}

func (tester *updateTester) Disconnected(id string, err error) {
	tester.shell.Printf("Client %s disconnected\n", id)

	tester.removeClient(id)
}

func (tester *updateTester) Status(id, state, err string) {
	tester.shell.Printf("Status received, id: %s, state: %s, err: %s\n", id, state, err)

	tester.addClient(id)
}

func (tester *updateTester) addClient(id string) {
	for _, client := range tester.clients {
		if client == id {
			return
		}
	}

	tester.clients = append(tester.clients, id)
}

func (tester *updateTester) removeClient(id string) {
	for i, client := range tester.clients {
		if client == id {
			tester.clients = append(tester.clients[:i], tester.clients[i+1:]...)
			return
		}
	}
}
