package ordermanager

import (
	config "Project-go/Config"
	masterslavedist "Project-go/MasterSlaveDist"
	elevfsm "Project-go/SingleElev/ElevFsm"
	elevio "Project-go/SingleElev/Elevio"
	requests "Project-go/SingleElev/Requests"

	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
)

var (
	allActiveOrders [config.NumberElev][config.NumberFloors][config.NumberBtn]bool
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

func GetAllActiveOrder() [config.NumberElev][config.NumberFloors][config.NumberBtn]bool {
	return allActiveOrders
}

func GetOrderCounter() [config.NumberElev]int {
	return orderCounter
}

func UpdateOrderCounter(newOrderCounter [config.NumberElev]int) {
	orderCounter = newOrderCounter
}

func UpdateOrders(e elevio.Elevator, activeOrderChan chan [config.NumberElev][config.NumberFloors][config.NumberBtn]bool) {
	newRequests := [config.NumberElev][config.NumberFloors][config.NumberBtn]bool{}

	ElevState[e.ElevatorID] = e

	allActiveOrders = requests.RequestClearAtCurrentFloor(e, allActiveOrders)

	//Create a maxCounterValue, where every value in e.requests higher than this
	//value is considered a new order
	maxCounterValue := orderCounter[e.ElevatorID]

	//If we have a new order we redistribute hall orders and set new order counter
	CheckIfNewOrders(e, &maxCounterValue, &newRequests)

	aliveElevatorStates := masterslavedist.FetchAliveElevators(ElevState)
	orderCounter[e.ElevatorID] = maxCounterValue
	input, cabRequests := formatInput(aliveElevatorStates, allActiveOrders, newRequests)
	allActiveOrders = assignRequests(input, cabRequests)

	//Send updated orders to all elevators
	activeOrderChan <- allActiveOrders
}

func CheckIfNewOrders(e elevio.Elevator, maxCounterValue *int, NewRequests *[config.NumberElev][config.NumberFloors][config.NumberBtn]bool) bool {
	//Check if there are new orders in the system

	for i := 0; i < config.NumberFloors; i++ {
		for j := 0; j < config.NumberBtn; j++ {
			//Based on the counter values in e.Requests we can determine if we have a new order
			if e.Requests[i][j] > orderCounter[e.ElevatorID] {
				NewRequests[e.ElevatorID][i][j] = true
				if e.Requests[i][j] > *maxCounterValue {
					//Find the highest counter value in the elevator
					*maxCounterValue = e.Requests[i][j]
				}
			}
		}
	}

	return *maxCounterValue > orderCounter[e.ElevatorID]
}

// Format input to be used in the cost function
func formatInput(elevators []elevio.Elevator, allActiveOrders [config.NumberElev][config.NumberFloors][config.NumberBtn]bool,
	newRequests [config.NumberElev][config.NumberFloors][config.NumberBtn]bool) (HRAInput, [config.NumberElev][]bool) {

	hallRequests := make([][2]bool, config.NumberFloors)
	cabRequests := [config.NumberElev][]bool{}

	//Init cabRequests
	for i := range cabRequests {
		cabRequests[i] = make([]bool, config.NumberFloors)
	}

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
	return input, cabRequests
}

func assignRequests(input HRAInput, cabRequests [config.NumberElev][]bool) [config.NumberElev][config.NumberFloors][config.NumberBtn]bool {

	hraExecutable := ""
	switch runtime.GOOS {
	case "linux":
		hraExecutable = "OrderManager/hall_request_assigner"
	case "windows":
		hraExecutable = "./OrderManager/hall_request_assigner.exe"
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

	return transformOutput(ret, input, cabRequests)

}

// Transform the output from the cost function to a format that can be used in the ordermanager
func transformOutput(ret []byte, input HRAInput, cabRequests [config.NumberElev][]bool) [config.NumberElev][config.NumberFloors][config.NumberBtn]bool {

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
				newAllActiveOrders[elevatorID][i][j] = orders[i][j]
			}
		}

	}

	//Add cab orders to set of active orders
	for elevID := 0; elevID < config.NumberElev; elevID++ {
		for floor := 0; floor < config.NumberFloors; floor++ {
			newAllActiveOrders[elevID][floor][2] = cabRequests[elevID][floor]
		}
	}

	return newAllActiveOrders
}

// Apply backup to new master
func ApplyBackupOrders(setMaster chan bool, activeOrderChan chan [config.NumberElev][config.NumberFloors][config.NumberBtn]bool) {
	for {
		select {
		case a := <-setMaster:
			if a {

				allActiveOrders = elevfsm.GetAllActiveOrders()
			}
		}
	}
}

func ResetOrderCounter(elevDied chan int) {
	for {
		select {
		case a := <-elevDied:
			orderCounter[a] = 0
		}
	}
}
