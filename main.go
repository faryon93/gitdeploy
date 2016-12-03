package main

import (
    "fmt"
    "log"
    "flag"
    "os"
    "path/filepath"
    "strings"
    "sync"
    "time"

    "git.1vh.de/maximilian.pachl/gitdeploy/config"
    "git.1vh.de/maximilian.pachl/gitdeploy/git"
)


// ----------------------------------------------------------------------------------
//  constants
// ----------------------------------------------------------------------------------

const (
    CONFIG_FILE_NAME = "Deployfile"
)


// ----------------------------------------------------------------------------------
//  global variables
// ----------------------------------------------------------------------------------

// command line arguments
var oneshot bool
var cycleTime int

// globale states
var processors sync.WaitGroup = sync.WaitGroup{}


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
            if f != nil && !f.IsDir() && strings.HasSuffix(f.Name(), CONFIG_FILE_NAME) {
                // load the Deployfile
                conf, err := config.Load(path)
                if err != nil {
                    log.Println("failed to load", path + ":", err.Error())
                    return err
                }

                // process the Deployfiles in paralell
                if conf.Provider == config.PROVIDER_GIT {
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

        // sleep until the next cycle   
        } else {
            time.Sleep(time.Duration(cycleTime) * time.Second)
        }
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
        if !strings.Contains(out, git.UP_TO_DATE) {
            log.Printf("pulled incoming changes: %s", out)
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
