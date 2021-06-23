## Atlas chain 

Atlas chain is a truly fast, permissionless, secure and scalable public blockchain platform.

<a href="https://github.com/mapprotocol/atlas/blob/main/COPYING"><img src="https://img.shields.io/badge/license-GPL%20%20Atlas-lightgrey.svg"></a>

## Building the source


Building atlas requires both a Go (version 1.14 or later) and a C compiler.
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
$ atlas --singlenode  console
```

