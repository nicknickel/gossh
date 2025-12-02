# gossh

A command line ssh connection and authentication manager. I wanted to learn go so chose a problem to solve. Inspired by nccm (NCurses Connection Manager).

You first need a yaml file with the connection settings for your connections. To ease transition from nccm, the tool looks for it's configuration files. Possible locations include:
*  `./gossh.yml`
* `./nccm.yml`
* All yaml files in `/etc/nccm.d/`
* The below files in the user home directory:
  * `.config/nccm/nccm.yml`
  * `nccm.yml`
  * `.nccm.yml`

The structure of the file is as follows where each ssh server is a separate connection name:
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


Several environment variables are also supported:
* `GOSSH_TMUX`: (string) When not empty will attempt to set the tmux window name
* `GOSSH_PASSPHRASE`: (string) Uses contents as passphrase to decrypt `age` encrypted password file indicated by `passfile` key on connection
* `GOSSH_PASSPHRASE_FILE`: (path to file) Uses contents of file as passphrase to decrypt `age` encrypted password file indicated by `passfile` key on connection

## Roadmap
- [ ] Multiple tmux panes for multiple connections
- [x] Encrypted password files using an ssh key
- [ ] Improve and unify logging
- [ ] Increase test coverage
