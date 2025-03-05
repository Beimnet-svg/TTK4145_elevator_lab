# TTK4145_elevator_lab
Tripple elevator lab in ttk4145

Overall design: 
The elevator is designed as a master-slave. We have a master which has the ability to give orders to elevators, and slaves which only services the orders which they get sent from the master. The slaves are the backups. The slaves will always store a backup of the newest orders which have been delegated to all the elevators and take over if a master fails. 

Config file -> Configure your elevator to your liking. Add number of elevators and floors, etc. 

Driver-go -> The basic one elevator system such as how the FSM is designed. The elevator is designed as a typical FSM, much like the given C-code. 

MasterSlaveDist -> Master slave distributor. This is designed such that the elevator with the lowest ID and the elevator which has been working the longest, in other words, has the updated info stays master. And the master which reconnects steps down. 

Networking -> A common sender and reciever function for recieving and sending data. The data being sent over is an "Order message"-struct, which is being decoded. 

OrderManager -> This is where all the orders are being processed. When new buttons are pressed in either the maste ror the slaves, the ordermanager will make sure that the elevator which has the cost-optimal path takes this order. This then gets sent from tha master to the slaves. 

Current bugs:

- When pressing up and down in a floor, it removes both of them when reaching the floor


