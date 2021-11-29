## Validator CLI
### What is the Validator CLI​
The Command Line Interface allows users to interact with the Atlas Protocol smart contracts.

### Building ValidatorCli

```bash
go build -o ValidatorCli  *.go
```

###  registerGroup
```bash

registered Validator Group

USAGE
  $ ValidatorCli  registerGroup --rpcaddr localhost  --rpcport 8545 --keystore <your group keystore> 

OPTIONS
  --rpcaddr localhost                                        atlas host

  --rpcport                                                  atlas rpcport
                                                             
  --keystore                                                 your atlas account private key	file
  
  --password                                                 your atlas account password
  
  --commission                                               register group param,This represents the share of the epoch rewards given to elected Validators that goes to the group they are a member of

EXAMPLES
  ValidatorCli registerGroup --rpcaddr localhost  --rpcport 8545 --keystore /root/keystore/UTC--2021-07-19T02-07-11.808701800Z--3e3429f72450a39ce227026e8ddef331e9973e4d --password "123456"
  ```

###  registerValidator
```bash

registered Validator Validator

USAGE
  $ ValidatorCli  registerValidator --rpcaddr localhost  --rpcport 8545 --keystore <your group keystore> 

OPTIONS
  --rpcaddr localhost                                        atlas host

  --rpcport                                                  atlas rpcport
                                                             
  --keystore                                                 your atlas account private key	file
  
  --password                                                 your atlas account password

EXAMPLES
  ValidatorCli registerValidator --rpcaddr localhost  --rpcport 8545 --keystore /root/keystore/UTC--2021-07-19T02-09-17.552426700Z--81f02fd21657df80783755874a92c996749777bf --password "123456"
  ```

###  queryGroups
```bash

Retrun all of Groups address 

USAGE
  $ ValidatorCli  registerValidator --rpcaddr localhost  --rpcport 8545 --keystore <your group keystore> 

OPTIONS
  --rpcaddr localhost                                        atlas host

  --rpcport                                                  atlas rpcport
                                                             
  --keystore                                                 your atlas account private key	file
  
  --password                                                 your atlas account password

EXAMPLES
  ValidatorCli queryGroups --rpcaddr localhost  --rpcport 8545 --keystore /root/keystore/UTC--2021-07-19T02-09-17.552426700Z--81f02fd21657df80783755874a92c996749777bf  --password "123456"
  ```

###  getRegisteredValidatorSigners

  ```bash

Retrun all of Register address 

USAGE
  $ ValidatorCli  getRegisteredValidatorSigners --rpcaddr localhost  --rpcport 8545 --keystore <your group keystore> 

OPTIONS
  --rpcaddr localhost                                        atlas host

  --rpcport                                                  atlas rpcport
                                                             
  --keystore                                                 your atlas account private key file
  
  --password                                                 your atlas account password

EXAMPLES
  ValidatorCli registerValidator --rpcaddr localhost  --rpcport 8545 --keystore /root/keystore/UTC--2021-07-19T02-09-17.552426700Z--81f02fd21657df80783755874a92c996749777bf --password "123456"
  ```

###  getTopGroupValidators

  ```bash

Return the first topNum Validators at this group

USAGE
  $ ValidatorCli  getTopGroupValidators --rpcaddr localhost  --rpcport 8545 --keystore <your group keystore> --topNum 5

OPTIONS
  --rpcaddr localhost                                        atlas host

  --rpcport                                                  atlas rpcport
                                                             
  --keystore                                                 your atlas group account private key	file
  
  --password                                                 your atlas group account password
  
  --topNum                                                   param about top num Validators


EXAMPLES
  ValidatorCli registerValidator --rpcaddr localhost  --rpcport 8545  --topNum 5 --keystore /root/keystore/UTC--2021-07-19T02-09-17.552426700Z--81f02fd21657df80783755874a92c996749777bf --password "123456"
  ```



###  addFirstMember

  ```bash

add the first Validator to group

USAGE
  $ ValidatorCli  addFirstMember --rpcaddr localhost  --rpcport 8545 --keystore <your group keystore> 

OPTIONS
  --rpcaddr localhost                                        atlas host

  --rpcport                                                  atlas rpcport
                                                             
  --keystore                                                 your atlas group account private key	file
  
  --password                                                 your atlas group account password
  
  --readConfig                                               get validators by validatorCfg.json

EXAMPLES
  ValidatorCli addFirstMember --rpcaddr localhost  --rpcport 8545  --keystore /root/keystore/UTC--2021-07-19T02-09-17.552426700Z--81f02fd21657df80783755874a92c996749777bf --password "123456" --readConfig
  ```




###  addToGroup

  ```bash

add validator to group

USAGE
  $ ValidatorCli  addToGroup --rpcaddr localhost  --rpcport 8545 --keystore <your group keystore> 

OPTIONS
  --rpcaddr localhost                                        atlas host

  --rpcport                                                  atlas rpcport
                                                             
  --keystore                                                 your atlas group account private key	file
  
  --password                                                 your atlas group account password
  
  --readConfig                                               get validators by validatorCfg.json

EXAMPLES
  ValidatorCli addToGroup --rpcaddr localhost  --rpcport 8545  --keystore /root/keystore/UTC--2021-07-19T02-09-17.552426700Z--81f02fd21657df80783755874a92c996749777bf --password "123456" --readConfig
  ``` 



###  removeMember

  ```bash

remove the Validators in group

USAGE
  $ ValidatorCli  removeMember --rpcaddr localhost  --rpcport 8545 --keystore <your group keystore> 

OPTIONS
  --rpcaddr localhost                                        atlas host

  --rpcport                                                  atlas rpcport
                                                             
  --keystore                                                 your atlas group account private key	file
  
  --password                                                 your atlas group account password
  
  --readConfig                                               get validators by validatorCfg.json

EXAMPLES
  ValidatorCli removeMember --rpcaddr localhost  --rpcport 8545  --keystore /root/keystore/UTC--2021-07-19T02-09-17.552426700Z--81f02fd21657df80783755874a92c996749777bf --password "123456" --readConfig
  ``` 


###  deregisterValidatorGroup

```bash

deregister group account

USAGE
  $ ValidatorCli  deregisterValidatorGroup --rpcaddr localhost  --rpcport 8545 --keystore <your group keystore> 

OPTIONS
  --rpcaddr localhost                                        atlas host

  --rpcport                                                  atlas rpcport
                                                             
  --keystore                                                 your atlas group account private key	file
  
  --password                                                 your atlas group account password
  

EXAMPLES
  ValidatorCli deregisterValidatorGroup --rpcaddr localhost  --rpcport 8545  --keystore /root/keystore/UTC--2021-07-19T02-09-17.552426700Z--81f02fd21657df80783755874a92c996749777bf --password "123456"
``` 

###  deregisterValidator

```bash

deregister Validator account

USAGE
  $ ValidatorCli  deregisterValidator --rpcaddr localhost  --rpcport 8545 --keystore <your group keystore> 

OPTIONS
  --rpcaddr localhost                                        atlas host

  --rpcport                                                  atlas rpcport
                                                             
  --keystore                                                 your atlas Validator account private key	file
  
  --password                                                 your atlas Validator account password
  

EXAMPLES
  ValidatorCli deregisterValidator --rpcaddr localhost  --rpcport 8545  --keystore /root/keystore/UTC--2021-07-19T02-09-17.552426700Z--81f02fd21657df80783755874a92c996749777bf --password "123456"
``` 



###  setMaxGroupSize

```bash

set the max group size

USAGE
  $ ValidatorCli  setMaxGroupSize --rpcaddr localhost  --rpcport 8545 --keystore <your group keystore> 

OPTIONS
  --rpcaddr localhost                                        atlas host

  --rpcport                                                  atlas rpcport
                                                             
  --keystore                                                 your atlas group account private key	file
  
  --password                                                 your atlas group account password
  
  --maxSize                                                  the max group size
 
EXAMPLES
  ValidatorCli setMaxGroupSize --rpcaddr localhost  --rpcport 8545  --keystore /root/keystore/UTC--2021-07-19T02-09-17.552426700Z--81f02fd21657df80783755874a92c996749777bf --password "123456"
``` 
  
  
  
  
  
  
  
  
  
  
  
  
  
  
  
  
  
  
  
  
  
  
  
  