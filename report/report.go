package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode implementation
type SimpleChaincode struct {
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	var err error

	if len(args) == 0 {
		return nil, errors.New("Incorrect number of arguments. At least one Meter's name is required.")
	}

	for _, name := range args {
		if len(name) == 0 {
			continue
		}
		err = stub.PutState("kwh_"+name, []byte(strconv.Itoa(0)))
		if err != nil {
			return nil, errors.New("Meter cannot be created")
		}
		err = stub.PutState(name, []byte(strconv.Itoa(0)))
		if err != nil {
			return nil, errors.New("Meter cannot be created")
		}

		// meter status disable/enabled
		err = stub.PutState("status_"+name, []byte(strconv.FormatBool(true)))
		if err != nil {
			return nil, errors.New("Meter status cannot be created")
		}
	}

	return nil, nil
}

// Deletes an entity from state
func (t *SimpleChaincode) settle(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	//var err error
	var key string
	var val float64
	var exchange_rate, previous_val, amount float64

	exchange_rate = -1

	//TODO iteration by id is not an option. you should load keys from fabric
	for i := 1; i < 10; i++ {
		key = strconv.Itoa(i)
		value, err := stub.GetState("kwh_" + key)
		if err != nil {
			continue
		}
		if value == nil {
			continue
		}
		val, _ = strconv.ParseFloat(string(value), 64)
		amount = float64(val) * exchange_rate
		//f := "change"
		//queryArgs := []string{name,string(amount)}
		// // InvokeChaincode doesnt work. Submitted issues to fabric.
		//_, err := stub.InvokeChaincode("2780b7463c57f343a9e107854c4b53150018cdd8fd74ca970c028de6bfa707f6e9f6cf2b20f0af4fdd04d2167651eb29c7bfabf19e6a93ae2aff65f55202d0e6", f, queryArgs)
		//if err != nil {
		//	errStr := fmt.Sprintf("Failed to query chaincode. Got error: %s", err.Error())
		//	fmt.Printf(errStr)
		//	return nil, errors.New(errStr)
		//}
		coins, err := stub.GetState(key)
		if err != nil {
			jsonResp := "{\"Error\":\"Failed to get state for " + string(key) + "\"}"
			return nil, errors.New(jsonResp)
		}
		if coins == nil {
			previous_val = 0
		} else {
			previous_val, _ = strconv.ParseFloat(string(coins), 64)
		}

		err = stub.PutState(key, []byte(strconv.FormatFloat(amount+previous_val, 'f', 6, 64)))

		if err != nil {
			return nil, err
		}
		err = stub.PutState("kwh_"+key, []byte(strconv.Itoa(0)))
		if err != nil {
			return nil, errors.New("Meter cannot be updated")
		}
	}

	return nil, nil
}

func (t *SimpleChaincode) change(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	var name string
	var val, previous_val float64
	var err error

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}

	name = args[0]
	// Get the state from the ledger
	value, err := stub.GetState(name)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}
	if value == nil {
		previous_val = 0
	} else {
		previous_val, _ = strconv.ParseFloat(string(value), 64)
	}

	val, _ = strconv.ParseFloat(string(args[1]), 64)

	err = stub.PutState(name, []byte(strconv.FormatFloat(val+previous_val, 'f', 6, 64)))

	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (t *SimpleChaincode) setmeterstatus(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name string
	var val bool
	var err error

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}

	name = args[0]
	// Get the state from the ledger
	_, err = stub.GetState(name)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get stus for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}

	val, err = strconv.ParseBool(string(args[1]))
	if err != nil {
		jsonResp := "{\"Error\":\"Second argument to setmeterstatus should be boolean (false/true).\"}"
		return nil, errors.New(jsonResp)
	}

	err = stub.PutState("status_"+name, []byte(strconv.FormatBool(val)))

	if err != nil {
		jsonResp := "{\"Error\":\"Putstate failed.\"}"
		return nil, errors.New(jsonResp)
	}

	return nil, nil
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if function == "settle" {
		return t.settle(stub, args)
	}

	if function == "change" {
		return t.change(stub, args)
	}

	if function == "setmeterstatus" {
		return t.setmeterstatus(stub, args)
	}

	if function != "report" {
		return nil, errors.New("Unimplemented '" + function + "' invoked")
	}

	var name string // Entities
	var val float64 // Asset holdings
	var err error

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting     2")
	}

	name = args[0]
	val, _ = strconv.ParseFloat(string(args[1]), 64)

	err = stub.PutState("kwh_"+name, []byte(strconv.FormatFloat(val, 'f', 6, 64)))

	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Query callback representing the query of a chaincode
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	if function == "balance" {
		return t.balance(stub, args)
	}

	if function == "status" {
		return t.status(stub, args)
	}

	if function == "complete_meter_state" {
		return t.complete_meter_state(stub, args)
	}

	if function != "reported_kwh" {
		return nil, errors.New("Invalid query function name. Expecting \"reported_kwh\"")
	}

	return t.reported_kwh(stub, args)
}

func (t *SimpleChaincode) reported_kwh(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name string // Entities
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the Meter to query")
	}

	name = args[0]

	// Get the state from the ledger
	value, err := stub.GetState("kwh_" + name)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}

	if value == nil {
		jsonResp := "{\"Error\":\"Nil amount for Meter" + name + "\"}"
		return nil, errors.New(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + name + "\",\"Amount\":\"" + string(value) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	return value, nil
}

func (t *SimpleChaincode) balance(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name string // Entities
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the Meter to query")
	}

	name = args[0]

	// Get the state from the ledger
	value, err := stub.GetState(name)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}

	if value == nil {
		jsonResp := "{\"Error\":\"Nil amount for Meter " + name + "\"}"
		return nil, errors.New(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + name + "\",\"Amount\":\"" + string(value) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	return value, nil
}

// get meter status enabled/disabled
func (t *SimpleChaincode) status(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name string // Entities
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the Meter to query")
	}

	name = args[0]

	value, err := t.MeterEnabled(stub, name)

	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get status for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + name + "\",\"Status\":\"" + strconv.FormatBool(value) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)

	val := []byte(strconv.FormatBool(value))

	return val, nil
}

func (t *SimpleChaincode) complete_meter_state(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name string // Entities
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the Meter to query")
	}

	name = args[0]

	var balance []byte
	balance, err = t.balance(stub, args)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get complete state for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}

	var status []byte
	status, err = t.status(stub, args)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get complete state for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}

	var reported_kwh []byte
	reported_kwh, err = t.reported_kwh(stub, args)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get complete state for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}

	jsonMessage := "{"
	jsonMessage = jsonMessage + "\"name\":\"" + name + "\""
	jsonMessage = jsonMessage + ",\"balance\":\"" + string(balance) + "\""
	jsonMessage = jsonMessage + ",\"status\":\"" + string(status) + "\""
	jsonMessage = jsonMessage + ",\"reported_kwh\":\"" + string(reported_kwh) + "\""
	jsonMessage = jsonMessage + "}"

	return []byte(jsonMessage), nil
}

func (t *SimpleChaincode) MeterEnabled(stub shim.ChaincodeStubInterface, meterName string) (bool, error) {
	var name string // Entities
	var err error

	name = "status_" + meterName
	// Get the state from the ledger
	value, err := stub.GetState("status_" + meterName)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + name + "\"}"
		return false, errors.New(jsonResp)
	}

	if value == nil {
		jsonResp := "{\"Error\":\"Nil amount for name " + name + "\"}"
		return false, errors.New(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + name + "\",\"Amount\":\"" + string(value) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)

	value_b, _ := strconv.ParseBool(string(value))
	return value_b, nil
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
