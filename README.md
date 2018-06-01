 
# LemoChain core

LemoChain is a decentralized, open source platform for companies of all sizes to monetize and exchange their structured business data. 
LemoChain will accelerate blockchain’s integration into our every-day lives by means of increasing it's universal commercial relevance.
   
Many projects are trying to use blockchain technology to create a C2C market place for individuals to sell their own data. The problem
wesee that in order to achieve that goal, the first step is to create the technology for B2B market. The value in 1 persons data is not
high, and in order for there to be any real value, there would need to be a market of millions and millions of users. What we are trying 
to do is provide businesses of all sizes the ability to exchange their data with another business, allowing them to find new customers, 
lower costs and profit from the data exchanges. 
    
We aim to improve the security of existing data purchasing platforms through the removal of a hackable centralized data storage silo 
and instead, implementing the distributed IPFS storage method. Additionally, we want to improve the efficiency and profitability of data 
purchase; through secure multi-party computation and homomorphic encryption, we will integrate a ‘data matchmaking’ ecosystem, whereby 
only the relevant data is transferred between ideal candidates, instead of purchasing a large packet of data where only some of which is 
useful.

## Installation Instructions
```
go install -v ./cmd/glemo
```
Start up LemoChain's built-in interactive JavaScript console, (via the trailing `console` subcommand) through which you can invoke all official `lemo` methods. You can simply interact with the LemoChain network: create accounts; transfer funds; deploy and interact with contracts. To do so:
```
$ glemo console
```
...

## Thank you Ethereum
Ethereum is an outstanding open source blockchain project, with Turing complete virtual machine, convenient account data storage mechanism, and is a proven mature and stable system. However, Ethereum also has obvious problems of low throughput and low consensus efficiency. In order to create prototype rapidly and fast verify our innovative consensus algorithms and application scenarios, LemoChain will perform in-depth optimization based on Ethereum 1.8.3, gradually replacing its consensus mechanism, introducing new smart contract language, and transforming the account system. And it's highly possible we will implement more thorough reconstruction of the code.

Thanks to all contributors of the Ethereum Project.
