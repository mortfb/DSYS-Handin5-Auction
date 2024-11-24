# DSYS-Handin5-Auction

## How to run the program.
Currently the program only supports having three server (auction) nodes
Start by starting the three auctions (servers), by entering the Auction folder and typing go run auction.go in seperate terminals. You will then be requested to provide one of following ports (5050, 5051 and 5052) each server will be using. Each port must have one server using it.

The server nodes, will then start an election and decide who will be the leader among them. The leader is the one responsible for updating the data and sending it out to the other nodes.

Then using new terminals launch the clients, by entering the client folder and typing client.go in seperate terminals. The recommended number of clients are 2-3. 
You will then be asked to type a name for the client. And now you have entered the auction. A client can enter at any time, as long as the Auction is still running and not about to shut down.
Now a client can start bidding, by entering their bid in the terminal. The bid must be higher than the current highest bid, which will be sent to the clients, otherwise the Auction wont accept the bid.
The client will print the result/status of the auction in the terminal every 5 seconds. 

The auction runs until its updateCounter reaches 100, then the leader will send the final result of the auction and every server node and client will terminate.

Client commands:
- A client can type '-1' instead of an actual bid, which then will show the current result of the action manually.
- A client can type '-2' instead of an actual bid, to terminate, leaving the auction.

## How to terminate a node
The program can currently handle ONE node failure in the group of server nodes. Go to the terminal belonging to the server node you want to fail and then press ctrl + c, to terminate the node. If it is the leader node, that you decide to crash, a new election will be held between the remaining nodes, otherwise the leader will stop sending updates to the now dead node.  
