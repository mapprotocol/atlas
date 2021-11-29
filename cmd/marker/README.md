
# Marker

`marker` is a developer utility to easy running atlas blockchain testnets and related jobs around testnets.

Its main advantage over previous solutions is that it's able to create a `genesis.json` where all core conctracts are already deployed in it.

### Building marker

```bash
go build -o marker *.go
```

## Using marker

### Generating a genesis.json

create the `genesis.json`;

first you need to config the markerConfig.json like this:
```bash

{
  "AdminInfo": {
    "Account": "your Admin account keysore path",
    "Password": "your Admin account password"
  },
  "Groups": [
    {
      "Group": {
        "Account":  "your Group account keysore path",
        "Password":  "your Group account password"
      },
      "Validators": [
        {
          "Account": "your Validators account keysore path",
          "Password":"your Validators account password"
        },
        {
          "Account": "your Validators account keysore path",
          "Password":"your Validators account password"
        },
        {
          "Account":"your Validators account keysore path",
          "Password": "your Validators account password"
        },
        {
          "Account": "your Validators account keysore path",
          "Password":"your Validators account password"
        }
      ]
    }
  ]
}
```
then to do so run:

```bash
marker genesis --buildpath path/to/protocol/build
```

Where `buildpath` is the path to truffle compile output folder. By default it will use `MAP_CONTRACTS` environment variable as `$MAP_CONTRACTS/build/contracts`.

This will create a `genesis.json`.


### Configuring Genesis

Genesis creation has many configuration options, for that `marker` use the concept of templates.

```bash
marker genesis --template=[local|loadtest|monorepo]
```

Additionally, you can override template options via command line, chedk `marker genesis --help` for options:

```bash
   --validators value    Number of Validators (default: 0)
   --epoch value         Epoch size (default: 0)
```













