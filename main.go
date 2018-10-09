package main

// gitdeploy
// Copyright (C) 2016 Maximilian Pachl

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// ---------------------------------------------------------------------------------------
//  imports
// ---------------------------------------------------------------------------------------

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/kballard/go-shellquote"

	"github.com/faryon93/gitdeploy/config"
	"github.com/faryon93/gitdeploy/git"
)

// ----------------------------------------------------------------------------------
//  constants
// ----------------------------------------------------------------------------------

const (
	ConfigFileName = "Deployfile"
)

// ----------------------------------------------------------------------------------
//  global variables
// ----------------------------------------------------------------------------------

var (
	// command line arguments
	oneshot   bool
	cycleTime int

	// globale states
	processors = sync.WaitGroup{}
)

// ----------------------------------------------------------------------------------
//  application entry
// ----------------------------------------------------------------------------------

func main() {
	// parse command line args
	flag.BoolVar(&oneshot, "one-shot", false, "")
	flag.IntVar(&cycleTime, "cycle-time", 60, "")
	flag.Parse()

	// make sure all command-line args are supplied
	if len(flag.Args()) < 1 {
		fmt.Println("usage: gitdeploy [--one-shot] [--cycle-time] watch-dir")
		os.Exit(-1)
	}

	// check if the git cli client is installed
	if !git.IsInstalled() {
		fmt.Println("Git client or ssh not found. Please install git/ssh with your package manager.")
		os.Exit(-1)
	}

	for {
		// recursively check the watch directory
		// if any Deployfiles exit
		filepath.Walk(flag.Args()[0], func(path string, f os.FileInfo, err error) error {
			if f != nil && !f.IsDir() && strings.HasSuffix(f.Name(), ConfigFileName) {
				// load the Deployfile
				conf, err := config.Load(path)
				if err != nil {
					log.Println("failed to load", path+":", err.Error())
					return err
				}

				// process the Deployfiles in paralell
				if conf.Provider == config.ProviderGit {
					processors.Add(1)
					go process(*conf)
				}
			}
			return nil
		})

		// wait until all Deployment files have been processed
		processors.Wait()

		// we are finished -> exit the application
		if oneshot {
			return
		}

		time.Sleep(time.Duration(cycleTime) * time.Second)
	}
}

// ----------------------------------------------------------------------------------
//  private functions
// ----------------------------------------------------------------------------------

func process(config config.Config) {
	defer processors.Done()

	// some metadata
	dir := filepath.Dir(config.Path)

	// repo already cloned -> just pull it
	if git.IsRepository(dir) {
		// pull incoming changes
		out, err := git.Pull(dir, config.IdentityFile)
		if err != nil {
			log.Printf("failed to pull incoming changes: %s\n%s", err.Error(), out)
			return
		}

		// check if incoming changes were pulled
		if !strings.Contains(out, git.StdOutUpToDate) {
			log.Printf("deployed new HEAD in \"%s\"\n%s", dir, out)

			// a reload command is specified
			if config.Command != "" {
				cmd, err := shellquote.Split(config.Command)
				if err != nil {
					log.Println("failed to prepare command:", err.Error())
					return
				}

				stdout, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
				if err == nil {
					log.Printf("executed command \"%s\"\n%s",
						config.Command, stdout)
				} else {
					log.Println("failed to execute command:", err.Error())
				}
			}
		}

		// clone the repo, because its not yet cloned
	} else {
		log.Println("repository", dir, "not cloned, initializing deployment...")

		// clone the repository
		out, err := git.Clone(dir, config.Url, config.IdentityFile, config.Branch)
		if err != nil {
			log.Printf("failed to clone repository %s: %s\n%s", dir, err.Error(), out)
			return
		}

		log.Printf("successfully deployed git repository to %s\n%s", dir, out)
	}
}
