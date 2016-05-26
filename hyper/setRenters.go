package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"errors"
)

func (t *SimpleChaincode) setRenters(stub *shim.ChaincodeStub, id string, fromAcct string, toAcct string) ([]byte, error) {

	if fromAcct == "" && toAcct == "" {
		return nil, errors.New("Invalid arguments passed - require at least 1 argument")
	}

	var p PTY
	var r Renter

	r.RenterID = toAcct

	propertyBytes, err := stub.GetState(ptyPrefix + id)
						    if err != nil {
						        fmt.Println("Error Getting state of - " + ptyPrefix + p.CUSIP)
						        return nil, errors.New("Error retrieving property " + p.CUSIP)
						    }
	err = json.Unmarshal(propertyBytes, &p)
							if err != nil {
						        fmt.Println("error invalid Data issue")
						        fmt.Println("error: ",err)
						        return nil, errors.New("Invalid Data issue")
						    }

	if fromAcct == "" && toAcct != "" {
		fmt.Println("New renter")
		for i:=0; i<len(p.Renters); i++ {
			if p.Renters[i].RenterID == toAcct {
				return nil, errors.New("Renter already exists")
			}
		}
		p.Renters = append(p.Renters, r)

		bytesToWrite, err := json.Marshal(p)
				            if err != nil {
				                fmt.Println("Error marshalling property struct")
				                return nil, errors.New("Error marshalling the keys")
				            }
		err = stub.PutState(ptyPrefix + id, bytesToWrite)
					        if err != nil {
					            fmt.Println("Error writting keys back")
					            return nil, errors.New("Error writing the keys back")
					        }
	} else if fromAcct != "" && toAcct == "" {
		fmt.Println("Removing Renter")
		for i:=0; i<len(p.Renters); i++ {
			if p.Renters[i].RenterID == fromAcct {
				p.Renters = append(p.Renters[:i], p.Renters[i+1:]...)
			}
		}

		bytesToWrite, err := json.Marshal(p)
				            if err != nil {
				                fmt.Println("Error marshalling property struct")
				                return nil, errors.New("Error marshalling the keys")
				            }
		err = stub.PutState(ptyPrefix + id, bytesToWrite)
					        if err != nil {
					            fmt.Println("Error writting keys back")
					            return nil, errors.New("Error writing the keys back")
					        }
	} else {
		fmt.Println("Transfer Renters")

		var fromExists = false
		var toExists = false

		for i:=0; i<len(p.Renters); i++ {
			if p.Renters[i].RenterID == fromAcct {
				fromExists = true
				p.Renters = append(p.Renters[:i], p.Renters[i+1:]...)
			}
		}

		for i:=0; i<len(p.Renters); i++ {
			if p.Renters[i].RenterID == toAcct {
				toExists = true
			}
		}

		if fromExists == true && toExists == false {
			p.Renters = append(p.Renters, r)
		} else if fromExists == true && toExists == true {
			fmt.Println("no need to do anything - renter already exists")
		} else {
			return nil, errors.New("Cannot find renter to replace")
		}

		bytesToWrite, err := json.Marshal(p)
				            if err != nil {
				                fmt.Println("Error marshalling property struct")
				                return nil, errors.New("Error marshalling the keys")
				            }
		err = stub.PutState(ptyPrefix + id, bytesToWrite)
					        if err != nil {
					            fmt.Println("Error writting keys back")
					            return nil, errors.New("Error writing the keys back")
					        }
	}

	return nil, nil	
}