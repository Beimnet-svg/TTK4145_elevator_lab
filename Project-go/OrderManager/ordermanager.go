package ordermanager

import (
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
	//Fix config here
	AllActiveOrders [3][4][3]bool
	NewRequests     [3][4][3]bool
	orderCounter    [3]int
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

func UpdateOrders(e elevio.Elevator, receiver chan [3][4][3]bool) {
	//Clear orders at current floor based on elevator state
	AllActiveOrders = requests.RequestClearAtCurrentFloor(e, AllActiveOrders)
	//Fetch active elevators from master-slave module
	elevators := masterslavedist.FetchElevators()

	maxCounterValue := orderCounter[e.ElevatorID]
	for i := 0; i < e.NumFloors; i++ {
		for j := 0; j < 3; j++ {
			//Based on the counter values in e.Requests we can determine if we have a new order
			if e.Requests[i][j] > orderCounter[e.ElevatorID] {
				NewRequests[i][j][e.ElevatorID] = true
				if e.Requests[i][j] > maxCounterValue {
					//Find the highest counter value in the elevator
					maxCounterValue = e.Requests[i][j]
				}
			}
		}
	}

	//If we have a new order we redistribute hall orders and set new order counter
	if maxCounterValue > orderCounter[e.ElevatorID] {
		orderCounter[e.ElevatorID] = maxCounterValue
		input := formatInput(elevators, AllActiveOrders, NewRequests)
		AllActiveOrders = assignRequests(input)
		receiver <- AllActiveOrders
	}

}

func formatInput(elevators []elevio.Elevator, allActiveOrders [3][4][3]bool,
	newRequests [3][4][3]bool) HRAInput {
	hallRequests := [][2]bool{}
	cabRequests := [3][]bool{}

	for i := 0; i < 3; i++ {
		for j := 0; j < 4; j++ {
			for k := 0; k < 2; k++ {
				//Extract hallrequests from current and new orders
				hallRequests[j][k] = hallRequests[j][k] || allActiveOrders[i][j][k] || newRequests[i][j][k]
			}
		}
		for j := 0; j < 4; j++ {
			//Extract cabrequests from current and new orders
			cabRequests[i][j] = allActiveOrders[i][j][2] || newRequests[i][j][2]
		}

	}

	input := HRAInput{
		HallRequests: hallRequests,
		States:       map[string]HRAElevState{},
	}

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

func assignRequests(input HRAInput) [3][4][3]bool {

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

	ret, err := exec.Command("../hall_request_assigner/"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Println("exec.Command error: ", err)
		fmt.Println(string(ret))
	}

	return transformOutput(ret, input)

}

func transformOutput(ret []byte, input HRAInput) [3][4][3]bool {

	tempOutput := new(map[string][][2]bool)
	newAllActiveOrders := [3][4][3]bool{}
	err := json.Unmarshal(ret, &tempOutput)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
	}

	for ID, orders := range *tempOutput {
		elevatorID, _ := strconv.Atoi(ID)
		for i := 0; i < 4; i++ {
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

//Add function here polling on the msgArrived arrived channel, when a new order comes
