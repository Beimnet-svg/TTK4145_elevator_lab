package requests

import (
	"Driver-go/elevio"
	"testing"
)

func TestRequestShouldStop(t *testing.T) {
    tests := []struct {
        elevator elevio.Elevator
        expected bool
    }{
        {
            elevator: elevio.Elevator{
                Direction:    elevio.MD_Up,
                CurrentFloor: 1,
                Requests: []elevio.ButtonEvent{
                    {Floor: 1, Button: elevio.BT_HallUp},
                },
            },
            expected: true,
        },
        {
            elevator: elevio.Elevator{
                Direction:    elevio.MD_Down,
                CurrentFloor: 2,
                Requests: []elevio.ButtonEvent{
                    {Floor: 2, Button: elevio.BT_HallDown},
                },
            },
            expected: true,
        },
        {
            elevator: elevio.Elevator{
                Direction:    elevio.MD_Stop,
                CurrentFloor: 3,
                Requests:     []elevio.ButtonEvent{},
            },
            expected: true,
        },
    }

    for _, test := range tests {
        t.Run("", func(t *testing.T) {
            result := RequestShouldStop(test.elevator)
            if result != test.expected {
                t.Errorf("expected %v, got %v", test.expected, result)
            }
        })
    }
}

func TestRequestChooseDir(t *testing.T) {
    tests := []struct {
        elevator  elevio.Elevator
        button    elevio.ButtonType
        expectedD elevio.MotorDirection
        expectedB elevio.ElevatorBehaviour
    }{
        {
            elevator: elevio.Elevator{
                Direction:    elevio.MD_Up,
                CurrentFloor: 1,
                Requests: []elevio.ButtonEvent{
                    {Floor: 2, Button: elevio.BT_HallUp},
                },
            },
            button:    elevio.BT_HallUp,
            expectedD: elevio.MD_Up,
            expectedB: elevio.EB_Moving,
        },
        {
            elevator: elevio.Elevator{
                Direction:    elevio.MD_Down,
                CurrentFloor: 2,
                Requests: []elevio.ButtonEvent{
                    {Floor: 1, Button: elevio.BT_HallDown},
                },
            },
            button:    elevio.BT_HallDown,
            expectedD: elevio.MD_Down,
            expectedB: elevio.EB_Moving,
        },
        {
            elevator: elevio.Elevator{
                Direction:    elevio.MD_Stop,
                CurrentFloor: 3,
                Requests:     []elevio.ButtonEvent{},
            },
            button:    elevio.BT_Cab,
            expectedD: elevio.MD_Stop,
            expectedB: elevio.EB_Idle,
        },
    }

    for _, test := range tests {
        t.Run("", func(t *testing.T) {
            resultD, resultB := RequestChooseDir(test.elevator, test.button)
            if resultD != test.expectedD || resultB != test.expectedB {
                t.Errorf("expected (%v, %v), got (%v, %v)", test.expectedD, test.expectedB, resultD, resultB)
            }
        })
    }
}

func TestReqestShouldClearImmideatly(t *testing.T) {
    tests := []struct {
        elevator  elevio.Elevator
        floor     int
        button    elevio.ButtonType
        expected  bool
    }{
        {
            elevator: elevio.Elevator{
                Direction:    elevio.MD_Up,
                CurrentFloor: 1,
            },
            floor:    1,
            button:   elevio.BT_HallUp,
            expected: true,
        },
        {
            elevator: elevio.Elevator{
                Direction:    elevio.MD_Down,
                CurrentFloor: 2,
            },
            floor:    2,
            button:   elevio.BT_HallDown,
            expected: true,
        },
        {
            elevator: elevio.Elevator{
                Direction:    elevio.MD_Stop,
                CurrentFloor: 3,
            },
            floor:    3,
            button:   elevio.BT_Cab,
            expected: true,
        },
    }

    for _, test := range tests {
        t.Run("", func(t *testing.T) {
            result := ReqestShouldClearImmideatly(test.elevator, test.floor, test.button)
            if result != test.expected {
                t.Errorf("expected %v, got %v", test.expected, result)
            }
        })
    }
}

func TestRequestClearAtCurrentFloor(t *testing.T) {
    tests := []struct {
        elevator  elevio.Elevator
        expected  elevio.Elevator
    }{
        {
            elevator: elevio.Elevator{
                CurrentFloor: 1,
                Requests: []elevio.ButtonEvent{
                    {Floor: 1, Button: elevio.BT_Cab},
                    {Floor: 2, Button: elevio.BT_HallUp},
                },
            },
            expected: elevio.Elevator{
                CurrentFloor: 1,
                Requests: []elevio.ButtonEvent{
                    {Floor: 2, Button: elevio.BT_HallUp},
                },
            },
        },
        {
            elevator: elevio.Elevator{
                CurrentFloor: 2,
                Requests: []elevio.ButtonEvent{
                    {Floor: 2, Button: elevio.BT_Cab},
                    {Floor: 3, Button: elevio.BT_HallDown},
                },
            },
            expected: elevio.Elevator{
                CurrentFloor: 2,
                Requests: []elevio.ButtonEvent{
                    {Floor: 3, Button: elevio.BT_HallDown},
                },
            },
        },
    }

    for _, test := range tests {
        t.Run("", func(t *testing.T) {
            result := RequestClearAtCurrentFloor(test.elevator)
            if len(result.Requests) != len(test.expected.Requests) {
                t.Errorf("expected %v, got %v", test.expected.Requests, result.Requests)
            }
            for i := range result.Requests {
                if result.Requests[i] != test.expected.Requests[i] {
                    t.Errorf("expected %v, got %v", test.expected.Requests[i], result.Requests[i])
                }
            }
        })
    }
}

func TestHasRequestAtFloor(t *testing.T) {
    tests := []struct {
        elevator  elevio.Elevator
        btnType   elevio.ButtonType
        expected  bool
    }{
        {
            elevator: elevio.Elevator{
                CurrentFloor: 1,
                Requests: []elevio.ButtonEvent{
                    {Floor: 1, Button: elevio.BT_HallUp},
                },
            },
            btnType:  elevio.BT_HallUp,
            expected: true,
        },
        {
            elevator: elevio.Elevator{
                CurrentFloor: 2,
                Requests: []elevio.ButtonEvent{
                    {Floor: 1, Button: elevio.BT_HallUp},
                },
            },
            btnType:  elevio.BT_HallUp,
            expected: false,
        },
        {
            elevator: elevio.Elevator{
                CurrentFloor: 3,
                Requests: []elevio.ButtonEvent{
                    {Floor: 3, Button: elevio.BT_Cab},
                },
            },
            btnType:  elevio.BT_Cab,
            expected: true,
        },
        {
            elevator: elevio.Elevator{
                CurrentFloor: 4,
                Requests: []elevio.ButtonEvent{
                    {Floor: 3, Button: elevio.BT_Cab},
                },
            },
            btnType:  elevio.BT_Cab,
            expected: false,
        },
    }

    for _, test := range tests {
        t.Run("", func(t *testing.T) {
            result := hasRequestAtFloor(test.elevator, test.btnType)
            if result != test.expected {
                t.Errorf("expected %v, got %v", test.expected, result)
            }
        })
    }
}