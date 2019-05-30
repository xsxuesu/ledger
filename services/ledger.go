package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"ledger/common"
	"ledger/log"
	"ledger/model"
	"strconv"
	"strings"
	"time"
)

/*token 發行*/
func LedgerIssue(stub shim.ChaincodeStubInterface)pb.Response{

	_,args := stub.GetFunctionAndParameters()

	if len(args) != 1{
		return shim.Error("Parameters error ,please check Parameters")
	}

	issueJson := args[0]
	log.Logger.Info(issueJson)
    issueParam   :=	model.LedgerIssueParam{}
    err := json.Unmarshal([]byte(issueJson),&issueParam)
    if err != nil {
		log.Logger.Error("Unmarshal:",err)
    	return shim.Error(err.Error())
	}

    token, err := TokenGet(stub,issueParam.Token)
	if err != nil {
		log.Logger.Error("TokenGet:",err)
		return shim.Error(err.Error())
	}

    curUserName,err  := common.GetCommonName(stub)
	if err != nil {
		log.Logger.Error("GetCommonName:",err)
		return shim.Error(err.Error())
	}
    //// super admin
    if ! common.IsSuperAdmin(stub) {
		/// 發行人
		if strings.ToUpper(token.Issuer) != strings.ToUpper(curUserName) {
			return common.SendError(common.Right_ERR,"only super admin and token issuer can issue the token")
		}
	}

	if token.Status == false {
		return common.SendError(common.TKNERR_LOCKED,fmt.Sprintf("%s token not enable",token.Name))
	}

	holder, err  := common.GetCommonName(stub)
	if err != nil {
		log.Logger.Error("GetCommonName:",err)
		return shim.Error(err.Error())
	}
	if common.IsSuperAdmin(stub){
		if issueParam.Holder == "" {
			return  common.SendError(common.Param_ERR,"the token holder not allownce empty")
		}

		accout, err := AccountGetByName(stub,issueParam.Holder)
		if err != nil {
			log.Logger.Error("AccountGetByName:",err)
			return shim.Error(err.Error())
		}

		if accout.Status == false {
			return  common.SendError(common.ACCOUNT_NOT_EXIST,"the holder is not exist or the holder is disable")
		}
		holder = accout.DidName
	}


	log.Logger.Info("holder:",holder)

	leder := model.Ledger{}

	key, err := stub.CreateCompositeKey(common.CompositeIndexName, []string{common.Ledger_PRE, strings.ToUpper(token.Name),  strings.ToUpper(holder)})
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not create a composite key for %s-%s: %s", token.Name, holder, err.Error()))
	}

	leder.Holder = holder
	leder.Token = token.Name
	leder.Amount = issueParam.Amount
	leder.Desc = fmt.Sprintf("%s issue %s token amount:%s",curUserName,token.Name,strconv.FormatFloat(issueParam.Amount,'f',2,64))

	ledgerByte, err := json.Marshal(leder)
	if err != nil {
		log.Logger.Error("Marshal:",err)
		return shim.Error(err.Error())
	}

	err = stub.PutState(key,ledgerByte)

	if err != nil {
		log.Logger.Error("PutState:",err)
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}
// get balance
func LedgerGetBalance(stub shim.ChaincodeStubInterface)pb.Response  {

	_,args := stub.GetFunctionAndParameters()

	if len(args) != 1{
		return shim.Error("Parameters error ,please check Parameters")
	}

	balancejson := args[0]
	log.Logger.Info(balancejson)
	balance := model.LedgerBalanceParam{}

	err  := json.Unmarshal([]byte(balancejson),&balance)
	if err != nil {
		return shim.Error(err.Error())
	}
	/////////////////////////////// ==============account accountFrom, err := AccountGetByName(stub,transfer.From)
	account,err := AccountGetByName(stub,balance.Holder)
	if err != nil {
		return shim.Error(err.Error())
	}
	////////////////////////////// =============token
	token , err := TokenGet(stub,balance.Token)
	if err != nil {
		log.Logger.Error("TokenGet:",err)
		return shim.Error(err.Error())
	}
	/////////////////////////////// =================== token and account check
	if account.Status == false {
		return  common.SendError(common.ACCOUNT_NOT_EXIST,"the holder is not exist or the holder is disable")
	}
	if token.Status == false {
		return common.SendError(common.TKNERR_LOCKED,fmt.Sprintf("%s token not enable",token.Name))
	}

	key, err := stub.CreateCompositeKey(common.CompositeIndexName, []string{common.Ledger_PRE, strings.ToUpper(token.Name),  strings.ToUpper(account.DidName)})
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not create a composite key for %s-%s: %s", token.Name, account.DidName, err.Error()))
	}

	ledgerByte,err := stub.GetState(key)

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(ledgerByte)
}
// get history
func LedgerGetHistory(stub shim.ChaincodeStubInterface)pb.Response{

	_,args := stub.GetFunctionAndParameters()

	if len(args) != 1{
		return shim.Error("Parameters error ,please check Parameters")
	}

	balancejson := args[0]

	balance := model.LedgerBalanceParam{}

	err  := json.Unmarshal([]byte(balancejson),&balance)

	if err != nil {
		return shim.Error(err.Error())
	}

	account, err := AccountGetByName(stub,balance.Holder)
	if err != nil {
		return shim.Error(err.Error())
	}
	token , err := TokenGet(stub,balance.Token)
	if err != nil {
		return shim.Error(err.Error())
	}
	if account.Status == false {
		return  common.SendError(common.ACCOUNT_NOT_EXIST,"the holder is not exist or the holder is disable")
	}
	if token.Status == false {
		return common.SendError(common.TKNERR_LOCKED,fmt.Sprintf("%s token not enable",token.Name))
	}

	key, err := stub.CreateCompositeKey(common.CompositeIndexName, []string{common.Ledger_PRE, strings.ToUpper(token.Name),  strings.ToUpper(account.DidName)})
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not create a composite key for %s-%s: %s", token.Name, account.DidName, err.Error()))
	}

	history, err := stub.GetHistoryForKey(key)

	if err != nil {
		return shim.Error(err.Error())
	}

	defer  history.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false

	for history.HasNext(){
		response ,err := history.Next()
		if err != nil {
			continue
		}
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}
		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	return shim.Success(buffer.Bytes())
}

func LedgerTransfer(stub shim.ChaincodeStubInterface)pb.Response{
	_,args := stub.GetFunctionAndParameters()

	if len(args) != 1{
		return shim.Error("Parameters error ,please check Parameters")
	}

	transferjson := args[0]
	log.Logger.Info(transferjson)
	transfer := model.LedgerTransferParam{}
	err := json.Unmarshal([]byte(transferjson),&transfer)
	if err != nil {
		return shim.Error(err.Error())
	}

	curUserName,err  := common.GetCommonName(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	if strings.ToUpper(strings.TrimSpace(curUserName)) != strings.ToUpper(strings.TrimSpace(transfer.From)){
		return common.SendError(common.ACCOUNT_PREMISSION,fmt.Sprintf("%s is not current login user :%s",transfer.From,curUserName))
	}

	accountFrom, err := AccountGetByName(stub,transfer.From)
	if err != nil {
		return shim.Error(err.Error())
	}
	accountTo, err := AccountGetByName(stub,transfer.To)
	if err != nil {
		return shim.Error(err.Error())
	}

	if accountFrom.DidName == accountTo.DidName {
		return common.SendError(common.ACCOUNT_PREMISSION,fmt.Sprintf("from : %s can not equal to user :%s",transfer.From,transfer.To))
	}

	if accountFrom.Status == false || accountTo.Status == false {
		return common.SendError(common.ACCOUNT_LOCK,fmt.Sprintf("from : %s OR to  :%s is locked",transfer.From,transfer.To))
	}

	token , err := TokenGet(stub,transfer.Token)
	if err != nil {
		log.Logger.Error("TokenGet:",err)
		return shim.Error(err.Error())
	}

	if token.Status == false {
		return common.SendError(common.TKNERR_LOCKED,fmt.Sprintf("%s Token is disable",token.Name))
	}

	//////////// from
	key, err := stub.CreateCompositeKey(common.CompositeIndexName, []string{common.Ledger_PRE, strings.ToUpper(token.Name),  strings.ToUpper(accountFrom.DidName)})
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not create a composite key for %s-%s: %s", token.Name, accountFrom.DidName, err.Error()))
	}

	ledgerByte,err := stub.GetState(key)
	if err != nil {
		log.Logger.Error("GetState:",err)
		return shim.Error(err.Error())
	}


	///////////////////////////from
	ledger := model.Ledger{}

	err = json.Unmarshal(ledgerByte,&ledger)
	if err != nil {
		log.Logger.Error("Unmarshal:",err)
		return shim.Error(err.Error())
	}
	if ledger.Amount < transfer.Amount {
		return common.SendError(common.Balance_NOT_ENOUGH,fmt.Sprintf("the %s token balance not enough",token.Name))
	}
	ledger.Amount = ledger.Amount - transfer.Amount
	ledger.Desc = fmt.Sprintf("From : %s transfer To : %s , value : %s ",accountFrom.DidName,accountTo.DidName,strconv.FormatFloat(transfer.Amount,'f',2,64))

	ledgerByted , err := json.Marshal(ledger)
	if err != nil {
		log.Logger.Error("Marshal:",err)
		return shim.Error(err.Error())
	}

	err = stub.PutState(key,ledgerByted)
	if err != nil {
		log.Logger.Error("PutState:",err)
		return shim.Error(err.Error())
	}

	//////////////////////to
	tokey, err := stub.CreateCompositeKey(common.CompositeIndexName, []string{common.Ledger_PRE, strings.ToUpper(token.Name),  strings.ToUpper(accountTo.DidName)})
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not create a composite key for %s-%s: %s", token.Name, accountTo.DidName, err.Error()))
	}
	toledgerByte,err := stub.GetState(tokey)
	if err != nil{
		log.Logger.Error("TO GetState:",err)
		return shim.Error(err.Error())
	}
	log.Logger.Info("toledgerByte:",toledgerByte)

	toledger := model.Ledger{}
	if toledgerByte == nil {
		toledger.Holder = strings.ToUpper(accountTo.DidName)
		toledger.Token = strings.ToUpper(token.Name)
		toledger.Desc = fmt.Sprintf("From : %s transfer To : %s , value : %s ",accountFrom.DidName,accountTo.DidName,strconv.FormatFloat(transfer.Amount,'f',2,64))
		toledger.Amount = toledger.Amount + transfer.Amount
	}else {
		err = json.Unmarshal(toledgerByte,&toledger)
		if err != nil {
			log.Logger.Error("Unmarshal22:",err)
			return shim.Error(err.Error())
		}
		toledger.Desc = fmt.Sprintf("From : %s transfer To : %s , value : %s ",accountFrom.DidName,accountTo.DidName,strconv.FormatFloat(transfer.Amount,'f',2,64))
		toledger.Amount = toledger.Amount + transfer.Amount
	}
	toTransferedByted , err := json.Marshal(toledger)
	if err != nil {
		log.Logger.Error("Marshal333:",err)
		return shim.Error(err.Error())
	}
	err = stub.PutState(tokey,toTransferedByted)
	if err != nil {
		log.Logger.Error("PutState222:",err)
		return shim.Error(err.Error())
	}

	////////////////// send event
	ts, err := stub.GetTxTimestamp()
	if err != nil {
		return shim.Error(err.Error())
	}
	//////// set event
	evt := model.LedgerEvent{
		Type: common.Evt_payment,
		Txid:   stub.GetTxID(),
		Time:   ts.GetSeconds(),
		From:   transfer.From,
		To:     transfer.To,
		Amount: strconv.FormatFloat(transfer.Amount,'f',2,64) ,
		Token:  transfer.Token,
	}

	eventJSONasBytes, err := json.Marshal(evt)
	if err != nil {
		log.Logger.Error("Marshal2222:",err)
		return shim.Error(err.Error())
	}

	err = stub.SetEvent(fmt.Sprintf(common.TOPIC, transfer.To), eventJSONasBytes)
	if err != nil {
		log.Logger.Error("SetEvent:",err)
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func LedgerScale(stub shim.ChaincodeStubInterface)pb.Response  {
	_,args := stub.GetFunctionAndParameters()

	if len(args) != 1{
		return shim.Error("Parameters error ,please check Parameters")
	}

	scaleJson := args[0]
	log.Logger.Info(scaleJson)
	scaleParam := model.LedgerScaleParam{}

	err := json.Unmarshal([]byte(scaleJson),&scaleParam)

	if err != nil {
		return shim.Error(err.Error())
	}

	token, err := TokenGet(stub,scaleParam.Token)

	if err != nil {
		return shim.Error(err.Error())
	}

	if token.Status == false {
		return common.SendError(common.TKNERR_LOCKED,fmt.Sprintf("%s Token is disable",token.Name))
	}

	if common.IsSuperAdmin(stub) == false {
		return common.SendError(common.ACCOUNT_PREMISSION,"only super admin can break up or merge operation")
	}
	///////////////////////////////////////////////////////================================修改持有token賬戶
	resultIterator, err := stub.GetStateByPartialCompositeKey(common.CompositeIndexName, []string{common.Ledger_PRE,strings.ToUpper(token.Name)})
	if err != nil {
		log.Logger.Error("GetStateByPartialCompositeKey:",err)
		return shim.Error(err.Error())
	}
	defer resultIterator.Close()

	var i int
	for i=0; resultIterator.HasNext();i++ {
		iterObj,err := resultIterator.Next()
		if err != nil {
			log.Logger.Error("resultIterator:",err)
			return shim.Error(err.Error())
		}
		key := iterObj.Key

		log.Logger.Info(key)

		ledger := model.Ledger{}
		err = json.Unmarshal(iterObj.Value,&ledger)
		if err != nil {
			log.Logger.Error("Unmarshal:",err)
			return shim.Error(err.Error())
		}
		log.Logger.Info("amount compute")
		log.Logger.Info("amount pre :",ledger.Amount)
		ledger.Amount = ledger.Amount * scaleParam.Scale
		log.Logger.Info("amount aft:",ledger.Amount)
		if scaleParam.Scale > float64(1.0) {
			ledger.Desc =  fmt.Sprintf("%s token break up , breake up scale %s",scaleParam.Token, strconv.FormatFloat(scaleParam.Scale,'f',2,64) )
		}else{
			ledger.Desc =  fmt.Sprintf("%s token merge , merge scale %s",scaleParam.Token, strconv.FormatFloat(scaleParam.Scale,'f',2,64) )
		}

		ledgerByte,err  := json.Marshal(ledger)
		log.Logger.Info("Marshal:")
		if err != nil {
			log.Logger.Error("Marshal:",err)
			return shim.Error(err.Error())
		}
		err = stub.PutState(key,ledgerByte)
		log.Logger.Info("putstate end")
		if err != nil {
			log.Logger.Error("PutState:",err)
			return shim.Error(err.Error())
		}
	}

	///////////////////////////////////////////////////////===============================修改未簽名的轉賬

	signIter , err := stub.GetStateByPartialCompositeKey(common.CompositeRequestIndexName,[]string{common.SIGN_PRE,strings.ToUpper(token.Name)})
	if err != nil {
		log.Logger.Error("GetStateByPartialCompositeKey:",err)
		return shim.Error(err.Error())
	}
	defer signIter.Close()
	var j int
	for j=0;signIter.HasNext();j++{
		iterSignObj,err := signIter.Next()
		if err != nil {
			log.Logger.Error("signIter:",err)
			return shim.Error(err.Error())
		}
		signkey := iterSignObj.Key

		log.Logger.Info(signkey)

		signReq := model.SignRequest{}
		err = json.Unmarshal(iterSignObj.Value,&signReq)
		if err != nil {
			log.Logger.Error("Unmarshal:",err)
			return shim.Error(err.Error())
		}
		if signReq.Status == common.PENDING_SIGN {
			log.Logger.Info("amount pre :",signReq.Amount)
			signReq.Amount = signReq.Amount * scaleParam.Scale
			log.Logger.Info("amount aft:",signReq.Amount)
			////////// desc info
			if scaleParam.Scale > float64(1.0) {
				signReq.Desc =  fmt.Sprintf("%s token break up , breake up scale %s",scaleParam.Token, strconv.FormatFloat(scaleParam.Scale,'f',2,64) )
			}else{
				signReq.Desc =  fmt.Sprintf("%s token merge , merge scale %s",scaleParam.Token, strconv.FormatFloat(scaleParam.Scale,'f',2,64) )
			}
			ledgerByte,err  := json.Marshal(signReq)

			if err != nil {
				log.Logger.Error("Marshal:",err)
				return shim.Error(err.Error())
			}
			err = stub.PutState(signkey,ledgerByte)

			if err != nil {
				log.Logger.Error("PutState:",err)
				return shim.Error(err.Error())
			}
		}else{
			continue
		}
	}


	returnString := fmt.Sprintf("had scale %d token holders, had scale %d pending for sign tx",i,j)
	return shim.Success([]byte(returnString))

}
