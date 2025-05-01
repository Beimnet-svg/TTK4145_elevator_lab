# TTK4145_elevator_lab

Tripple elevator lab in ttk4145

Overall design:
The elevator is designed as a master-slave. We have a master which has the ability to give orders to elevators, and slaves which only services the orders which they get sent from the master. The slaves are the backups. The slaves will always store a backup of the newest orders which have been delegated to all the elevators and take over if a master fails. Everything is sent through UDP, with a tunable sending frequency

Config file -> Configure your elevator to your liking. Add number of elevators and floors, etc.

Driver-go -> The basic one elevator system with a FSM structure. The elevator is designed like the given C-code, apart from the deleting of orders, which is done in the ordermanager in the master.

MasterSlaveDist -> Master slave distributor. We always want there to be only one master of the active system (If an elevator is disconnected it can be a master of itself), and that master being the one with the newest information. We therefore have a disconnected bool that is set to true when an elevator doesn't recieve any alive messages, and all elevators are initiallized as slaves to ensure smooth behaviour after power toggle. When the master dies the slave with lowest elevator ID that is still connected becomes the master.

When a disconnected elevator reconnects it starts a timer looking for master messages(waitForMasterMsg). If it does not recieve one in the given timer time it sets itself as master, if it does, it sets itself as slave and stores its active orders as new requests to send to its new master. If two disconnected masters reconnect, both will turn to slave. We therefore also have a timer, waiting for master messages, which triggers a master election if no masters are observed within the given timer time.

Networking -> A common sender and reciever module for recieving and sending data. The data being sent over is an "Order message"-struct, which is being decoded. The slave sends its elevator struct, containing new requests and elevator state, and the master sends all active orders in the system. Both also send their ID. These are sent out periodically through UDP and works as a heartbeat as well.

OrderManager -> This is where all the orders are being processed. When new buttons are pressed in either the master or the slaves, we increment our order counter by one and add the order counter value to the request array. This is then compared with a value in the master ordermanager to determine if requests are new or old. The ordermanager utilizes the new requests combined with the current request to destribute orders to the elevator which has the cost-optimal path. This then gets sent from the master to the slaves.

For report:

- When a disconnected elevator reconnects it will get its cab orders from before disconnecting. This was done to ensure no orders are lost when an elevators dies or disconnects. This will hurt performance but ensure fault tolerance
- Using get functions even though go has other functionallity to ensure that it is clear where variables from other modules are used
- Boolean values sent on channels sometimes gets stuck, so have to send two times. When we didn't need to send boolean values we used int instead
- Comment on choice of timer structure
- If packet loss over some threshold, go into single elevator mode, as our current approach is not reliable when packet loss > 80-90%
- Comment on get functions
- Just send elevator, dont need anything else
- UpdateOrderCOunter -> This was made, knowing that it would not lead to race conditions, but after reasessing the usecase, and go language, it should have been changed to a chanel.

Stuff to do before delivering

- Go through everything again, see if we are able to detect errors/bad code
