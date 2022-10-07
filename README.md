# A Toolbox for Migrating the Blockchain-based Application from Ethereum to Hyperledger Fabric

It is a toolbox designed for developers who want to migrate the blockchain-based applications from the Ethereum public blockchain to the permissioned blockchain built by Hyperledger Fabric.

The low transaction capacity (i.e., 15 transaction per second (TPS)), high transaction cost, and long-term privacy concerns of the current Ethereum are forcing developers to seek alternative blockchain platforms to migrate their application in order to reduce their applications' use-cost and improve their applications' user experience.

Hyperledger Fabric (HLF) platform introduces a new modular and extensible blockchain architecture with resiliency, flexibility, scalability, and confidentiality. It allows developers to flexibly select appropriate components, such as consensus protocols and membership services, according to their specific requirements. Compared to the Ethereum platform, developers can use the HLF platform to build blockchain-based applications with higher transaction capacity, lower use-cost, and better  privacy protection capabilities.

Hence, we designed a toolbox to help developers automatically migrate their applications from Ethereum to HLF. Specially, the toolbox provides following functions to ease the migration process

- Account Migration

  Ethereum is a public blockchain. The user can use almost all blockchain-based applications on the Ethereum platform as long as he/she generates a pair of public/private keys using the wallet (e.g., MetaMask).  On the contray, the blockchain built by HLF is a permissioned blockchain where the user has to obtain a valid and authorized identity in the permissioned network. Hence, The public key generated by the user’s Ethereum wallet will be not recognized after the application migration.  To solve the problem, the toolbox provides the account migration function to enable users to continue to use their identities.

- Interaction Transformation

  The blockchain based application on the Ethereum platform typically uses specific libraries (e.g., Web3.js or Ethers.js) to interact with the Ethereum node. If the migrated application still uses those libraries, it will fail to interact with the node in the HLF blockchain network, which may prevent the application from working properly.  To solve the problem, the toolbox provides automatic interaction transformation.

- Smart Contract Migration

  Besides migrating the Ethereum smart contract code, the toolbox can be used to migrate the state of the smart contract.

- Binding Mechanism

  Integrating the EVM into the HLF platform results in two smart contract deployment methods on the HLF platform.  One is the smart
  contract deployment method of the Ethereum platform, and the other is the smart contract deployment method of the HLF platform. On the Ethereum platform, any smart contract code of the blockchain-based application is allowed to be deployed without reviewing and testing. The platform relies on the gas fee mechanism to defend against the malicious codes (e.g., codes containing infinite loop) that may stall the platform. On the HLF platform, the gas fee mechanism is removed to reduce the application’s use-cost. The platform relies on mandatory code checks before deploying the smart contract to detect the malicious codes . Suppose the HLF platform integrates the EVM to support the application migration. In that case, attackers can bypass the code checks and inject the malicious codes into the HLF platform by selecting the smart contract deployment method of the Ethereum platform. Without the protection of the gas fee mechanism, any node in the HLF will be stalled by running the malicious codes, thereby destroying the application’s availability. To solve the problem, the toolbox will generates a chaincode and binds it with the migrated smart contract.

  ![image-20221007092643929](C:\Users\desly\AppData\Roaming\Typora\typora-user-images\image-20221007092643929.png)

## Deploying the toolbox

1. Bring up a permission blockchain network using HLF

   Please refer to [using the fabric test network](https://hyperledger-fabric.readthedocs.io/en/latest/test_network.html) and [deploying a production network](https://hyperledger-fabric.readthedocs.io/en/latest/deployment_guide_overview.html#step-one-decide-on-your-network-configuration).

2. Install and instantiate the EVM written in chaincode to the network

   the EVM chaincode (evmcc) refer to the [emvcc](https://github.com/zhaizhonghao/toolbox_migration/blob/main/evmcc/evmcc.go). Below is an example of installation and instantiation through the peer cli.

   ```shell
   peer chaincode install -n evmcc -l golang -v 0 -p github.com/zhaizhonghao/toolbox_migration/evmcc
   peer chaincode instantiate -n evmcc -v 0 -C <channel-name> -c '{"Args":[]}' -o <orderer-address> --tls --cafile <orderer-ca>
   ```

3. Run the interaction transformer

   The interaction transformer, called Fab3,  utilizes the SDK (e.g., Fabric Nodejs SDK) provided by the HLF platform to re-implement a new library. The library exposes the same APIs as the library provided by the Ethereum platform. The migrated application can use those APIs to interact with the HLF network. The APIs are refer to [ethservice](https://github.com/zhaizhonghao/toolbox_migration/blob/main/fab3/ethservice.go).

   To create the Fab3 binary, the repository must be checked out in the GOPATH. Noting that compiling Fab3 requires golang version 1.11 and up. 

   ```shell
   mkdir -p $(go env GOPATH)/src/github.com/hyperledger/
   git clone https://github.com/zhaizhonghao/toolbox_migration.git $(go env GOPATH)/src/github.com/zhaizhonghao/toolbox_migration
   cd $(go env GOPATH)/src/github.com/zhaizhonghao/toolbox_migration
   ```

   Run following command at the root of this repository:

   ```shell
   make Fab3
   ```

   A binary named `fab3 ` will be created in the bin directory. 
   
   To use Fab3, the user needs to provide a Fabric SDK Config and Fabric user information. To specify the Fabric user, the organization and user id needs to be provided which corresponds to the credentials provided in the SDK config. We provide a sample [config](examples/first-network-sdk-config.yaml) that can be used with the first network example from the [fabric-samples](https://github.com/hyperledger/fabric-samples) repository. The credentials specified in the config, are expected to be in the directory format that the [cryptogen](https://hyperledger-fabric.readthedocs.io/en/release-1.4/commands/cryptogen.html) binary outputs.   

## Migrate the smart contract code

Migrate Ethereum smart contract code by invoking the evmcc. Noting that `To` field is the zero address and the `Data` field is the Ethereum smart contract's bytecode to be migrated.

```
peer chaincode invoke -n evmcc -C <channel-name> -c '{"Args":["0000000000000000000000000000000000000000",<compiled-bytecode>]}' -o <orderer-address> --tls --cafile <orderer-ca>
```

## Test the migrated Ethereum blockchain-based application

1. install the web3

   ```
   npm install web3@0.20.2
   ```

2. Invoke the migrated blockchain-based application deployed in the blockchain built by HLF

   ```js
   Web3 = require('web3')
   ...
   # port 5000 is exposed by the HLF blockchain network
   web3 = new Web3(new Web3.providers.HttpProvider('http://localhost:5000'))
   ```

   
