# chaincode

## DRAFT

## Few Notes

The current chaincode to use with IBM's bluemix blockchain is under the hyper folder. I am keeping the old code in the root directory solely because I haven't upgraded my dev enviornment to the hyper ledger space. Thus logic testing will be done using my old code. What all this means is that if you want to use my code for the blockchain, reference: 

```
chaincode-master/hyper
```

as the directory instead of the normal chaincode-master

Cheers.

## Explanation of how to set up the chaincode in a developer environment

Follow the guides at: 
* https://github.com/openblockchain/obc-docs/blob/master/dev-setup/devenv.md 
* https://github.com/openblockchain/obc-peer/blob/master/README.md
* https://github.com/openblockchain/obc-docs/blob/master/api/SandboxSetup.md

To set up your environment, and make sure you turn security on and privacy OFF. Otherwise majority of the invoke functions will FAIL

## Function Breakdown

### Deploy

The deploy functions are something you have to run first. This will be associtaed with the function **init** and will basically initialize where the commercial papers will be stored. Do this only once ever.

### Invoke

Invoke has a few functions, primarily creating an account as well as issuing the property tokens. The arguments that is taken in need to fit the mapping laid out in the beginning of the code.

#### issuePropertyToken

this command is called to create a property token. The structure of the object is shown as below:
```
type PTY struct {
	CUSIP		string 	   `json:"cusip"`
	Name		string 	   `json:"name"`
    AdrStreet   string     `json:"adrStreet"`
    AdrCity     string     `json:"adrCity"`
    AdrPostcode string     `json:"adrPostcode"`
    AdrState    string     `json:"adrState"`
    BuyValue    float64    `json:"buyval"`
    MktValue    float64    `json:"mktval"`
    Qty         int        `json:"quantity"`
    Owners      []Owner    `json:"owner"`
    PT4Sale     []ForSale  `json:"forsale"`
    Issuer      string     `json:"issuer"`
    IssueDate   string     `json:"issueDate"`
```
All of the data (with the exception of Owners and PT4Sale) 

You do not need to pass anything in for Owners or PT4Sale as it will automatically populate Owners

#### transferPaper

Transfers property tokens from a "ForSale" batch to an owner provided that enough funds are in the account balance. Transfers require a structure to be sent to the chaincode shown below

```
type Transaction struct {
	CUSIP       string   `json:"cusip"`
	FromCompany string   `json:"fromCompany"`
	ToCompany   string   `json:"toCompany"`
	Quantity    int      `json:"quantity"`
}
```

#### updateMktVal

Updates the market value of a certain property. JSON passed in will be in this format:

```
type UpdateMktVal struct {
    CUSIP       string   `json:"cusip"`
    MktValue    float64  `json:"mktval"`
}
```

#### createAccount

Creates an account for use on the blockchain. Takes in a name.

#### createAccounts

Takes in an int and creates users with the names company<num>. 

### Query

Query simply queries the blockchain for details. Note that the structure of this is to send two arguments. The first is the query function you want to run, the second is any other variable you may need to include. For functions like GetAllCPs this will just require a blank arugment, however for something like GetCompany you will need to provide the name of the company you're querying.

As a note all these queries will return a json. 

#### GetAllPTYs

Simply returns all property tokens. Does not require other arguments

#### GetCompany

Requires a second argument of the company you're querying

