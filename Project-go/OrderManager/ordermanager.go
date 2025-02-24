package ordermanager

import (
	requests "Project-go/driver-go/Requests"
	"Project-go/driver-go/elevio"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
)

var (
	AllActiveOrders [4][3][3]bool
	NewRequests     [4][3][3]bool
	orderCounter    [3]int
	elevators       [3]elevio.Elevator
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

func RecievedOrdersSlave(e elevio.Elevator) {

	e = requests.RequestClearAtCurrentFloor(e)
	elevators[e.ElevatorID] = e

	maxCounterValue := orderCounter[e.ElevatorID]
	for i := 0; i < e.NumFloors; i++ {
		for j := 0; j < 3; j++ {
			if e.Requests[i][j] > orderCounter[e.ElevatorID] {
				NewRequests[i][j][e.ElevatorID] = true
				if e.Requests[i][j] > maxCounterValue {
					maxCounterValue = e.Requests[i][j]
				}
			}
		}
	}
	orderCounter[e.ElevatorID] = maxCounterValue

	input := formatRequests(elevators, AllActiveOrders, NewRequests)

	AllActiveOrders = requestAssigner(input)
}

func formatRequests(elevators [3]elevio.Elevator, allActiveOrders [4][3][3]bool,
	newRequests [4][3][3]bool) HRAInput {
	h := [][2]bool{}
	c1 := []bool{}
	c2 := []bool{}
	c3 := []bool{}

	for i := 0; i < elevators[0].NumFloors; i++ {
		for j := 0; j < 2; j++ {
			for k := 0; k < 3; k++ {
				h[i][j] = h[i][j] || allActiveOrders[i][j][k] || newRequests[i][j][k]
			}
		}

		c1[i] = allActiveOrders[i][2][0] || newRequests[i][2][0]
		c2[i] = allActiveOrders[i][2][1] || newRequests[i][2][1]
		c3[i] = allActiveOrders[i][2][2] || newRequests[i][2][2]

	}

	input := HRAInput{
		HallRequests: h,
		States: map[string]HRAElevState{
			"one": HRAElevState{
				Behavior:    behaviorToString[elevators[0].Behaviour],
				Floor:       elevators[0].CurrentFloor,
				Direction:   motorDirectionToString[elevators[0].Direction],
				CabRequests: c1,
			},
			"two": HRAElevState{
				Behavior:    behaviorToString[elevators[1].Behaviour],
				Floor:       elevators[1].CurrentFloor,
				Direction:   motorDirectionToString[elevators[1].Direction],
				CabRequests: c2,
			},
			"three": HRAElevState{
				Behavior:    behaviorToString[elevators[2].Behaviour],
				Floor:       elevators[2].CurrentFloor,
				Direction:   motorDirectionToString[elevators[2].Direction],
				CabRequests: c3,
			},
		},
	}
	return input
}

func requestAssigner(input HRAInput) [4][3][3]bool {

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

	output := new(map[string][][2]bool)
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
	}

	fmt.Printf("output: \n")
	for k, v := range *output {
		fmt.Printf("%6v :  %+v\n", k, v)
	}

	var newAllActiveOrders [4][3][3]bool

	// Map the output back to the [4][3][3]bool format
	//Plz fix someone
	for floor := 0; floor < 4; floor++ {
		for button := 0; button < 3; button++ {
			for elev := 0; elev < 3; elev++ {
				if button < 2 {
					newAllActiveOrders[floor][button][elev] = (*output)[fmt.Sprintf("%d-%d", floor, button)][elev]
				} else {
					newAllActiveOrders[floor][button][elev] = (*output)[fmt.Sprintf("%d-%d", floor, button)][0]
				}
			}
		}
	}

	return newAllActiveOrders

}

//Add function here polling on the msgArrived arrived channel, when a new order comes
//Also don't run everything if no new requests
