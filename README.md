# gitdeploy

Cyclically pulls changes from a remote git repostitory. The cycle time can be configured with in seconds the `--cycle-time` option.

## Example Deployfile

    provider = "git"
    url = "git@github.com:faryon93/test.git"
    identity_file = "~/id_rsa"
    branch = "master"

