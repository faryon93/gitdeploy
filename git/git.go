package git

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
	"os"
	"os/exec"
	"path/filepath"
)

// ----------------------------------------------------------------------------------
//  constants
// ----------------------------------------------------------------------------------

const (
	GitDirectory = ".git"
	GitBinary    = "git"
	GitRemote    = "origin"

	SshBinary = "ssh"

	StdOutUpToDate = "Already up-to-date."
)

// ----------------------------------------------------------------------------------
//  public functions
// ----------------------------------------------------------------------------------

// IsRepository returns true if the given directory is a git repository.
func IsRepository(dir string) bool {
	path := filepath.Join(dir, GitDirectory)
	_, err := os.Stat(path)

	return !os.IsNotExist(err)
}

func IsInstalled() bool {
	git, err := exec.LookPath(GitBinary)
	if err != nil {
		return false
	}

	ssh, err := exec.LookPath(SshBinary)
	if err != nil {
		return false
	}

	return len(git) > 0 && len(ssh) > 0
}

// Clone clones the given repostiory to the filesystem.
func Clone(path string, url string, identity string, branch string) (string, error) {
	// initialize the repository
	out, err := execGit(path, identity, "init")
	if err != nil {
		return out, err
	}

	// add the url as a new remote
	out, err = execGit(path, identity, "remote", "add", GitRemote, url)
	if err != nil {
		return out, err
	}

	// pull the master branch
	out, err = execGit(path, identity, "fetch")
	if err != nil {
		return out, err
	}

	// setup upstream brunch
	return execGit(path, identity, "checkout", "-t", GitRemote+"/"+branch)
}

// Pull pulls new changes from a remote repository.
func Pull(path string, identity string) (string, error) {
	return execGit(path, identity, "pull")
}

// ----------------------------------------------------------------------------------
//  private functions
// ----------------------------------------------------------------------------------

func execGit(path string, identity string, args ...string) (string, error) {
	// find the absolute path to the
	// git cli bianry
	binary, err := exec.LookPath(GitBinary)
	if err != nil {
		return "", err
	}

	// build the git command
	cmd := exec.Command(binary, args...)
	cmd.Dir = path
	cmd.Env = []string{"GIT_SSH_COMMAND=" + SshBinary + " " +
		"-F /dev/null " +
		"-o UserKnownHostsFile=/dev/null " +
		"-o StrictHostKeyChecking=no " +
		"-i " + identity}

	// get the stdout/stderr of the git process
	// CombinedOutput() will block until end
	out, err := cmd.CombinedOutput()
	return string(out), err
}
