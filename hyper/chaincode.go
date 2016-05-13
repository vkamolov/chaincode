package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

var ptyPrefix = "pty:"
var accountPrefix = "acct:"
var accountsKey = "accounts"

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
    Status      string     `json:"status"`
}

type IP struct {
    UNIQUE      string      `json:"uqe"`
    AssetID     string      `json:"aid"`
    Owners      []Owner     `json:"owner"`
}

type Owner struct {
	InvestorID string    `json:"invid"`
	Quantity int      `json:"quantity"`
}

type ForSale struct {
    InvestorID string   `json:"invid"`
    Quantity   int      `json:"quantity"`
    SellVal    float64  `json:"sellval"`
}

type Transaction struct {
	CUSIP       string   `json:"cusip"`
	FromCompany string   `json:"fromCompany"`
	ToCompany   string   `json:"toCompany"`
	Quantity    int      `json:"quantity"`
}

type AddForSale struct {
    CUSIP       string   `json:"cusip"`
    FromCompany string   `json:"fromCompany"`
    Quantity    int      `json:"quantity"`
    SellVal     float64  `json:"sellval"`
}

type Account struct {
	ID          string  `json:"id"`
	Prefix      string  `json:"prefix"`
    CashBalance float64 `json:"cashBalance"`
	AssetsIds   []string `json:"assetIds"`
}

type UpdateMktVal struct {
    CUSIP       string   `json:"cusip"`
    MktValue    float64  `json:"mktval"`
}

type SimpleChaincode struct {
}

const (
    millisPerSecond     = int64(time.Second / time.Millisecond)
    nanosPerMillisecond = int64(time.Millisecond / time.Nanosecond)
)

func msToTime(ms string) (time.Time, error) {
    msInt, err := strconv.ParseInt(ms, 10, 64)
    if err != nil {
        return time.Time{}, err
    }

    return time.Unix(msInt/millisPerSecond,
        (msInt%millisPerSecond)*nanosPerMillisecond), nil
}



func genHash(issueDate string, days int) (string, error) {

    t, err := msToTime(issueDate)
    if err != nil {
        return "", err
    }

    maturityDate := t.AddDate(0, 0, days)
    month := int(maturityDate.Month())
    day := maturityDate.Day()

    suffix := seventhDigit[month] + eigthDigit[day]
    return suffix, nil

}

func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
    // Initialize the collection of commercial paper keys
    fmt.Println("Initializing Property keys collection")
	var blank []string
	blankBytes, _ := json.Marshal(&blank)
	err := stub.PutState("PtyKeys", blankBytes)
    if err != nil {
        fmt.Println("Failed to initialize property key collection")
    }

	fmt.Println("Initialization complete")

	return nil, nil
}

func (t *SimpleChaincode) createAccounts(stub *shim.ChaincodeStub, args []string) ([]byte, error) {

    //                  0
    // "number of accounts to create"
    var err error
    numAccounts, err := strconv.Atoi(args[0])
    if err != nil {
        fmt.Println("error creating accounts with input")
        return nil, errors.New("createAccounts accepts a single integer argument")
    }
    //create a bunch of accounts
    var account Account
    counter := 1
    for counter <= numAccounts {
        var prefix string
        suffix := "000A"
        if counter < 10 {
            prefix = strconv.Itoa(counter) + "0" + suffix
        } else {
            prefix = strconv.Itoa(counter) + suffix
        }
        var assetIds []string
        account = Account{ID: "company" + strconv.Itoa(counter), Prefix: prefix, CashBalance: 10000000.0, AssetsIds: assetIds}
        accountBytes, err := json.Marshal(&account)
        if err != nil {
            fmt.Println("error creating account" + account.ID)
            return nil, errors.New("Error creating account " + account.ID)
        }
        err = stub.PutState(accountPrefix+account.ID, accountBytes)
        counter++
        fmt.Println("created account" + accountPrefix + account.ID)
    }

    fmt.Println("Accounts created")
    return nil, nil

}

func (t *SimpleChaincode) createAccount(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
    // Obtain the username to associate with the account
    if len(args) != 1 {
        fmt.Println("Error obtaining username")
        return nil, errors.New("createAccount accepts a single username argument")
    }
    username := args[0]
    fmt.Println(username)
    fmt.Println("thats the username!")
    // Build an account object for the user
    var assetIds []string
    suffix := "000A"
    prefix := username + suffix
    var account = Account{ID: username, Prefix: prefix, CashBalance: 10000000.0, AssetsIds: assetIds}
    accountBytes, err := json.Marshal(&account)
    fmt.Println("Creating accounts")
    if err != nil {
        fmt.Println("error creating account" + account.ID)
        return nil, errors.New("Error creating account " + account.ID)
    }
    
    fmt.Println("Attempting to get state of any existing account for " + account.ID)
    existingBytes, err := stub.GetState(accountPrefix + account.ID)
	if err == nil {
        
        var company Account
        err = json.Unmarshal(existingBytes, &company)
        if err != nil {
            fmt.Println("Error unmarshalling account " + account.ID + "\n--->: " + err.Error())
            
            if strings.Contains(err.Error(), "unexpected end") {
                fmt.Println("No data means existing account found for " + account.ID + ", initializing account.")
                err = stub.PutState(accountPrefix+account.ID, accountBytes)
                
                if err == nil {
                    fmt.Println("created account" + accountPrefix + account.ID)
                    return nil, nil
                } else {
                    fmt.Println("failed to create initialize account for " + account.ID)
                    return nil, errors.New("failed to initialize an account for " + account.ID + " => " + err.Error())
                }
            } else {
                return nil, errors.New("Error unmarshalling existing account " + account.ID)
            }
        } else {
            fmt.Println("Account already exists for " + account.ID + " " + company.ID)
		    return nil, errors.New("Can't reinitialize existing user " + account.ID)
        }
    } else {
        
        fmt.Println("No existing account found for " + account.ID + ", initializing account.")
        err = stub.PutState(accountPrefix+account.ID, accountBytes)
        
        if err == nil {
            fmt.Println("created account" + accountPrefix + account.ID)
            return nil, nil
        } else {
            fmt.Println("failed to create initialize account for " + account.ID)
            return nil, errors.New("failed to initialize an account for " + account.ID + " => " + err.Error())
        }
        
    }
    
    
}

func (t *SimpleChaincode) updateMktVal(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
    if len(args) != 1 {
        fmt.Println("error invalid arguments")
        return nil, errors.New("Incorrect number of arguments. Expecting commercial paper record")
    }

    /*
        type UpdateMktVal struct {
        CUSIP       string   `json:"cusip"`
        MktValue    float64  `json:"mktval"`
}   */

    var cp UpdateMktVal
    var err error

    var newstring = args[0]
    newstring = strings.Replace(args[0],"'","\"",-1)

    fmt.Println("Unmarshalling CP")
    err = json.Unmarshal([]byte(newstring), &cp)
    if err != nil {
        fmt.Println("error invalid paper issue")
        fmt.Println("error: ",err)
        return nil, errors.New("Invalid commercial paper issue")
    }


    fmt.Println("Getting State on CP " + cp.CUSIP)
    cpRxBytes, err := stub.GetState(ptyPrefix+cp.CUSIP)

    if cpRxBytes != nil {
        fmt.Println("CUSIP exists")
        
        var cprx PTY
        fmt.Println("Unmarshalling CP " + cp.CUSIP)
        err = json.Unmarshal(cpRxBytes, &cprx)
        if err != nil {
            fmt.Println("Error unmarshalling cp " + cp.CUSIP)
            return nil, errors.New("Error unmarshalling cp " + cp.CUSIP)
        }

        cprx.MktValue = cp.MktValue
        cprx.Status = "Approved"

        cpWriteBytes, err := json.Marshal(&cprx)
        if err != nil {
            fmt.Println("Error marshalling cp")
            return nil, errors.New("Error issuing commercial paper")
        }
        err = stub.PutState(ptyPrefix+cp.CUSIP, cpWriteBytes)
        if err != nil {
            fmt.Println("Error issuing paper")
            return nil, errors.New("Error issuing commercial paper")
        }

        fmt.Println("Updated commercial paper %+v\n", cprx)
        return nil, nil
    } else {
        return nil, errors.New("Could not find Property Token")
    }

}

func (t *SimpleChaincode) issuePropertyToken(stub *shim.ChaincodeStub, args []string) ([]byte, error) {

    /*      0
        json
        {
            "Name":  "name of the investment pool",
            "par": 0.00,
            "qty": 10,
            "discount": 7.5,
            "maturity": 30,
            "owners": [ // This one is not required
                {
                    "company": "company1",
                    "quantity": 5
                },
                {
                    "company": "company3",
                    "quantity": 3
                },
                {
                    "company": "company4",
                    "quantity": 2
                }
            ],              
            "issuer":"company2",
            "issueDate":"1456161763790"  (current time in milliseconds as a string)

        }
    */
    //need one arg
    if len(args) != 1 {
        fmt.Println("error invalid arguments")
        return nil, errors.New("Incorrect number of arguments. Expecting commercial paper record")
    }

    var cp PTY
    var err error
    var account Account

    var newstring = args[0]
    newstring = strings.Replace(args[0],"'","\"",-1)

    fmt.Println("Unmarshalling CP")
    err = json.Unmarshal([]byte(newstring), &cp)
    if err != nil {
        fmt.Println("error invalid paper issue")
        fmt.Println("error: ",err)
        return nil, errors.New("Invalid commercial paper issue")
    }

    fmt.Println("Hey guys, this is what we got:")
    fmt.Println("CP.name is   : ", cp.Name)
    fmt.Println("CP.Address is: ", cp.AdrStreet)
    fmt.Println("CP.Address is: ", cp.AdrCity)
    fmt.Println("CP.Address is: ", cp.AdrPostcode)
    fmt.Println("CP.Address is: ", cp.AdrState)
    if cp.CUSIP == "" {
        fmt.Println("No CUSIP, returning error")
        return nil, errors.New("CUSIP cannot be blank")
    }
    fmt.Println("Getting state of - " + accountPrefix + cp.Issuer)
    accountBytes, err := stub.GetState(accountPrefix + cp.Issuer)
    if err != nil {
        fmt.Println("Error Getting state of - " + accountPrefix + cp.Issuer)
        return nil, errors.New("Error retrieving account " + cp.Issuer)
    }
    err = json.Unmarshal(accountBytes, &account)
    if err != nil {
        fmt.Println("Error Unmarshalling accountBytes")
        return nil, errors.New("Error retrieving account " + cp.Issuer)
    }
    
    account.AssetsIds = append(account.AssetsIds, cp.CUSIP)

    var owner Owner
    owner.InvestorID = cp.Issuer
    owner.Quantity = cp.Qty

    cp.Owners = append(cp.Owners, owner)
    
    fmt.Println("Getting State on CP " + cp.CUSIP)
    cpRxBytes, err := stub.GetState(ptyPrefix+cp.CUSIP)
    if cpRxBytes == nil {
        fmt.Println("CUSIP does not exist, creating it")
        cpBytes, err := json.Marshal(&cp)
        if err != nil {
            fmt.Println("Error marshalling cp")
            return nil, errors.New("Error issuing commercial paper")
        }
        err = stub.PutState(ptyPrefix+cp.CUSIP, cpBytes)
        if err != nil {
            fmt.Println("Error issuing paper")
            return nil, errors.New("Error issuing commercial paper")
        }

        fmt.Println("Marshalling account bytes to write")
        accountBytesToWrite, err := json.Marshal(&account)
        if err != nil {
            fmt.Println("Error marshalling account")
            return nil, errors.New("Error issuing commercial paper")
        }
        err = stub.PutState(accountPrefix + cp.Issuer, accountBytesToWrite)
        if err != nil {
            fmt.Println("Error putting state on accountBytesToWrite")
            return nil, errors.New("Error issuing commercial paper")
        }
        
        
        // Update the paper keys by adding the new key
        fmt.Println("Getting Property Keys")
        keysBytes, err := stub.GetState("PtyKeys")
        if err != nil {
            fmt.Println("Error retrieving paper keys")
            return nil, errors.New("Error retrieving paper keys")
        }
        var keys []string
        err = json.Unmarshal(keysBytes, &keys)
        if err != nil {
            fmt.Println("Error unmarshel keys")
            return nil, errors.New("Error unmarshalling paper keys ")
        }
        
        fmt.Println("Appending the new key to Property Keys")
        foundKey := false
        for _, key := range keys {
            if key == ptyPrefix+cp.CUSIP {
                foundKey = true
            }
        }
        if foundKey == false {
            keys = append(keys, ptyPrefix+cp.CUSIP)
            keysBytesToWrite, err := json.Marshal(&keys)
            if err != nil {
                fmt.Println("Error marshalling keys")
                return nil, errors.New("Error marshalling the keys")
            }
            fmt.Println("Put state on Propert Keys")
            err = stub.PutState("PtyKeys", keysBytesToWrite)
            if err != nil {
                fmt.Println("Error writting keys back")
                return nil, errors.New("Error writing the keys back")
            }
        }
        fmt.Println("Issue commercial paper %+v\n", cp)
        fmt.Println("Whee done")
        return nil, nil
    } else {
        fmt.Println("You can't tokenize an asset that already exists")
        return nil, errors.New("Can't tokenize asset that already exists")
    }
}

func (t *SimpleChaincode) setForSale(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
    //   0
    // json
    // {
    //     CUSIP       string   `json:"cusip"`
    //     FromCompany string   `json:"fromCompany"`
    //     Quantity    int      `json:"quantity"`
    //     SellVal     float64  `json:"sellval"`
    // }

    //need one arg
    if len(args) != 1 {
        return nil, errors.New("Incorrect number of arguments. Expecting commercial paper record")
    }
    
    var fs AddForSale

    fmt.Println("Unmarshalling ForSale")
    err := json.Unmarshal([]byte(strings.Replace(args[0],"'","\"",-1)), &fs)
    if err != nil {
        fmt.Println("Error Unmarshalling ForSale")
        return nil, errors.New("Invalid forsale issue")
    }

    fmt.Println("Getting State on CP " + fs.CUSIP)
    cpBytes, err := stub.GetState(ptyPrefix+fs.CUSIP)
    if err != nil {
        fmt.Println("CUSIP not found")
        return nil, errors.New("CUSIP not found " + fs.CUSIP)
    }

    var cp PTY
    fmt.Println("Unmarshalling CP " + fs.CUSIP)
    err = json.Unmarshal(cpBytes, &cp)
    if err != nil {
        fmt.Println("Error unmarshalling cp " + fs.CUSIP)
        return nil, errors.New("Error unmarshalling cp " + fs.CUSIP)
    }

    var fromCompany Account
    fmt.Println("Getting State on fromCompany " + fs.FromCompany)   
    fromCompanyBytes, err := stub.GetState(accountPrefix+fs.FromCompany)
    if err != nil {
        fmt.Println("Account not found " + fs.FromCompany)
        return nil, errors.New("Account not found " + fs.FromCompany)
    }

    fmt.Println("Unmarshalling FromCompany ")
    err = json.Unmarshal(fromCompanyBytes, &fromCompany)
    if err != nil {
        fmt.Println("Error unmarshalling account " + fs.FromCompany)
        return nil, errors.New("Error unmarshalling account " + fs.FromCompany)
    }

    // Check for all the possible errors
    ownerFound := false 
    quantity := 0
    for _, owner := range cp.Owners {
        if owner.InvestorID == fs.FromCompany {
            ownerFound = true
            quantity = owner.Quantity
        }
    }
    
    // If fromCompany doesn't own this paper
    if ownerFound == false {
        fmt.Println("The company " + fs.FromCompany + "doesn't own any of this paper")
        return nil, errors.New("The company " + fs.FromCompany + "doesn't own any of this paper")   
    } else {
        fmt.Println("The FromCompany does own this paper")
    }
    
    // If fromCompany doesn't own enough quantity of this paper
    if quantity < fs.Quantity {
        fmt.Println("The company " + fs.FromCompany + "doesn't own enough of this paper")       
        return nil, errors.New("The company " + fs.FromCompany + "doesn't own enough of this paper")            
    } else {
        fmt.Println("The FromCompany owns enough of this paper")
    }

    FromOwnerFound := false
    for key, owner := range cp.Owners {
        if owner.InvestorID == fs.FromCompany {
            fmt.Println("Reducing Quantity from the FromCompany")
            cp.Owners[key].Quantity -= fs.Quantity
//          owner.Quantity -= fs.Quantity
        }
    }
    for key, forsale := range cp.PT4Sale {
        if (forsale.InvestorID == fs.FromCompany) {
            FromOwnerFound = true
            fmt.Println("Found company in For Sale")
            cp.PT4Sale[key].Quantity += fs.Quantity
            cp.PT4Sale[key].SellVal = fs.SellVal
        }
    }
    
    if FromOwnerFound == false {
        var newOwner ForSale
        fmt.Println("As FromOwner was not found in ForSale, appending the owner to the CP")
        newOwner.Quantity = fs.Quantity
        newOwner.InvestorID = fs.FromCompany
        newOwner.SellVal = fs.SellVal
        cp.PT4Sale = append(cp.PT4Sale, newOwner)
    }

    // Write everything back
    // To Company
        
    // From company
    fromCompanyBytesToWrite, err := json.Marshal(&fromCompany)
    if err != nil {
        fmt.Println("Error marshalling the fromCompany")
        return nil, errors.New("Error marshalling the fromCompany")
    }
    fmt.Println("Put state on fromCompany")
    err = stub.PutState(accountPrefix+fs.FromCompany, fromCompanyBytesToWrite)
    if err != nil {
        fmt.Println("Error writing the fromCompany back")
        return nil, errors.New("Error writing the fromCompany back")
    }
    
    // cp
    cpBytesToWrite, err := json.Marshal(&cp)
    if err != nil {
        fmt.Println("Error marshalling the cp")
        return nil, errors.New("Error marshalling the cp")
    }
    fmt.Println("Put state on CP")
    err = stub.PutState(ptyPrefix+fs.CUSIP, cpBytesToWrite)
    if err != nil {
        fmt.Println("Error writing the cp back")
        return nil, errors.New("Error writing the cp back")
    }
    
    fmt.Println("Successfully completed Invoke")
    return nil, nil
}

func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
    //need one arg
    if len(args) < 1 {
        return nil, errors.New("Incorrect number of arguments. Expecting ......")
    }

    if args[0] == "GetCompany" {
        fmt.Println("Getting the company")
        company, err := GetCompany(args[1], stub)
        if err != nil {
            fmt.Println("Error from getCompany")
            return nil, err
        } else {
            companyBytes, err1 := json.Marshal(&company)
            if err1 != nil {
                fmt.Println("Error marshalling the company")
                return nil, err1
            }   
            fmt.Println("All success, returning the company")
            return companyBytes, nil         
        }
    } else if args[0] == "GetAllPTYs" {
        fmt.Println("Getting all CPs")
        allCPs, err := GetAllPTYs(stub)
        if err != nil {
            fmt.Println("Error from GetAllPTYs")
            return nil, err
        } else {
            allCPsBytes, err1 := json.Marshal(&allCPs)
            if err1 != nil {
                fmt.Println("Error marshalling allptys")
                return nil, err1
            }   
            fmt.Println("All success, returning allptys")
            return allCPsBytes, nil      
        }
    } else {
        fmt.Println("I don't do shit!")
        fmt.Println("Generic Query call")
        bytes, err := stub.GetState(args[0])

        if err != nil {
            fmt.Println("Some error happenend")
            return nil, errors.New("Some Error happened")
        }

        fmt.Println("All success, returning from generic")
        return bytes, nil       
    }

    
    // if args[0] == "GetAllPTYs" {
    //     fmt.Println("Getting all CPs")
    //     allCPs, err := GetAllPTYs(stub)
    //     if err != nil {
    //         fmt.Println("Error from GetAllPTYs")
    //         return nil, err
    //     } else {
    //         allCPsBytes, err1 := json.Marshal(&allCPs)
    //         if err1 != nil {
    //             fmt.Println("Error marshalling allcps")
    //             return nil, err1
    //         }   
    //         fmt.Println("All success, returning allcps")
    //         return allCPsBytes, nil      
    //     }
    // }
    return nil, nil
}

func (t *SimpleChaincode) transferPaper(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
    /*      0
        json
        {
              "CUSIP": "",
              "fromCompany":"",
              "toCompany":"",
              "quantity": 1
        }
    */
    //need one arg
    if len(args) != 1 {
        return nil, errors.New("Incorrect number of arguments. Expecting commercial paper record")
    }
    
    var tr Transaction

    fmt.Println("Unmarshalling Transaction")
    err := json.Unmarshal([]byte(strings.Replace(args[0],"'","\"",-1)), &tr)
    if err != nil {
        fmt.Println("Error Unmarshalling Transaction")
        fmt.Println("err: ", err)
        return nil, errors.New("Invalid commercial paper issue")
    }

    fmt.Println("Getting State on CP " + tr.CUSIP)
    cpBytes, err := stub.GetState(ptyPrefix+tr.CUSIP)
    if err != nil {
        fmt.Println("CUSIP not found")
        return nil, errors.New("CUSIP not found " + tr.CUSIP)
    }

    var cp PTY
    fmt.Println("Unmarshalling CP " + tr.CUSIP)
    err = json.Unmarshal(cpBytes, &cp)
    if err != nil {
        fmt.Println("Error unmarshalling cp " + tr.CUSIP)
        return nil, errors.New("Error unmarshalling cp " + tr.CUSIP)
    }

    var fromCompany Account
    fmt.Println("Getting State on fromCompany " + tr.FromCompany)   
    fromCompanyBytes, err := stub.GetState(accountPrefix+tr.FromCompany)
    if err != nil {
        fmt.Println("Account not found " + tr.FromCompany)
        return nil, errors.New("Account not found " + tr.FromCompany)
    }

    fmt.Println("Unmarshalling FromCompany ")
    err = json.Unmarshal(fromCompanyBytes, &fromCompany)
    if err != nil {
        fmt.Println("Error unmarshalling account " + tr.FromCompany)
        return nil, errors.New("Error unmarshalling account " + tr.FromCompany)
    }

    var toCompany Account
    fmt.Println("Getting State on ToCompany " + tr.ToCompany)
    toCompanyBytes, err := stub.GetState(accountPrefix+tr.ToCompany)
    if err != nil {
        fmt.Println("Account not found " + tr.ToCompany)
        return nil, errors.New("Account not found " + tr.ToCompany)
    }

    fmt.Println("Unmarshalling tocompany")
    err = json.Unmarshal(toCompanyBytes, &toCompany)
    if err != nil {
        fmt.Println("Error unmarshalling account " + tr.ToCompany)
        return nil, errors.New("Error unmarshalling account " + tr.ToCompany)
    }

    // Check for all the possible errors
    ownerFound := false 
    quantity := 0
    price := 0.00
    for _, owner := range cp.PT4Sale {
        if owner.InvestorID == tr.FromCompany {
            ownerFound = true
            quantity = owner.Quantity
            price = owner.SellVal
        }
    }
    
    // If fromCompany doesn't own this paper
    if ownerFound == false {
        fmt.Println("The company " + tr.FromCompany + "doesn't own any of this paper")
        return nil, errors.New("The company " + tr.FromCompany + "doesn't own any of this paper")   
    } else {
        fmt.Println("The FromCompany does own this paper")
    }
    
    // If fromCompany doesn't own enough quantity of this paper
    if quantity < tr.Quantity {
        fmt.Println("The company " + tr.FromCompany + "doesn't own enough of this paper")       
        return nil, errors.New("The company " + tr.FromCompany + "doesn't own enough of this paper")            
    } else {
        fmt.Println("The FromCompany owns enough of this paper")
    }
    
    amountToBeTransferred := float64(tr.Quantity) * price
    
    // If toCompany doesn't have enough cash to buy the papers
    if toCompany.CashBalance < amountToBeTransferred {
        fmt.Println("The company " + tr.ToCompany + "doesn't have enough cash to purchase the papers")      
        return nil, errors.New("The company " + tr.ToCompany + "doesn't have enough cash to purchase the papers")   
    } else {
        fmt.Println("The ToCompany has enough money to be transferred for this paper")
    }

    // Checking to see if the shares are revoked
    if tr.FromCompany != tr.ToCompany {
        toCompany.CashBalance -= amountToBeTransferred
        fromCompany.CashBalance += amountToBeTransferred
    }

    toOwnerFound := false
    for key, owner := range cp.PT4Sale {
        if owner.InvestorID == tr.FromCompany {
            fmt.Println("Reducing Quantity from the FromCompany")
            cp.PT4Sale[key].Quantity -= tr.Quantity
//          owner.Quantity -= tr.Quantity
        }
        
    }

    for key, owner := range cp.Owners {
        if owner.InvestorID == tr.ToCompany {
            fmt.Println("Increasing Quantity from the ToCompany")
            toOwnerFound = true
            cp.Owners[key].Quantity += tr.Quantity
//          owner.Quantity += tr.Quantity
        }
    }
    
    if toOwnerFound == false {
        var newOwner Owner
        fmt.Println("As ToOwner was not found, appending the owner to the CP")
        newOwner.Quantity = tr.Quantity
        newOwner.InvestorID = tr.ToCompany
        cp.Owners = append(cp.Owners, newOwner)
    }
    
    fromCompany.AssetsIds = append(fromCompany.AssetsIds, tr.CUSIP)

    // Write everything back
    // To Company
    toCompanyBytesToWrite, err := json.Marshal(&toCompany)
    if err != nil {
        fmt.Println("Error marshalling the toCompany")
        return nil, errors.New("Error marshalling the toCompany")
    }
    fmt.Println("Put state on toCompany")
    err = stub.PutState(accountPrefix+tr.ToCompany, toCompanyBytesToWrite)
    if err != nil {
        fmt.Println("Error writing the toCompany back")
        return nil, errors.New("Error writing the toCompany back")
    }
        
    // From company
    fromCompanyBytesToWrite, err := json.Marshal(&fromCompany)
    if err != nil {
        fmt.Println("Error marshalling the fromCompany")
        return nil, errors.New("Error marshalling the fromCompany")
    }
    fmt.Println("Put state on fromCompany")
    err = stub.PutState(accountPrefix+tr.FromCompany, fromCompanyBytesToWrite)
    if err != nil {
        fmt.Println("Error writing the fromCompany back")
        return nil, errors.New("Error writing the fromCompany back")
    }
    
    // cp
    cpBytesToWrite, err := json.Marshal(&cp)
    if err != nil {
        fmt.Println("Error marshalling the cp")
        return nil, errors.New("Error marshalling the cp")
    }
    fmt.Println("Put state on CP")
    err = stub.PutState(ptyPrefix+tr.CUSIP, cpBytesToWrite)
    if err != nil {
        fmt.Println("Error writing the cp back")
        return nil, errors.New("Error writing the cp back")
    }
    
    fmt.Println("Successfully completed Invoke")
    return nil, nil
}

func GetAllPTYs(stub *shim.ChaincodeStub) ([]PTY, error){
    
    var allCPs []PTY
    
    // Get list of all the keys
    keysBytes, err := stub.GetState("PtyKeys")
    if err != nil {
        fmt.Println("Error retrieving Property keys")
        return nil, errors.New("Error retrieving Property keys")
    }
    var keys []string
    err = json.Unmarshal(keysBytes, &keys)
    if err != nil {
        fmt.Println("Error unmarshalling Property keys")
        return nil, errors.New("Error unmarshalling Property keys")
    }

    // Get all the cps
    for _, value := range keys {
        cpBytes, err := stub.GetState(value)
        
        var cp PTY
        err = json.Unmarshal(cpBytes, &cp)
        if err != nil {
            fmt.Println("Error retrieving cp " + value)
            return nil, errors.New("Error retrieving cp " + value)
        }
        
        fmt.Println("Appending CP" + value)
        allCPs = append(allCPs, cp)
    }   
    
    return allCPs, nil
}

func GetCompany(companyID string, stub *shim.ChaincodeStub) (Account, error){
    var company Account
    companyBytes, err := stub.GetState(accountPrefix+companyID)
    if err != nil {
        fmt.Println("Account not found " + companyID)
        return company, errors.New("Account not found " + companyID)
    }

    err = json.Unmarshal(companyBytes, &company)
    if err != nil {
        fmt.Println("Error unmarshalling account " + companyID + "\n err:" + err.Error())
        return company, errors.New("Error unmarshalling account " + companyID)
    }
    
    return company, nil
}

// Run callback representing the invocation of a chaincode
// This chaincode will manage two accounts A and B and will fsansfer X units from A to B upon invoke
func (t *SimpleChaincode) Run(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {


    fmt.Println("run is running " + function)
    return t.Invoke(stub, function, args)
    // Handle different functions
    // if function == "init" {
    //     // Initialize the entities and their asset holdings
    //     return t.init(stub,"init", args)
    // } else if function == "issuePropertyToken" {
    //     // transaction makes payment of X units from A to B
    //     return t.issuePropertyToken(stub, args)
    // } else if function == "createAccount" {
    //     // Deletes an entity from its state
    //     return t.createAccount(stub, args)
    // } else if function == "createAccounts" {
    //     // Deletes an entity from its state
    //     return t.createAccounts(stub, args)
    // } else if function == "setForSale" {
    //     // Deletes an entity from its state
    //     return t.setForSale(stub, args)
    // } else if function == "transferPaper" {
    //     // Deletes an entity from its state
    //     return t.transferPaper(stub, args)
    // } else if function == "updateMktVal" {
    //     // Deletes an entity from its state
    //     return t.updateMktVal(stub, args)
    // }

    // return nil, errors.New("Received unknown function invocation")
}

func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
    fmt.Println("invoke is running " + function)

    if function == "Init" {
        // Initialize the entities and their asset holdings
        return t.Init(stub,"init", args)
    } else if function == "issuePropertyToken" {
        // transaction makes payment of X units from A to B
        return t.issuePropertyToken(stub, args)
    } else if function == "createAccount" {
        // Deletes an entity from its state
        return t.createAccount(stub, args)
    } else if function == "createAccounts" {
        // Deletes an entity from its state
        return t.createAccounts(stub, args)
    } else if function == "setForSale" {
        // Deletes an entity from its state
        return t.setForSale(stub, args)
    } else if function == "transferPaper" {
        // Deletes an entity from its state
        fmt.Println("firing transferPaper")
        return t.transferPaper(stub, args)
    } else if function == "updateMktVal" {
        // Deletes an entity from its state
        return t.updateMktVal(stub, args)
    }

    fmt.Println("Function"+ function +" was not found under invocation")
    return nil, errors.New("Received unknown function invocation")
}

func main() {
    err := shim.Start(new(SimpleChaincode))
    if err != nil {
        fmt.Println("Error starting Simple chaincode: %s", err)
    }
}

var seventhDigit = map[int]string{
    1:  "A",
    2:  "B",
    3:  "C",
    4:  "D",
    5:  "E",
    6:  "F",
    7:  "G",
    8:  "H",
    9:  "J",
    10: "K",
    11: "L",
    12: "M",
    13: "N",
    14: "P",
    15: "Q",
    16: "R",
    17: "S",
    18: "T",
    19: "U",
    20: "V",
    21: "W",
    22: "X",
    23: "Y",
    24: "Z",
}

var eigthDigit = map[int]string{
    1:  "1",
    2:  "2",
    3:  "3",
    4:  "4",
    5:  "5",
    6:  "6",
    7:  "7",
    8:  "8",
    9:  "9",
    10: "A",
    11: "B",
    12: "C",
    13: "D",
    14: "E",
    15: "F",
    16: "G",
    17: "H",
    18: "J",
    19: "K",
    20: "L",
    21: "M",
    22: "N",
    23: "P",
    24: "Q",
    25: "R",
    26: "S",
    27: "T",
    28: "U",
    29: "V",
    30: "W",
    31: "X",
}
