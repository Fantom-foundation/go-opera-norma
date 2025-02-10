# We've Moved!

This repository has found a new home! 

You can now find everything at [https://github.com/0xsoniclabs/sonic](https://github.com/0xsoniclabs/sonic). 

Please update your bookmarks and any links you have to this repo.

# Sonic 

EVM-compatible chain secured by the Lachesis consensus algorithm.

## Building the source

Building Sonic requires both a Go (version 1.21 or later) and a C compiler. You can install
them using your favourite package manager. Once the dependencies are installed, run:

```shell
make all
```
The build output are ```build/sonicd``` and ```build/sonictool``` executables.

## Initialization of the Sonic Database

You will need a genesis file to join a network. Please check the following
site for details how to get one: https://github.com/Fantom-foundation/lachesis_launch
Once you obtain the most recent genesis file available, you need to use the `sonictool`
create a starting DB.

```shell
sonictool --datadir=<target DB path> genesis <path to the genesis file>
```

## Running `sonicd`

Going through all the possible command line flags is out of scope here,
but we've enumerated a few common parameter combos to get you up to speed quickly
on how you can run your own `sonicd` instance.

### Launching a network

Launching `sonicd` readonly (non-validator) node for network specified by the genesis file:

```shell
sonicd --datadir=<DB path>
```

### Configuration

As an alternative to passing the numerous flags to the `sonicd` binary, you can also pass a
configuration file via:

```shell
sonicd --datadir=<DB path> --config /path/to/your/config.toml
```

To get an idea how the file should look like you can use the `dumpconfig` subcommand to
export the default configuration:

```shell
sonictool --datadir=<DB path> dumpconfig
```

### Validator

New validator private key may be created with `sonictool --datadir=<DB path> validator new` command.

To launch a validator, you have to use `--validator.id` and `--validator.pubkey` flags to enable 
events emitter. Check the [Fantom Documentation](https://docs.fantom.foundation) for the detailed process 
of obtaining the validator ID and registering your initial stake.

```shell
sonicd --datadir=<DB path> --validator.id=YOUR_ID --validator.pubkey=0xYOUR_PUBKEY
```

`sonicd` will prompt you for a password to decrypt your validator private key. Optionally, you can
specify password with a file using `--validator.password` flag.

#### Participation in discovery

Optionally you can specify your public IP to straighten connectivity of the network.
Ensure your TCP/UDP p2p port (5050 by default) isn't blocked by your firewall.

```shell
sonicd --datadir=<DB path> --nat=extip:1.2.3.4
```

