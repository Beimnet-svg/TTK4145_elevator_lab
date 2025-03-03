package ordermanager

import (
	config "Project-go/Config"
	masterslavedist "Project-go/MasterSlaveDist"
	requests "Project-go/driver-go/Requests"
	"Project-go/driver-go/elevio"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
)

var (
	AllActiveOrders [config.NumberElev][config.NumberFloors][config.NumberBtn]bool
	NewRequests     [config.NumberElev][config.NumberFloors][config.NumberBtn]bool
	orderCounter    [config.NumberElev]int
	ElevState       [config.NumberElev]elevio.Elevator
)

var motorDirectionToString = map[elevio.MotorDirection]string{
	elevio.MD_Up:   "up",
	elevio.MD_Down: "down",
	elevio.MD_Stop: "stop",
}

var behaviorToString = map[elevio.ElevatorBehaviour]string{
	elevio.EB_Idle:     "idle",
	elevio.EB_Moving:   "moving",
	elevio.EB_DoorOpen: "doorOpen",
}

type HRAElevState struct {
	Behavior    string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [][2]bool               `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
}

func UpdateOrders(e elevio.Elevator, receiver chan [config.NumberElev][config.NumberFloors][config.NumberBtn]bool) {
	//Update elevator states
	ElevState[e.ElevatorID] = e
	
	//Clear orders at current floor based on elevator state
	AllActiveOrders = requests.RequestClearAtCurrentFloor(e, AllActiveOrders)

	//Create a maxCounterValue, where every value in e.requests higher than this 
	//value is considered a new order
	maxCounterValue := orderCounter[e.ElevatorID]

	//If we have a new order we redistribute hall orders and set new order counter
	if  CheckIfNewOrders(e, &maxCounterValue) {
		//Fetch active elevators from master-slave module
		elevators := masterslavedist.FetchAliveElevators(ElevState)
		orderCounter[e.ElevatorID] = maxCounterValue
		input := formatInput(elevators, AllActiveOrders, NewRequests)
		AllActiveOrders = assignRequests(input)
	}

	//Send updated orders to all elevators
	receiver <- AllActiveOrders
}

func CheckIfNewOrders(e elevio.Elevator, maxCounterValue *int) bool{
	//Check if there are new orders in the system
	
	for i := 0; i < e.NumFloors; i++ {
		for j := 0; j < 3; j++ {
			//Based on the counter values in e.Requests we can determine if we have a new order
			if e.Requests[i][j] > orderCounter[e.ElevatorID] {
				NewRequests[i][j][e.ElevatorID] = true
				if e.Requests[i][j] > *maxCounterValue {
					//Find the highest counter value in the elevator
					*maxCounterValue = e.Requests[i][j]
				}
			}
		}
	}

	return *maxCounterValue > orderCounter[e.ElevatorID]
}


func formatInput(elevators []elevio.Elevator, allActiveOrders [config.NumberElev][config.NumberFloors][config.NumberBtn]bool,
	newRequests [config.NumberElev][config.NumberFloors][config.NumberBtn]bool) HRAInput {
	hallRequests := [][2]bool{}
	cabRequests := [config.NumberElev][]bool{}

	for i := 0; i < config.NumberElev; i++ {
		for j := 0; j < config.NumberFloors; j++ {
			for k := 0; k < 2; k++ {
				//Extract hallrequests from current and new orders
				hallRequests[j][k] = hallRequests[j][k] || allActiveOrders[i][j][k] || newRequests[i][j][k]
			}
		}
		for j := 0; j < config.NumberFloors; j++ {
			//Extract cabrequests from current and new orders
			cabRequests[i][j] = allActiveOrders[i][j][2] || newRequests[i][j][2]
		}

	}

	input := HRAInput{
		HallRequests: hallRequests,
		States:       map[string]HRAElevState{},
	}
	//Add all active elevator states to cost func input
	for _, e := range elevators {
		input.States[strconv.Itoa(e.ElevatorID)] = HRAElevState{
			Behavior:    behaviorToString[e.Behaviour],
			Floor:       e.CurrentFloor,
			Direction:   motorDirectionToString[e.Direction],
			CabRequests: cabRequests[e.ElevatorID][:],
		}
	}
	return input
}

func assignRequests(input HRAInput) [config.NumberElev][config.NumberFloors][config.NumberBtn]bool {

	hraExecutable := ""
	switch runtime.GOOS {
	case "linux":
		hraExecutable = "hall_request_assigner"
	case "windows":
		hraExecutable = "hall_request_assigner.exe"
	default:
		panic("OS not supported")
	}

	jsonBytes, err := json.Marshal(input)
	if err != nil {
		fmt.Println("json.Marshal error: ", err)
	}

	ret, err := exec.Command(hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Println("exec.Command error: ", err)
		fmt.Println(string(ret))
	}

	return transformOutput(ret, input)

}

func transformOutput(ret []byte, input HRAInput) [config.NumberElev][config.NumberFloors][config.NumberBtn]bool {

	tempOutput := new(map[string][][2]bool)
	newAllActiveOrders := [config.NumberElev][config.NumberFloors][config.NumberBtn]bool{}
	err := json.Unmarshal(ret, &tempOutput)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
	}

	for ID, orders := range *tempOutput {
		elevatorID, _ := strconv.Atoi(ID)
		for i := 0; i < config.NumberFloors; i++ {
			for j := 0; j < 2; j++ {
				//Add hall orders to set of active orders
				newAllActiveOrders[i][j][elevatorID] = orders[i][j]
			}
			//Add cab orders to set of active orders
			newAllActiveOrders[elevatorID][i][2] = input.States[ID].CabRequests[i]

		}

	}

	return newAllActiveOrders
}
