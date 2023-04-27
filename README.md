## Atlas chain 

Atlas chain is a truly fast, permissionless, secure and scalable public blockchain platform. [more](https://docs.maplabs.io/)

<a href="https://github.com/mapprotocol/atlas/blob/main/COPYING"><img src="https://img.shields.io/badge/license-GPL%20%20Atlas-lightgrey.svg"></a>

## Building the source


Building atlas requires both a Go (version 1.16 or later) and a C compiler.
You can install them using your favourite package manager.
Once the dependencies are installed, run

```
    make atlas
```

## Running atlas

Going `atlas -h` can get help infos.

### Running on the atlas main network

```
$ atlas console
```


### Running on the Atlas Chain singlenode(private) network

To start a g
instance for single node,  run it with these flags:

```
$ atlas --single  console
```

### Docker quick start
One of the quickest ways to get Atlas up and running on your machine is by using Docker:
```
$ docker run --name atlas --rm -id mapprotocol/atlas:latest
```

## Executables

The go-atlas project comes with several wrappers/executables found in the `cmd`
directory.

|    Command    | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| :-----------: | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
|   `relayer_mock`    | Test cases of relayer module and headerstore module from the perspective of customers.You can get relayer information and storeHeaderMod through this tool    |
|  `relayer_cli`   | Test cases of relayer module  from the perspective of customers. You can get relayer information through this tool                 



### Configuration

As an alternative to passing the numerous flags to the `atlas` binary, you can also pass a configuration file via:

```shell
$ atlas --config /path/to/your_config.toml
```

To get an idea how the file should look like you can use the `dumpconfig` subcommand to
export your existing configuration:

```shell
$ atlas --your-favourite-flags dumpconfig
```


Do not forget `--http.addr 0.0.0.0`, if you want to access RPC from other containers and/or hosts. By default, `atlas` binds to the local interface and RPC endpoints is not accessible from the outside.

### DefultParams
| param    | MainnetChain| TestnetChain      | DevnetChain       | SingleNetChain   |
| :-----------: | :--------------| :---------------- | :----------------| :---------------|
| ChainId       |     22776      |     212      |     213      |     214    |
| NetworkId     |     22776      |     212      |     213      |     214      |
| Port          |     20101      |     20101      |     20101      |20101|
| RpcPort       |     7445      |     7445      |     7445      |     7445   |

| param      | value| comment
| :-----------:  | :--------------  | :--------------   | 
| miner.threads  |     0            |                   |   
| miner.gaslimit |     8000000      |                   |   
| miner.gasprice |    1e9Wei        |                   | 
| TxGas       |    21000         |Minimum gas of creating a transaction |

### Programmatically interfacing `atlas` nodes

As a developer, sooner rather than later you'll want to start interacting with `atlas` and the Atlas network via your own programs and not manually through the console. To aid this, `atlas` has built-in support for a JSON-RPC based APIs standard APIs.
These can be exposed via HTTP, WebSockets and IPC (UNIX sockets on UNIX based
platforms, and named pipes on Windows).

The IPC interface is enabled by default and exposes all the APIs supported by `atlas`,
whereas the HTTP and WS interfaces need to manually be enabled and only expose a subset of APIs due to security reasons. These can be turned on/off and configured as you'd expect.

HTTP based JSON-RPC API options:
  * `--http` Enable the HTTP-RPC server
  * `--http.addr` HTTP-RPC server listening interface (default: `localhost`)
  * `--http.port` HTTP-RPC server listening port (default: `7445`)
  * `--http.api` API's offered over the HTTP-RPC interface (default: `eth,net,web3`)
  * `--http.corsdomain` Comma separated list of domains from which to accept cross origin requests (browser enforced)
  * `--ws` Enable the WS-RPC server
  * `--ws.addr` WS-RPC server listening interface (default: `localhost`)
  * `--ws.port` WS-RPC server listening port (default: `8546`)
  * `--ws.api` API's offered over the WS-RPC interface (default: `eth,net,web3`)
  * `--ws.origins` Origins from which to accept websockets requests
  * `--ipcdisable` Disable the IPC-RPC server
  * `--ipcapi` API's offered over the IPC-RPC interface (default: `admin,debug,eth,miner,net,personal,shh,txpool,web3`)
  * `--ipcpath` Filename for IPC socket/pipe within the datadir (explicit paths escape it)

You'll need to use your own programming environments' capabilities (libraries, tools, etc) to
connect via HTTP, WS or IPC to a `atlas` node configured with the above flags and you'll
need to speak JSON-RPC on all transports. You can reuse the same connection for multiple requests!

**Note: Please understand the security implications of opening up an HTTP/WS based transport before doing so! Hackers on the internet are actively trying to subvert Atlas nodes with exposed APIs! Further, all browser tabs can access locally running web servers, so malicious web pages could try to subvert locally available APIs!**

## Contribution

Thank you for considering to help out with the source code! We welcome contributions
from anyone on the internet, and are grateful for even the smallest of fixes!

If you'd like to contribute to go-atlas, please fork, fix, commit and send a pull request
for the maintainers to review and merge into the main code base. If you wish to submit more complex changes though, please check up with the core devs first on our Discord Server to ensure those changes are in line with the general philosophy of the project and/or get
some early feedback which can make both your efforts much lighter as well as our review and merge procedures quick and simple.

Please make sure your contributions adhere to our coding guidelines:

 * Code must adhere to the official Go [formatting](https://golang.org/doc/effective_go.html#formatting)
   guidelines (i.e. uses [gofmt](https://golang.org/cmd/gofmt/)).
 * Code must be documented adhering to the official Go [commentary](https://golang.org/doc/effective_go.html#commentary)
   guidelines.
 * Pull requests need to be based on and opened against the `master` branch.
 * Commit messages should be prefixed with the package(s) they modify.
   * E.g. "eth, rpc: make trace configs optional"





