package ordermanager

import (
	"Project-go/driver-go/elevio"
)

var AllActiveOrders [4][3][3]int

func RecievedOrdersSlave(e elevio.Elevator) {

}

func RecievedOrdersMaster(orders [4][3][3]int) {

}

func FetchActiveOrders() [4][3][3]int{
	return AllActiveOrders
}