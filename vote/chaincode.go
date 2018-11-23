package main

import (
	"encoding/json"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"fmt"
	"strings"
	"strconv"
	"bytes"
	"crypto/md5"
)

type VoteChaincode struct{}

//
type People struct {
	Name string `json:"name"`
	Role string `json:"role"`
	Vote string `json:"vote"`
}

//参选人的集合，放的是参选人的公钥集合
//todo 若是同一时间两个参选人注册，fabric如何保证数据一致性
type Candidates struct {
	CandidatesSigncert []string
}
type Peoples struct {
	PeoplesSigncert []string
}

//用于存放选举的几个不同的状态
const (
	StartState    = "start"
	CandidateRole = "candidate"
	VoterRole     = "voter"
	CandidatesKey = "candidates"
	PeoplesKey    = "peoples"
	Founder       = "founder"
	BalanceState  = "balance"
	VoteState     = "vote"
	State         = "state"
)

func (f *VoteChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	creatorByte, _ := stub.GetCreator()
	fmt.Println(string(creatorByte))
	stub.PutState(Founder, creatorByte)
	stub.PutState(State, []byte(StartState))
	candidates, _ := json.Marshal(new(Candidates))
	fmt.Println(string(candidates))
	peoples, _ := json.Marshal(new(Peoples))
	fmt.Println(string(peoples))
	stub.PutState(CandidatesKey, []byte(candidates))
	stub.PutState(PeoplesKey, []byte(peoples))
	fmt.Println("初始化成功")
	return successReslut(nil)
}

func (f *VoteChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	//todo 选举状态的合法性判断
	fmt.Println("Invoke")
	function, args := stub.GetFunctionAndParameters()
	if len(args) == 0 {
		fmt.Println("没有参数传入")
	} else {
		fmt.Println("arg的长度为:", len(args))
		fmt.Println(strings.Join(args, ","))
	}
	switch function {
	case "register":
		if !checkState(stub, StartState) {
			return failReslut("不在此阶段")
		}
		return f.register(stub, args)
	case "getState":
		return f.getState(stub)
	case "changeState":
		return f.changeState(stub, args)
	case "getCandidates":
		return f.getCandidates(stub)
	case "get":
		return f.get(stub)
	case "getPeoples":
		//todo 未来去掉这个方法，用于测试用
		return f.getPeoples(stub)
	case "vote":
		if !checkState(stub, VoteState) {
			return failReslut("不在此阶段")
		}
		return f.vote(stub, args)
	case "balance":
		if !checkState(stub, BalanceState) {
			return failReslut("不在此阶段")
		}
		return f.balance(stub)
	case "getPeople":
		return f.getPeople(stub, args)
	}

	return failReslut("参数异常")
}
func (f *VoteChaincode)get(stub shim.ChaincodeStubInterface)pb.Response{
	creatorByte, _ := stub.GetCreator()
	result,_:=stub.GetState(handleByteMd5(creatorByte))
	return successReslut(result)
}
func (f *VoteChaincode) getPeoples(stub shim.ChaincodeStubInterface) pb.Response {
	//取出所有投票者
	peoples := new(Peoples)
	peoplesByte, _ := stub.GetState(PeoplesKey)
	json.Unmarshal(peoplesByte, peoples)
	peopleMap := make(map[string]People)
	for _, pS := range peoples.PeoplesSigncert {
		pB, _ := stub.GetState(pS)
		p := new(People)
		json.Unmarshal(pB, p)
		peopleMap[pS] = *p
	}
	result, _ := json.Marshal(peopleMap)
	return successReslut(result)
}
func handleByteMd5(input []byte) string{
	return fmt.Sprintf("%x", md5.Sum(input))
}
func checkState(stub shim.ChaincodeStubInterface, state string) bool {
	stateByte, _ := stub.GetState(State)
	if state != string(stateByte) {
		return false
	}
	return true
}
func (f *VoteChaincode) getCandidates(stub shim.ChaincodeStubInterface) pb.Response {
	candidatesByte, err := stub.GetState(CandidatesKey)
	if err != nil {
		return failReslut("从区块取出数据异常")
	}
	candidates := new(Candidates)
	json.Unmarshal(candidatesByte, candidates)
	candidateMap := make(map[string]People)
	//取出所有参选人
	for _, c := range candidates.CandidatesSigncert {
		cB, _ := stub.GetState(c)
		p := new(People)
		json.Unmarshal(cB, p)
		//todo 如何把字符串转为指针
		p.Vote =  "0"
		candidateMap[c] = *p
	}
	result, _ := json.Marshal(candidateMap)
	return successReslut(result)
}
func (f *VoteChaincode) balance(stub shim.ChaincodeStubInterface) pb.Response {
	candidatesByte, err := stub.GetState(CandidatesKey)
	if err != nil {
		return failReslut("从区块取出数据异常")
	}
	candidates := new(Candidates)
	json.Unmarshal(candidatesByte, candidates)
	candidateMap := make(map[string]*People)
	//取出所有参选人
	for _, c := range candidates.CandidatesSigncert {
		cB, _ := stub.GetState(c)
		p := new(People)
		json.Unmarshal(cB, p)
		//todo 如何把字符串转为指针
		p.Vote = "0"
		candidateMap[c] = p
	}
	//取出所有投票者
	peoples := new(Peoples)
	peoplesByte, _ := stub.GetState(PeoplesKey)
	json.Unmarshal(peoplesByte, peoples)
	for _, pS := range peoples.PeoplesSigncert {
		pB, _ := stub.GetState(pS)
		p := new(People)
		json.Unmarshal(pB, p)
		if len(p.Vote) != 0 {
			//todo 或许可以优化
			//if _,ok:=candidateMap[p.Vote];ok{
			//	v, _ := strconv.Atoi(candidateMap[p.Vote].Vote)
			if candidate,ok:=candidateMap[p.Vote];ok{
				v, _ := strconv.Atoi(candidate.Vote)
				v++
				candidateMap[p.Vote].Vote = strconv.Itoa(v)
			}
		}
	}
	result, _ := json.Marshal(candidateMap)
	return successReslut(result)
}
func (f *VoteChaincode) vote(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	userSigncertByte, _ := stub.GetCreator()
	userKey:=handleByteMd5(userSigncertByte)
	user, err := stub.GetState(userKey)
	if err != nil || user == nil {
		return failReslut("从区块中取出投票者信息异常")
	}
	people := new(People)
	json.Unmarshal(user, people)
	//参选人没有投票权利 | 没票的人没权利
	if len(people.Vote) != 0 || people.Role == CandidateRole {
		return failReslut("参选人异常")
	}
	//找到被选举者
	//todo 可以有废票吗？
	candidate, err := stub.GetState(args[0])
	if err != nil || candidate == nil {
		return failReslut("从区块中取出选举人信息异常")
	}
	people.Vote = args[0]
	//存入投票表信息
	peopleByte, _ := json.Marshal(people)
	if err := stub.PutState(userKey, peopleByte); err != nil {
		return failReslut("写入区块异常")
	}
	return successReslut([]byte("投票成功"))
}
func (f *VoteChaincode) changeState(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	userSigncertByte, _ := stub.GetCreator()
	creatorByte, _ := stub.GetState(Founder)
	if !bytes.Equal(userSigncertByte, creatorByte) {
		return failReslut("没有权限")
	}
	stub.PutState(State, []byte(args[0]))
	return successReslut([]byte("success"))
}
func (f *VoteChaincode) getState(stub shim.ChaincodeStubInterface) pb.Response {
	result, err := stub.GetState(State)
	if err != nil {
		failReslut("读取区块异常")
	}
	return successReslut(result)
}
func successReslut(res []byte) pb.Response {
	fmt.Println("执行成功:")
	fmt.Println(string(res))
	return shim.Success(res)
}
func failReslut(msg string) pb.Response {
	fmt.Println("出错啦:")
	fmt.Println(msg)
	return shim.Error(msg)
}
func (f *VoteChaincode) getPeople(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	user, _ := stub.GetState(args[0])
	return successReslut(user)
}
func (f *VoteChaincode) register(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//把传入的数组转为字符串，转成对象
	fmt.Println("注册用户")
	userSigncertByte, _ := stub.GetCreator()
	userKey:=handleByteMd5(userSigncertByte)
	user, err := stub.GetState(userKey)
	if err != nil {
		return failReslut("取数据时发生异常")
	} else {
		if user != nil {
			return failReslut("用户已注册")
		}
	}
	people := new(People)
	err = json.Unmarshal([]byte(args[0]), people)
	if err != nil {
		return failReslut("JSON转换异常")
	}
	if people.Name == "" {
		return failReslut("参数异常")
	}
	if people.Role == CandidateRole {
		//从区块中取出所有参选者
		value, _ := stub.GetState(CandidatesKey)
		candidatesInBlock := new(Candidates)
		json.Unmarshal(value, candidatesInBlock)

		candidatesInBlock.CandidatesSigncert = append(candidatesInBlock.CandidatesSigncert, userKey)
		byte, _ := json.Marshal(candidatesInBlock)
		//将此参选者加入区块中
		stub.PutState(CandidatesKey, byte)
	} else if people.Role == VoterRole {
		//从区块中取出所有投票者
		value, _ := stub.GetState(PeoplesKey)
		peoplesInBlock := new(Peoples)
		json.Unmarshal(value, peoplesInBlock)
		peoplesInBlock.PeoplesSigncert = append(peoplesInBlock.PeoplesSigncert, userKey)
		byte, _ := json.Marshal(peoplesInBlock)
		//将此投票者加入区块中
		stub.PutState(PeoplesKey, byte)
	} else {
		return failReslut("角色异常")
	}
	v, _ := json.Marshal(people)
	fmt.Println("注册的用户key为:", userKey)
	fmt.Println("注册的用户为:", string(v))
	stub.PutState(userKey, v)
	//res, err := stub.GetState(userKey)
	//if err != nil {
	//	return failReslut("拿出数据出错啦")
	//} else if res == nil {
	//	return failReslut("拿出数据为空")
	//}
	return successReslut([]byte("success"))
}

func main() {
	err := shim.Start(new(VoteChaincode))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}

