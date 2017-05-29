/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"errors"
	"fmt"
	"strconv"
	"encoding/json"
	"time"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

var dendataIndexStr = "_dendataindex"				//sysdenno for the key/value that will store a list of all known dendata
var openTradesStr = "_opentrades"				//sysdenno for the key/value that will store all open trades

type Dendata struct{
	Datsyu string `json:"datsyu"`
	Sysdenno string `json:"sysdenno"`					//the fieldtags are needed to keep case from bouncing around
	Tokcd string `json:"tokcd"`
	Bmncd string `json:"bmncd"`
	Hatdt string `json:"hatdt"`
	Noudt string `json:"noudt"`
	Nouydt string `json:"nouydt"`
	Bumcd string `json:"bumcd"`
	Msecd string `json:"msecd"`
	Msenm string `json:"msenm"`
	Kentencd string `json:"kentencd"`
	Kentennm string `json:"kentennm"`
	Denno string `json:"denno"`
	Denkb string `json:"denkb"`
	Binkb string `json:"binkb"`
	Sumsu int `json:"sumsu"`
	Sumgenkn int `json:"sumgenkn"`
	Sumurikn int `json:"sumurikn"`
	Gyono string `json:"gyono"`
	Shocd string `json:"shocd"`
	Jancd string `json:"jancd"`
	Shonma string `json:"shonma"`
	Shonmb string `json:"shonmb"`
	Untnm string `json:"untnm"`
	Irisu int `json:"irisu"`
	Cassu int `json:"cassu"`
	Gentk int `json:"gentk"`
	Uritk int `json:"uritk"`
	Suryo int `json:"suryo"`
	Genkn int `json:"genkn"`
	Urikn int `json:"urikn"`
	Riykb string `json:"riykb"`
}

type Description struct{
	Tokcd string `json:"tokcd"`
	Bmncd string `json:"bmncd"`
	Denno string `json:"denno"`
	Hatdt string `json:"hatdt"`
	Noudt string `json:"noudt"`
	Msecd string `json:"msecd"`
	Msenm string `json:"msenm"`
}

type AnOpenTrade struct{
	User string `json:"user"`					//user who created the open trade order
	Timestamp int64 `json:"timestamp"`			//utc timestamp of creation
	Want Description  `json:"want"`				//description of desired dendata
	Willing []Description `json:"willing"`		//array of dendata willing to trade away
}

type AllTrades struct{
	OpenTrades []AnOpenTrade `json:"open_trades"`
}

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// ============================================================================================================================
// Init - reset all the things
// ============================================================================================================================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	var Aval int
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	// Initialize the chaincode
	Aval, err = strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New("Expecting integer value for asset holding")
	}

	// Write the state to the ledger
	err = stub.PutState("abc", []byte(strconv.Itoa(Aval)))				//making a test var "abc", I find it handy to read/write to it right away to test the network
	if err != nil {
		return nil, err
	}
	
	var empty []string
	jsonAsBytes, _ := json.Marshal(empty)								//marshal an emtpy array of strings to clear the index
	err = stub.PutState(dendataIndexStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}
	
	var trades AllTrades
	jsonAsBytes, _ = json.Marshal(trades)								//clear the open trade struct
	err = stub.PutState(openTradesStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}
	
	return nil, nil
}

// ============================================================================================================================
// Run - Our entry point for Invocations - [LEGACY] obc-peer 4/25/2016
// ============================================================================================================================
func (t *SimpleChaincode) Run(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("run is running " + function)
	return t.Invoke(stub, function, args)
}

// ============================================================================================================================
// Invoke - Our entry point for Invocations
// ============================================================================================================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" {													//initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
	} else if function == "delete" {										//deletes an entity from its state
		res, err := t.Delete(stub, args)
		cleanTrades(stub)													//lets make sure all open trades are still valid
		return res, err
	} else if function == "write" {											//writes a value to the chaincode state
		return t.Write(stub, args)
	} else if function == "init_dendata" {									//create a new dendata
		return t.init_dendata(stub, args)
//	} else if function == "set_user" {										//change owner of a dendata
//		res, err := t.set_user(stub, args)
//		cleanTrades(stub)													//lets make sure all open trades are still valid
//		return res, err
//	} else if function == "open_trade" {									//create a new trade order
//		return t.open_trade(stub, args)
//	} else if function == "perform_trade" {									//forfill an open trade order
//		res, err := t.perform_trade(stub, args)
//		cleanTrades(stub)													//lets clean just in case
//		return res, err
//	} else if function == "remove_trade" {									//cancel an open trade order
//		return t.remove_trade(stub, args)
	}
	fmt.Println("invoke did not find func: " + function)					//error

	return nil, errors.New("Received unknown function invocation")
}

// ============================================================================================================================
// Query - Our entry point for Queries
// ============================================================================================================================
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" {													//read a variable
		return t.read(stub, args)
	}
	fmt.Println("query did not find func: " + function)						//error

	return nil, errors.New("Received unknown function query")
}

// ============================================================================================================================
// Read - read a variable from chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var sysdenno, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting sysdenno of the var to query")
	}

	sysdenno = args[0]
	valAsbytes, err := stub.GetState(sysdenno)									//get the var from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + sysdenno + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil													//send it onward
}

// ============================================================================================================================
// Delete - remove a key/value pair from state
// ============================================================================================================================
func (t *SimpleChaincode) Delete(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}
	
	sysdenno := args[0]
	err := stub.DelState(sysdenno)													//remove the key from chaincode state
	if err != nil {
		return nil, errors.New("Failed to delete state")
	}

	//get the dendata index
	dendataAsBytes, err := stub.GetState(dendataIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get dendata index")
	}
	var dendataIndex []string
	json.Unmarshal(dendataAsBytes, &dendataIndex)								//un stringify it aka JSON.parse()
	
	//remove dendata from index
	for i,val := range dendataIndex{
		fmt.Println(strconv.Itoa(i) + " - looking at " + val + " for " + sysdenno)
		if val == sysdenno{															//find the correct dendata
			fmt.Println("found dendata")
			dendataIndex = append(dendataIndex[:i], dendataIndex[i+1:]...)			//remove it
			for x:= range dendataIndex{											//debug prints...
				fmt.Println(string(x) + " - " + dendataIndex[x])
			}
			break
		}
	}
	jsonAsBytes, _ := json.Marshal(dendataIndex)									//save new index
	err = stub.PutState(dendataIndexStr, jsonAsBytes)
	return nil, nil
}

// ============================================================================================================================
// Write - write variable into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) Write(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var sysdenno, value string // Entities
	var err error
	fmt.Println("running write()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. sysdenno of the variable and value to set")
	}

	sysdenno = args[0]															//renumber for funsies
	value = args[1]
	err = stub.PutState(sysdenno, []byte(value))								//write the variable into the chaincode state
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// ============================================================================================================================
// Init dendata - create a new dendata, store into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) init_dendata(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error

//	//   0       1       2     3
//	// "asdf", "blue", "35", "bob"
	if len(args) != 32 {
		return nil, errors.New("Incorrect number of arguments. Expecting 4")
	}

	//input sanitation
	fmt.Println("- start init dendata")
	if len(args[0]) <= 0 {
		return nil, errors.New("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return nil, errors.New("2nd argument must be a non-empty string")
	}
//	if len(args[2]) <= 0 {
//		return nil, errors.New("3rd argument must be a non-empty string")
//	}
//	if len(args[3]) <= 0 {
//		return nil, errors.New("4th argument must be a non-empty string")
//	}
	if len(args[18]) <= 0 {
		return nil, errors.New("18th argument must be a non-empty string")
	}
	if len(args[19]) <= 0 {
		return nil, errors.New("19th argument must be a non-empty string")
	}
	datsyu   := args[0]
	sysdenno := args[1]
	tokcd    := args[2]
	bmncd    := args[3]
	hatdt    := args[4]
	noudt    := args[5]
	nouydt   := args[6]
	bumcd    := args[7]
	msecd    := args[8]
	msenm    := args[9]
	kentencd := args[10]
	kentennm := args[11]
	denno    := args[12]
	denkb    := args[13]
	binkb    := args[14]
	sumsu    := strconv.Atoi(args[15])
	sumgenkn := strconv.Atoi(args[16])
	sumurikn := strconv.Atoi(args[17])
	gyono    := args[18]
	shocd    := args[19]
	jancd    := args[20]
	shonma   := args[21]
	shonmb   := args[22]
	untnm    := args[23]
	irisu    := strconv.Atoi(args[24])
	cassu    := strconv.Atoi(args[25])
	gentk    := strconv.Atoi(args[26])
	uritk    := strconv.Atoi(args[27])
	suryo    := strconv.Atoi(args[28])
	genkn    := strconv.Atoi(args[29])
	urikn    := strconv.Atoi(args[30])
	riykb    := args[31]
//	if err != nil {
//		return nil, errors.New("urikn argument must be a numeric string")
//	}
	riykb    := args[30]
//	name := args[0]
//	color := strings.ToLower(args[1])
//	user := strings.ToLower(args[3])
//	size, err := strconv.Atoi(args[2])
//	if err != nil {
//		return nil, errors.New("3rd argument must be a numeric string")
//	}

	//check if dendata already exists
	dendataAsBytes, err := stub.GetState(sysdenno)
	if err != nil {
		return nil, errors.New("Failed to get dendata sysdenno")
	}
	res := dendata{}
	json.Unmarshal(dendataAsBytes, &res)
	if res.Name == sysdenno{
		fmt.Println("This dendata arleady exists: " + sysdenno)
		fmt.Println(res);
		return nil, errors.New("This dendata arleady exists")				//all stop a dendata by this sysdenno exists
	}
	
	//build the dendata json string manually
//	str := `{"sysdenno": "` + sysdenno + `", "color": "` + color + `", "size": ` + strconv.Itoa(size) + `, "user": "` + user + `"}`
	str := `{"datsyu": "` + datsyu + `", "sysdenno": "` + sysdenno + `", "tokcd": "` + tokcd + `", "bmncd": "` + bmncd + `", "hatdt": "` + hatdt + `", 
			"noudt": "` + noudt + `", "nouydt": "` + nouydt + `", "bumcd": "` + bumcd + `", "msecd": "` + msecd + `", "msenm": "` + msenm + `", 
			"kentencd": "` + kentencd + `", "kentennm": "` + kentennm + `", "denno": "` + denno + `", "denkb": "` + denkb + `", "binkb": "` + binkb + `", 
			"sumsu": ` + strconv.Itoa(sumsu) + `, "sumgenkn": ` + strconv.Itoa(sumgenkn) + `, "sumurikn": ` + strconv.Itoa(sumurikn) + `, "gyono": "` + gyono + `",, "shocd": "` + shocd + `", 
			"jancd": "` + jancd + `", "shonma": "` + shonma + `", "shonmb": "` + shonmb + `", "untnm": "` + untnm + `", "irisu": ` + strconv.Itoa(irisu) + `, 
			"cassu": ` + strconv.Itoa(cassu) + `, "gentk": ` + strconv.Itoa(gentk) + `, "uritk": ` + strconv.Itoa(uritk) + `, "suryo": ` + strconv.Itoa(suryo) + `, 
			"genkn": ` + strconv.Itoa(genkn) + `, "urikn": ` + strconv.Itoa(urikn) + `, "riykb": "` + riykb + `"}`
	err = stub.PutState(sysdenno, []byte(str))									//store dendata with id as key
	if err != nil {
		return nil, err
	}
		
	//get the dendata index
	dendataAsBytes, err := stub.GetState(dendataIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get dendata index")
	}
	var dendataIndex []string
	json.Unmarshal(dendataAsBytes, &dendataIndex)							//un stringify it aka JSON.parse()
	
	//append
	dendataIndex = append(dendataIndex, sysdenno)									//add dendata sysdenno to index list
	fmt.Println("! dendata index: ", dendataIndex)
	jsonAsBytes, _ := json.Marshal(dendataIndex)
	err = stub.PutState(dendataIndexStr, jsonAsBytes)						//store sysdenno of dendata

	fmt.Println("- end init dendata")
	return nil, nil
}

//// ============================================================================================================================
//// Set User Permission on Dendata
//// ============================================================================================================================
//func (t *SimpleChaincode) set_user(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
//	var err error
//	
//	//   0       1
//	// "name", "bob"
//	if len(args) < 2 {
//		return nil, errors.New("Incorrect number of arguments. Expecting 2")
//	}
//	
//	fmt.Println("- start set user")
//	fmt.Println(args[0] + " - " + args[1])
//	dendataAsBytes, err := stub.GetState(args[0])
//	if err != nil {
//		return nil, errors.New("Failed to get thing")
//	}
//	res := Dendata{}
//	json.Unmarshal(dendataAsBytes, &res)										//un stringify it aka JSON.parse()
//	res.User = args[1]														//change the user
//	
//	jsonAsBytes, _ := json.Marshal(res)
//	err = stub.PutState(args[0], jsonAsBytes)								//rewrite the dendata with id as key
//	if err != nil {
//		return nil, err
//	}
//	
//	fmt.Println("- end set user")
//	return nil, nil
//}

// ============================================================================================================================
// Open Trade - create an open trade for a dendata you want with dendata you have 
// ============================================================================================================================
func (t *SimpleChaincode) open_trade(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	var will_size int
	var trade_away Description
	
	//	0        1      2     3      4      5       6
	//["bob", "blue", "16", "red", "16"] *"blue", "35*
	if len(args) < 5 {
		return nil, errors.New("Incorrect number of arguments. Expecting like 5?")
	}
	if len(args)%2 == 0{
		return nil, errors.New("Incorrect number of arguments. Expecting an odd number")
	}

	size1, err := strconv.Atoi(args[2])
	if err != nil {
		return nil, errors.New("3rd argument must be a numeric string")
	}

	open := AnOpenTrade{}
	open.User = args[0]
	open.Timestamp = makeTimestamp()											//use timestamp as an ID
	open.Want.Color = args[1]
	open.Want.Size =  size1
	fmt.Println("- start open trade")
	jsonAsBytes, _ := json.Marshal(open)
	err = stub.PutState("_debug1", jsonAsBytes)

	for i:=3; i < len(args); i++ {												//create and append each willing trade
		will_size, err = strconv.Atoi(args[i + 1])
		if err != nil {
			msg := "is not a numeric string " + args[i + 1]
			fmt.Println(msg)
			return nil, errors.New(msg)
		}
		
		trade_away = Description{}
		trade_away.Color = args[i]
		trade_away.Size =  will_size
		fmt.Println("! created trade_away: " + args[i])
		jsonAsBytes, _ = json.Marshal(trade_away)
		err = stub.PutState("_debug2", jsonAsBytes)
		
		open.Willing = append(open.Willing, trade_away)
		fmt.Println("! appended willing to open")
		i++;
	}
	
	//get the open trade struct
	tradesAsBytes, err := stub.GetState(openTradesStr)
	if err != nil {
		return nil, errors.New("Failed to get opentrades")
	}
	var trades AllTrades
	json.Unmarshal(tradesAsBytes, &trades)										//un stringify it aka JSON.parse()
	
	trades.OpenTrades = append(trades.OpenTrades, open);						//append to open trades
	fmt.Println("! appended open to trades")
	jsonAsBytes, _ = json.Marshal(trades)
	err = stub.PutState(openTradesStr, jsonAsBytes)								//rewrite open orders
	if err != nil {
		return nil, err
	}
	fmt.Println("- end open trade")
	return nil, nil
}

//// ============================================================================================================================
//// Perform Trade - close an open trade and move ownership
//// ============================================================================================================================
//func (t *SimpleChaincode) perform_trade(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
//	var err error
//	
//	//	0		1					2					3				4					5
//	//[data.id, data.closer.user, data.closer.name, data.opener.user, data.opener.color, data.opener.size]
//	if len(args) < 6 {
//		return nil, errors.New("Incorrect number of arguments. Expecting 6")
//	}
//	
//	fmt.Println("- start close trade")
//	timestamp, err := strconv.ParseInt(args[0], 10, 64)
//	if err != nil {
//		return nil, errors.New("1st argument must be a numeric string")
//	}
//	
//	size, err := strconv.Atoi(args[5])
//	if err != nil {
//		return nil, errors.New("6th argument must be a numeric string")
//	}
//	
//	//get the open trade struct
//	tradesAsBytes, err := stub.GetState(openTradesStr)
//	if err != nil {
//		return nil, errors.New("Failed to get opentrades")
//	}
//	var trades AllTrades
//	json.Unmarshal(tradesAsBytes, &trades)															//un stringify it aka JSON.parse()
//	
//	for i := range trades.OpenTrades{																//look for the trade
//		fmt.Println("looking at " + strconv.FormatInt(trades.OpenTrades[i].Timestamp, 10) + " for " + strconv.FormatInt(timestamp, 10))
//		if trades.OpenTrades[i].Timestamp == timestamp{
//			fmt.Println("found the trade");
//			
//			
//			dendataAsBytes, err := stub.GetState(args[2])
//			if err != nil {
//				return nil, errors.New("Failed to get thing")
//			}
//			closersDendata := Dendata{}
//			json.Unmarshal(dendataAsBytes, &closersDendata)											//un stringify it aka JSON.parse()
//			
//			//verify if dendata meets trade requirements
//			if closersDendata.Color != trades.OpenTrades[i].Want.Color || closersDendata.Size != trades.OpenTrades[i].Want.Size {
//				msg := "dendata in input does not meet trade requriements"
//				fmt.Println(msg)
//				return nil, errors.New(msg)
//			}
//			
//			dendata, e := findDendata4Trade(stub, trades.OpenTrades[i].User, args[4], size)			//find a dendata that is suitable from opener
//			if(e == nil){
//				fmt.Println("! no errors, proceeding")
//
//				t.set_user(stub, []string{args[2], trades.OpenTrades[i].User})						//change owner of selected dendata, closer -> opener
//				t.set_user(stub, []string{dendata.Name, args[1]})									//change owner of selected dendata, opener -> closer
//			
//				trades.OpenTrades = append(trades.OpenTrades[:i], trades.OpenTrades[i+1:]...)		//remove trade
//				jsonAsBytes, _ := json.Marshal(trades)
//				err = stub.PutState(openTradesStr, jsonAsBytes)										//rewrite open orders
//				if err != nil {
//					return nil, err
//				}
//			}
//		}
//	}
//	fmt.Println("- end close trade")
//	return nil, nil
//}

//// ============================================================================================================================
//// findDendata4Trade - look for a matching dendata that this user owns and return it
//// ============================================================================================================================
//func findDendata4Trade(stub shim.ChaincodeStubInterface, user string, color string, size int )(m Dendata, err error){
//	var fail Dendata;
//	fmt.Println("- start find dendata 4 trade")
//	fmt.Println("looking for " + datsyu + ", " + sysdenno);
//
//	//get the dendata index
//	dendataAsBytes, err := stub.GetState(dendataIndexStr)
//	if err != nil {
//		return fail, errors.New("Failed to get dendata index")
//	}
//	var dendataIndex []string
//	json.Unmarshal(dendataAsBytes, &dendataIndex)								//un stringify it aka JSON.parse()
//	
//	for i:= range dendataIndex{													//iter through all the dendata
//		//fmt.Println("looking @ dendata name: " + dendataIndex[i]);
//
//		dendataAsBytes, err := stub.GetState(dendataIndex[i])						//grab this dendata
//		if err != nil {
//			return fail, errors.New("Failed to get dendata")
//		}
//		res := Dendata{}
//		json.Unmarshal(dendataAsBytes, &res)										//un stringify it aka JSON.parse()
//		//fmt.Println("looking @ " + res.User + ", " + res.Color + ", " + strconv.Itoa(res.Size));
//		
//		//check for user && color && size
//		if strings.ToLower(res.User) == strings.ToLower(user) && strings.ToLower(res.Color) == strings.ToLower(color) && res.Size == size{
//			fmt.Println("found a dendata: " + res.Name)
//			fmt.Println("! end find dendata 4 trade")
//			return res, nil
//		}
//	}
//	
//	fmt.Println("- end find dendata 4 trade - error")
//	return fail, errors.New("Did not find dendata to use in this trade")
//}

// ============================================================================================================================
// Make Timestamp - create a timestamp in ms
// ============================================================================================================================
func makeTimestamp() int64 {
    return time.Now().UnixNano() / (int64(time.Millisecond)/int64(time.Nanosecond))
}

// ============================================================================================================================
// Remove Open Trade - close an open trade
// ============================================================================================================================
func (t *SimpleChaincode) remove_trade(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	
	//	0
	//[data.id]
	if len(args) < 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}
	
	fmt.Println("- start remove trade")
	timestamp, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return nil, errors.New("1st argument must be a numeric string")
	}
	
	//get the open trade struct
	tradesAsBytes, err := stub.GetState(openTradesStr)
	if err != nil {
		return nil, errors.New("Failed to get opentrades")
	}
	var trades AllTrades
	json.Unmarshal(tradesAsBytes, &trades)																//un stringify it aka JSON.parse()
	
	for i := range trades.OpenTrades{																	//look for the trade
		//fmt.Println("looking at " + strconv.FormatInt(trades.OpenTrades[i].Timestamp, 10) + " for " + strconv.FormatInt(timestamp, 10))
		if trades.OpenTrades[i].Timestamp == timestamp{
			fmt.Println("found the trade");
			trades.OpenTrades = append(trades.OpenTrades[:i], trades.OpenTrades[i+1:]...)				//remove this trade
			jsonAsBytes, _ := json.Marshal(trades)
			err = stub.PutState(openTradesStr, jsonAsBytes)												//rewrite open orders
			if err != nil {
				return nil, err
			}
			break
		}
	}
	
	fmt.Println("- end remove trade")
	return nil, nil
}

// ============================================================================================================================
// Clean Up Open Trades - make sure open trades are still possible, remove choices that are no longer possible, remove trades that have no valid choices
// ============================================================================================================================
func cleanTrades(stub shim.ChaincodeStubInterface)(err error){
	var didWork = false
	fmt.Println("- start clean trades")
	
	//get the open trade struct
	tradesAsBytes, err := stub.GetState(openTradesStr)
	if err != nil {
		return errors.New("Failed to get opentrades")
	}
	var trades AllTrades
	json.Unmarshal(tradesAsBytes, &trades)																		//un stringify it aka JSON.parse()
	
	fmt.Println("# trades " + strconv.Itoa(len(trades.OpenTrades)))
	for i:=0; i<len(trades.OpenTrades); {																		//iter over all the known open trades
		fmt.Println(strconv.Itoa(i) + ": looking at trade " + strconv.FormatInt(trades.OpenTrades[i].Timestamp, 10))
		
		fmt.Println("# options " + strconv.Itoa(len(trades.OpenTrades[i].Willing)))
		for x:=0; x<len(trades.OpenTrades[i].Willing); {														//find a dendata that is suitable
			fmt.Println("! on next option " + strconv.Itoa(i) + ":" + strconv.Itoa(x))
			_, e := findDendata4Trade(stub, trades.OpenTrades[i].User, trades.OpenTrades[i].Willing[x].Color, trades.OpenTrades[i].Willing[x].Size)
			if(e != nil){
				fmt.Println("! errors with this option, removing option")
				didWork = true
				trades.OpenTrades[i].Willing = append(trades.OpenTrades[i].Willing[:x], trades.OpenTrades[i].Willing[x+1:]...)	//remove this option
				x--;
			}else{
				fmt.Println("! this option is fine")
			}
			
			x++
			fmt.Println("! x:" + strconv.Itoa(x))
			if x >= len(trades.OpenTrades[i].Willing) {														//things might have shifted, recalcuate
				break
			}
		}
		
		if len(trades.OpenTrades[i].Willing) == 0 {
			fmt.Println("! no more options for this trade, removing trade")
			didWork = true
			trades.OpenTrades = append(trades.OpenTrades[:i], trades.OpenTrades[i+1:]...)					//remove this trade
			i--;
		}
		
		i++
		fmt.Println("! i:" + strconv.Itoa(i))
		if i >= len(trades.OpenTrades) {																	//things might have shifted, recalcuate
			break
		}
	}

	if(didWork){
		fmt.Println("! saving open trade changes")
		jsonAsBytes, _ := json.Marshal(trades)
		err = stub.PutState(openTradesStr, jsonAsBytes)														//rewrite open orders
		if err != nil {
			return err
		}
	}else{
		fmt.Println("! all open trades are fine")
	}

	fmt.Println("- end clean trades")
	return nil
}