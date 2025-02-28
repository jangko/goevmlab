// Copyright 2022 Martin Holst Swende
// This file is part of the go-evmlab library.
//
// The library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the goevmlab library. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/log"
	"github.com/holiman/goevmlab/common"
	"github.com/holiman/goevmlab/fuzzing"
	"github.com/urfave/cli/v2"
)

var (
	targetFlag = &cli.StringSliceFlag{
		Name:  "target",
		Usage: "fuzzing-target",
	}
	app = initApp()
)

func initApp() *cli.App {
	app := cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Authors = []*cli.Author{{Name: "Martin Holst Swende"}}
	app.Usage = "Fuzzer with various targets"
	app.Flags = append(app.Flags, common.VmFlags...)
	app.Flags = append(app.Flags,
		common.SkipTraceFlag,
		common.ThreadFlag,
		common.LocationFlag,
		targetFlag,
	)
	app.Action = startFuzzer
	return app
}

func main() {
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlInfo, log.StreamHandler(os.Stderr, log.TerminalFormat(true))))
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func startFuzzer(c *cli.Context) error {
	fNames := c.StringSlice(targetFlag.Name)
	// At this point, we only do one at a time
	if len(fNames) == 0 {
		fmt.Printf("At least one fuzzer target needed. ")
		fmt.Printf("Available targets: %v\n", fuzzing.FactoryNames())
		return errors.New("missing target")
	}
	var factory common.GeneratorFn
	if len(fNames) == 1 {
		factory = fuzzing.Factory(fNames[0], "London")
		if factory == nil {
			return fmt.Errorf("unknown target %v", fNames[0])
		}
	} else {
		// Need to put together a meta-factory
		var factories []common.GeneratorFn
		for _, fName := range fNames {
			if f := fuzzing.Factory(fName, "London"); f == nil {
				return fmt.Errorf("unknown target %v", fName)
			} else {
				factories = append(factories, f)
			}
			log.Info("Added factory", "name", fName)
		}
		index := 0
		factory = func() *fuzzing.GstMaker {
			fn := factories[index%len(factories)]
			index++
			return fn()
		}
	}
	return common.ExecuteFuzzer(c, factory, "naivefuzz")
}
