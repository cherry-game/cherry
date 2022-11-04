package main

import (
	"fmt"
	cherryConst "github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/examples/game_cluster/nodes/center"
	"github.com/cherry-game/cherry/examples/game_cluster/nodes/game"
	"github.com/cherry-game/cherry/examples/game_cluster/nodes/gate"
	"github.com/cherry-game/cherry/examples/game_cluster/nodes/master"
	"github.com/cherry-game/cherry/examples/game_cluster/nodes/web"
	"github.com/urfave/cli/v2"
	"os"
)

func main() {
	app := &cli.App{
		Name:        "game cluster node",
		Description: "game cluster node examples",
		Commands: []*cli.Command{
			versionCommand(),
			gameCommand(),
			gateCommand(),
			webCommand(),
			masterCommand(),
			centerCommand(),
		},
	}

	_ = app.Run(os.Args)
}

func versionCommand() *cli.Command {
	return &cli.Command{
		Name:      "version",
		Aliases:   []string{"ver", "v"},
		Usage:     "view version",
		UsageText: "game cluster node version",
		Action: func(c *cli.Context) error {
			fmt.Println(cherryConst.Version())
			return nil
		},
	}
}

func gameCommand() *cli.Command {
	return &cli.Command{
		Name:      "game",
		Usage:     "run game node",
		UsageText: "./node game --path=./config --name=game-cluster --node=10001",
		Flags:     getFlag(),
		Action: func(c *cli.Context) error {
			path, name, node := getParameters(c)
			game.Run(path, name, node)
			return nil
		},
	}
}

func gateCommand() *cli.Command {
	return &cli.Command{
		Name:      "gate",
		Usage:     "run gate node",
		UsageText: "./node gate --path=./config --name=game-cluster --node=gate-1",
		Flags:     getFlag(),
		Action: func(c *cli.Context) error {
			path, name, node := getParameters(c)
			gate.Run(path, name, node)
			return nil
		},
	}
}

func webCommand() *cli.Command {
	return &cli.Command{
		Name:      "web",
		Usage:     "run web node",
		UsageText: "./node web --path=./config --name=game-cluster --node=web-1",
		Flags:     getFlag(),
		Action: func(c *cli.Context) error {
			path, name, node := getParameters(c)
			web.Run(path, name, node)
			return nil
		},
	}
}

func masterCommand() *cli.Command {
	return &cli.Command{
		Name:      "master",
		Usage:     "run master node",
		UsageText: "./node master --path=./config --name=game-cluster --node=master-1",
		Flags:     getFlag(),
		Action: func(c *cli.Context) error {
			path, name, node := getParameters(c)
			master.Run(path, name, node)
			return nil
		},
	}
}

func centerCommand() *cli.Command {
	return &cli.Command{
		Name:      "center",
		Usage:     "run center node",
		UsageText: "./node center --path=./config --name=game-cluster --node=center-1",
		Flags:     getFlag(),
		Action: func(c *cli.Context) error {
			path, name, node := getParameters(c)
			center.Run(path, name, node)
			return nil
		},
	}
}

func getParameters(c *cli.Context) (path, name, node string) {
	path = c.String("path")
	name = c.String("name")
	node = c.String("node")

	return path, name, node
}

func getFlag() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:     "path",
			Usage:    "profile config path",
			Required: false,
			Value:    "./examples/config",
		},
		&cli.StringFlag{
			Name:     "name",
			Usage:    "profile environment name",
			Required: false,
			Value:    "gc",
		},
		&cli.StringFlag{
			Name:     "node",
			Usage:    "node id name",
			Required: true,
			Value:    "",
		},
	}
}
