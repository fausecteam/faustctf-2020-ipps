# ![Logo](./web/static/img/logo_small.png "IPPS") Interplanetary Parcel Service (IPPS)

This is the repository of IPPS's web services.

## Building
`go build ./cmd/ipps`

## Running
1. Copy the default configuration file `configs/defaults.toml` to `./config.toml`
2. Apply changes to the configuration file as necessary for the server infrastructure.
3. Copy to the systemd configuration `init/systemd` to  `/etc/systemd/`
4. Start the systemd service
