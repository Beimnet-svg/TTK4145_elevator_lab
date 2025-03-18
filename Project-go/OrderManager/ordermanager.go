package ordermanager

import (
	config "Project-go/Config"
	masterslavedist "Project-go/MasterSlaveDist"
	elevfsm "Project-go/SingleElev/ElevFsm"
	elevio "Project-go/SingleElev/Elevio"
	requests "Project-go/SingleElev/Requests"

	"encoding/json"
	"os/exec"
	"runtime"
	"strconv"
)

var (
	allActiveOrders [config.NumberElev][config.NumberFloors][config.NumberBtn]bool
	orderCounter    [config.NumberElev]int
	elevState       [config.NumberElev]elevio.Elevator
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

	elevState[e.ElevatorID] = e

	allActiveOrders = requests.RequestClearAtCurrentFloor(e, allActiveOrders)

	maxCounterValue := orderCounter[e.ElevatorID]
	maxCounterValue, newRequests = findNewRequests(e, maxCounterValue, newRequests)
	orderCounter[e.ElevatorID] = maxCounterValue

	aliveElevatorStates := masterslavedist.FetchActiveElevators(elevState)
	input, cabRequests := formatInput(aliveElevatorStates, allActiveOrders, newRequests)
	allActiveOrders = assignRequests(input, cabRequests)

	//Send updated orders to all elevators
	activeOrderChan <- allActiveOrders
}

func findNewRequests(e elevio.Elevator, maxCounterValue int, NewRequests [config.NumberElev][config.NumberFloors][config.NumberBtn]bool) (int, [config.NumberElev][config.NumberFloors][config.NumberBtn]bool) {

	for i := 0; i < config.NumberFloors; i++ {
		for j := 0; j < config.NumberBtn; j++ {

			if e.Requests[i][j] > orderCounter[e.ElevatorID] {
				NewRequests[e.ElevatorID][i][j] = true

				if e.Requests[i][j] > maxCounterValue {
					maxCounterValue = e.Requests[i][j]
				}
			}
		}
	}

	return maxCounterValue, NewRequests
}

// Format input to be used in the cost function
func formatInput(aliveElevatorStates []elevio.Elevator, allActiveOrders [config.NumberElev][config.NumberFloors][config.NumberBtn]bool,
	newRequests [config.NumberElev][config.NumberFloors][config.NumberBtn]bool) (HRAInput, [config.NumberElev][]bool) {

	hallRequests := make([][2]bool, config.NumberFloors)
	cabRequests := [config.NumberElev][]bool{}

	for i := range cabRequests {
		cabRequests[i] = make([]bool, config.NumberFloors)
	}

	for i := 0; i < config.NumberElev; i++ {
		for j := 0; j < config.NumberFloors; j++ {
			for k := 0; k < 2; k++ {
				hallRequests[j][k] = hallRequests[j][k] || allActiveOrders[i][j][k] || newRequests[i][j][k]
			}
		}
		for j := 0; j < config.NumberFloors; j++ {
			cabRequests[i][j] = allActiveOrders[i][j][2] || newRequests[i][j][2]
		}

	}

	input := HRAInput{
		HallRequests: hallRequests,
		States:       map[string]HRAElevState{},
	}

	for _, e := range aliveElevatorStates {
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

	jsonBytes, _ := json.Marshal(input)

	ret, _ := exec.Command(hraExecutable, "-i", string(jsonBytes)).CombinedOutput()

	return transformOutput(ret, cabRequests)

}

// Transform the output from the cost function to a format that can be used in the ordermanager
func transformOutput(ret []byte, cabRequests [config.NumberElev][]bool) [config.NumberElev][config.NumberFloors][config.NumberBtn]bool {

	tempOutput := new(map[string][][2]bool)
	newAllActiveOrders := [config.NumberElev][config.NumberFloors][config.NumberBtn]bool{}
	json.Unmarshal(ret, &tempOutput)

	for ID, orders := range *tempOutput {
		elevatorID, _ := strconv.Atoi(ID)

		for i := 0; i < config.NumberFloors; i++ {
			for j := 0; j < 2; j++ {
				newAllActiveOrders[elevatorID][i][j] = orders[i][j]
			}
		}
	}

	for elevID := 0; elevID < config.NumberElev; elevID++ {
		for floor := 0; floor < config.NumberFloors; floor++ {
			newAllActiveOrders[elevID][floor][2] = cabRequests[elevID][floor]
		}
	}

	return newAllActiveOrders
}

// Apply backup to new master
func ApplyBackupOrders(setMaster chan bool, activeOrderChan chan [config.NumberElev][config.NumberFloors][config.NumberBtn]bool) {
	for a := range setMaster {
		if a {
			allActiveOrders = elevfsm.GetAllActiveOrders()
		}
	}
}

func ResetOrderCounter(elevDied chan int) {
	for ID := range elevDied {
		orderCounter[ID] = 0
	}
}
