# gossh

A command line ssh connection and authentication manager. I wanted to learn go so chose a problem to solve. Inspired by nccm (NCurses Connection Manager).

You first need a yaml file with the connection settings for your connections. Possible locations include:
*  `./gossh.yml`
* All yaml files at the path to which the `GOSSH_CONFIGDIR` environment variable points.
* The below files in the user home directory:
  * `.config/gossh/gossh.yml`
  * `gossh.yml`
  * `.gossh.yml`

The structure of the file is as follows (see `example_gossh.yml`) where each ssh server is a separate connection name:
```yaml
connection name:
  key: value
  ...
connection name:
  key: value
  ...
```

The following keys are supported for a given connection:
* `address`: The network address of the ssh connection. If not set, uses the connection name as the address.
* `user`: The user to connect as. If not set, leaves blank which defaults to current user.
* `comment`: Free text field to help indicate the connection. Helpful for filtering.
* `passfile`: Full path to the `age` encrypted file that contains the password for the ssh connection.
* `identity`: Full path to the private key file to use for the ssh connection.

> Note: Gossh checks for and uses a passfile parameter first, then an identity file. If you have both parameters, the passfile will be used (assuming sshpass is installed and in the PATH).

Several environment variables are also supported:
* `GOSSH_TMUX`: (string) When not empty will attempt to set the tmux window name
* `GOSSH_PASSPHRASE`: (string) Uses contents as passphrase to decrypt `age` encrypted password file indicated by `passfile` key on connection
* `GOSSH_LOG_ROLLOVER`: (integer) Sets the maximum size in bytes for the log file before rollover. Defaults to 1048576 (1MB) if not set.

## Roadmap
- [x] Encrypted password files using `age`
- [x] Improve and unify logging
- [x] Increase test coverage
- [ ] Refactor to simplify command execution
- [ ] SSH keys from OCI Vault?

## Logging

The application uses Bubbletea's logging mechanism. Logs are written to `~/.gossh.log` in the user's home directory. For debug-level logging, set the `GOSSH_DEBUG` environment variable to a non-empty value. If the log file cannot be opened, logging falls back to stderr.

The log file will rollover to `~/.gossh.log.old` once it reaches the size specified by `GOSSH_LOG_ROLLOVER` (default 1MB).


## Testing

You can run the tests with `go test ./...` (optionally add `-v`)
