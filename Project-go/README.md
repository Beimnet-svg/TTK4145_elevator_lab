# TTK4145_elevator_lab

Tripple elevator lab in ttk4145

Overall design:
The elevator is designed as a master-slave. We have a master which has the ability to give orders to elevators, and slaves which only services the orders which they get sent from the master. The slaves are the backups. The slaves will always store a backup of the newest orders which have been delegated to all the elevators and take over if a master fails.

Config file -> Configure your elevator to your liking. Add number of elevators and floors, etc.

Driver-go -> The basic one elevator system with a FSM structure. The elevator is designed like the given C-code, apart from the deleting of orders, which is done in the ordermanager in the master.

MasterSlaveDist -> Master slave distributor. We always want there to be only one master of the active system (If an elevator is disconnected it can be a master of itself, but should not try to send messages to other elevators), and that master being the one with the newest information. We therefore have a disconnected bool that is set to true when an elevator doesn't recieve any alive messages, and all elevators are initiallized as slaves to ensure smooth behaviour after power toggle. When the master dies the slave with lowest elevator ID that is still connected becomes the master.

Networking -> A common sender and reciever module for recieving and sending data. The data being sent over is an "Order message"-struct, which is being decoded. The slave sends its elevator struct, containing new requests and elevator state, and the master sends all active orders in the system. These are sent out periodically and works as a heartbeat as well.

OrderManager -> This is where all the orders are being processed. When new buttons are pressed in either the maste ror the slaves, the ordermanager will make sure that the elevator which has the cost-optimal path takes this order. This then gets sent from the master to the slaves.

Current bugs:

- When pressing up and down in a floor, it removes both of them when reaching the floor if they are the only orders left
- There is a wierd bug somewhere were an elevator just stops moving, should be an acceptence test for if elevator hasnt moved for a certain amount of time, restart it.
- All hall buttons should light up when there is an order there.
- Code quality: Instead of using setElevID in main, we can set config.ID and use that all over the place instead.
- If two elevators are running thinking they are disconnectd, when reconnecting both become slaves. Fix: send the disconnected bool, set the one with lowest id as master.

Questions for TA:

- When getting values from one module to the next, should we use a GetFunction, or make the variable global, or send them on channels.
