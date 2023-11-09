package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"strconv"
	"time"

	pb "homework4/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type NodeInfo struct {
	port              int32
	client            pb.NodeServiceClient
	connectedNodes    []pb.NodeServiceClient
	inCriticalSection bool //A node can enter the critical section only if already there
	timestamp         int32
	localQueue        []pb.AccessRequest
}
type Server struct {
	pb.UnimplementedNodeServiceServer
}

var connectedNodesMapPort = make(map[int32]NodeInfo)
var connectedNodesMapClient = make(map[pb.NodeServiceClient]NodeInfo)

func (n NodeInfo) AnnounceConnection(ctx context.Context, announcement *pb.ConnectionAnnouncement) (*pb.Confirmation, error) {
	//We have recieved a connection announcement, which means that a new node has established a connection to this client.
	//We must also establish a connection to this client in return. We have the information we need from the ConnectionAnnouncement
	transportCreds := insecure.NewCredentials()
	//Establish a grpc connection to the other node using addres and transport credentials
	address := ":" + strconv.Itoa(int(announcement.NodeID))
	fmt.Println("Dialing...")
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(transportCreds))
	if err != nil {
		log.Fatalf("Failed to connect  ... : %v\n", err)
	}
	fmt.Println("Making a new node")
	//we have establised a new connection to the new node. We add it to the list of node connections
	node := pb.NewNodeServiceClient(conn)
	//Add the node we have connected to our list of nodes in the system.
	//We also maintain a map which lets us find the node from its NodeID.
	nodeInfo := NodeInfo{port: announcement.NodeID, client: node}
	n.connectedNodes = append(n.connectedNodes, nodeInfo.client)
	connectedNodesMapPort[announcement.NodeID] = nodeInfo
	connectedNodesMapClient[node] = nodeInfo
	//We send back a confirmation message to indicate that the connection was esablished
	return &pb.Confirmation{}, nil
}

func (n *NodeInfo) broadcastLeavingMessage() {
	//We need to notify other connected nodes that this node is leaving the critical section.
	for _, connectedNode := range n.connectedNodes {
		node := connectedNodesMapClient[connectedNode]
		if node.port != n.port {
			_, err := connectedNode.AnnounceLeave(context.Background(), &pb.LeaveAnnouncement{
				NodeID: n.port, // Identify this node
			})
			if err != nil {
				log.Printf("Failed to send LeaveAnnouncement to node with port %d: %v\n", node.port, err)
			}
		}
	}
}

// This method is called to indicate that a node has left the critical section, which means we need to update the queue.
// we need to find the next qeueud request, that is, the request in the queue with the smallest timestamp (indicating that it was sent earlier)
func (n *NodeInfo) AnnounceLeave(ctx context.Context, announcement *pb.LeaveAnnouncement) (*pb.LeaveAnnouncementResponse, error) {

	//First we remove from the queue the request from the node which we have just learned has left the critical section, if it is there.
	n.removeRequest(announcement.NodeID)

	//We ensure the localQueue is sorted by timestamp in ascending order
	//We can sort the queue like this, where we specify a function which is used to sort the contents of the queue. Clever!
	sort.Slice(n.localQueue, func(i, j int) bool {
		return n.localQueue[i].Timestamp < n.localQueue[j].Timestamp
	})
	//Find the next request (request with the smallest timestamp, which is at the first index after sorting)
	nextRequest := n.localQueue[0]

	//find sender ID
	senderNodeID := nextRequest.NodeID
	// Find the corresponding connected node
	senderNode := connectedNodesMapPort[senderNodeID]
	//we assume that all nodes have the maintain a queue that is consistent with the queue in all other nodes. This means we don't need to go through
	//a whole lot of figuring out if all the nodes agree when we get out the next request from the queue. Instead, we can just tell the node
	//which sent the request we got out from the queue that it is free to go to the critical section directly (using an rpc call).
	senderNode.EnterCriticalSectionDirectly(context.Background(), &pb.AccessRequest{NodeID: senderNodeID, Timestamp: n.timestamp})

	// We then remove this request from the local queue, as it has been processed
	// We remove it only from the local queue, as the request will be removed from all other nodes queues when the request has been granted and the node
	//announces that it is leaving the critical section.
	if len(n.localQueue) > 0 {
		n.localQueue = n.localQueue[1:]
	}
	return &pb.LeaveAnnouncementResponse{}, nil
}

func (n *NodeInfo) AccessCriticalZone(port int32) {
	n.inCriticalSection = false
	var accessGrantedCount = 0
	fmt.Println("Peer node given access to critical section:", &n)
	for {
		if !n.inCriticalSection {
			for _, connectedNode := range n.connectedNodes {
				node := connectedNodesMapClient[connectedNode]
				if node.port != n.port {
					accessRequestResponse, err := connectedNode.RequestAccess(context.Background(), &pb.AccessRequest{NodeID: port, Timestamp: n.timestamp})
					if err != nil {
						log.Fatalf("Oh no! Something went wrong while requesting access to enter critical section")
					}
					if accessRequestResponse.Granted {
						accessGrantedCount++
					}
					//We need to increment the logical clock timestamp
					n.timestamp = UpdateTimestamp(accessRequestResponse.Timestamp, n.timestamp)

					if accessRequestResponse.Timestamp > n.timestamp {
						n.timestamp = accessRequestResponse.Timestamp
					}
					n.timestamp++
				}
				//We have now cycled through every node, sent out a request and recieved a response.
				//We check if the number of responses granting access matches the number of connected nodes (minus one, since we don't send the request to this node!)
				if accessGrantedCount == len(n.connectedNodes)-1 {
					//If it matches we enter (and leave) the critical section
					n.EnterCriticalSection()
					//After leaving we reset the inCriticalSection and accesGrantedCount variables.
					n.inCriticalSection = false
					accessGrantedCount = 0
				} else {
					//If all nodes did not grant access, we instead need to queue our request. This is already handled in each node by the call to RequestAccess,
					//but since we have not called that method on this node, we need to add the request manually to the local queue. We will then handle the request
					//in due time, the same way we handle the other requests in the queue.

					//OBS OBS OBS OBS we should probably also send out a message to all nodes saying to queue the request from this node in case they didn't already to make sure we keep the queue consistent between nodes.
					//There is a situation where nodes disagree where the queues become inconsistent between nodes which is not what we want.....
					n.localQueue = append(n.localQueue, pb.AccessRequest{NodeID: port, Timestamp: n.timestamp})
				}
			}
		}
		time.Sleep(time.Millisecond * time.Duration(1000))
	}
}

func (n *NodeInfo) EnterCriticalSectionDirectly(ctx context.Context, accessRequest *pb.AccessRequest) (*pb.Confirmation, error) {
	n.EnterCriticalSection()
	return &pb.Confirmation{}, nil
}

func (n *NodeInfo) EnterCriticalSection() {
	log.Printf("ENTERED THE CRITICAL ZONE ON PORT %v!", n.port)
	time.Sleep(time.Millisecond * time.Duration(1000))
	log.Printf("LEAVING THE CRITICAL ZONE ON PORT %v out!", n.port)
	time.Sleep(time.Millisecond * time.Duration(1000))

	// Next we need to inform all the connected nodes that we are releasing the critical section so that another node may be granted access.
	// We do this by calling broadcastLeaveMessage which calls the method AnnounceLeave on all connected nodes, which handles the logic of getting the next
	// request from the queued requests and handling it.
	n.broadcastLeavingMessage()
	//remove the request from the local queue. Since we don't need to call AnnounceLeave to our own node, we manually remove the request from the queue.
	n.removeRequest(n.port)
}

func (n *NodeInfo) removeRequest(nodeID int32) {
	var newQueue []pb.AccessRequest
	for _, request := range n.localQueue {
		if request.NodeID != nodeID {
			newQueue = append(newQueue, request)
		}
	}
	n.localQueue = newQueue
}

func (n NodeInfo) RequestAccess(ctx context.Context, accessRequest *pb.AccessRequest) (*pb.AccessRequestResponse, error) {

	senderTimestamp := accessRequest.Timestamp
	localTimestamp := n.timestamp

	n.localQueue = append(n.localQueue, pb.AccessRequest{NodeID: accessRequest.NodeID, Timestamp: senderTimestamp})

	//If the current node is not in the critical section, and if the requesting node has a lower time stamp, grant access, otherwise deny access.
	granted := false
	if !n.inCriticalSection && senderTimestamp < localTimestamp {
		granted = true
	} else {
		granted = false
		//if the above conditions are not met we queue the request until it becomes its turn to enter the critical section
		n.localQueue = append(n.localQueue, pb.AccessRequest{NodeID: accessRequest.NodeID, Timestamp: accessRequest.Timestamp})
	}
	//We return a response with information on whether or not the request has been granted.
	return &pb.AccessRequestResponse{Granted: granted, Timestamp: localTimestamp}, nil
}

func FindAnAvailablePort(standardPort int) (int, error) {
	fmt.Println("Trying to find a port...")
	for port := standardPort; port < standardPort+100; port++ {
		addr := "localhost:" + strconv.Itoa(port)
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			//The port is in use, increment and try the next one
			continue
		} else {
			fmt.Println("Connected to port:", port)
			//if no error, the port is free. Return the port.
			defer listener.Close()
			return port, nil
		}

	}
	return 0, fmt.Errorf("No free port found in the range")
}

// Cycle through the available ports in order to find the other nodes in the system and establish connections.
func EstablishConnectionToAllSystemClients(standardPort int, thisPort int, transportCreds credentials.TransportCredentials, connectedNodes []pb.NodeServiceClient) {
	for port := standardPort; port < standardPort+100; port++ {
		if port != thisPort {
			address := "localhost:" + strconv.Itoa(port)
			fmt.Printf("Trying port - %v", port)
			conn, err := grpc.Dial(address, grpc.WithTransportCredentials(transportCreds))
			if err != nil {

				fmt.Println("Hurrah! We found another node.")

				//We have found a node in the system at the port and established connection successfully.
				node := pb.NewNodeServiceClient(conn)
				//Send announcement of new connection to the node we have connected to.
				confirmed, err := node.AnnounceConnection(context.Background(), &pb.ConnectionAnnouncement{NodeID: int32(port)})
				if err != nil {
					fmt.Println(confirmed)
					log.Fatalf("Oh no! The node sent an announcement of a new connection but did not recieve a confirmation in return. Error: %v", err)
				}
				fmt.Println("Connection announced:", confirmed.String())
				//Add the node we have connected to to our list of nodes in the system.
				nodeInfo := NodeInfo{port: int32(port), client: node}
				connectedNodes = append(connectedNodes, nodeInfo.client)
			} else {
				fmt.Println("No node found at port", port)
			}
		}
	}
}

func UpdateTimestamp(incoming int32, local int32) (timestamp int32) {
	// Update the logical clock timestamp if the received timestamp is greater, then increment.
	if incoming > local {
		timestamp = incoming
	} else {
		timestamp = local
	}
	timestamp++
	return int32(timestamp)
}

func main() {

	timestamp := 0
	connectedNodes := []pb.NodeServiceClient{}

	//First we need to establish connection to the other nodes in the system
	//In order to enable us to make remote procedure calls to other nodes with grpc
	//For the purposes of this excersize we decided to simply configure the IP addresses and ports of the nodes manually.
	//To do this, we just specify a standard port and increment the port if the port is already in use and repeat until we find a free port.
	//It might have been better to do some more dynamic node discovery system.

	//we need to establish connection
	//First we find a port

	standardPort := 8000
	port, err := FindAnAvailablePort(standardPort)
	if err != nil {
		log.Fatalf("Failed to find port!: %v\n ", port)
	}
	fmt.Printf("Found port: %v\n ", port)

	fmt.Println("Registering the server now...")
	// Create a gRPC server
	grpcServer := grpc.NewServer()
	serverStructure := Server{}
	// Register your gRPC service with the server
	pb.RegisterNodeServiceServer(grpcServer, &serverStructure)

	//initialize the listener on the specified port. net.Listen listens for incoming connections with tcp socket
	fmt.Println("Listening for tcp...")
	listen, err := net.Listen("tcp", "localhost:"+strconv.Itoa(port))
	if err != nil {
		log.Fatalf("Could not listen at port: %d : %v", port, err)
	}
	fmt.Println("Listening on port", port)
	go func() {
		// Start gRPC server in a goroutine
		err := grpcServer.Serve(listen)
		if err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
		fmt.Println("gRPC server started")
	}()

	//When a port is found, next we need to establish connections to the existing nodes in the system.
	//because the ports and address are configured manually in this solution we already know the range
	//in which we expect to find those other nodes. We cycle through them

	//we create insecure transport credentials (in the context of this assignment we choose not to worry about security):
	transportCreds := insecure.NewCredentials()
	//Establish a grpc connection to the other node using addres and transport credentials
	address := ":" + strconv.Itoa(port)
	fmt.Println("Made address:", address)
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(transportCreds))
	if err != nil {
		log.Fatalf("Failed to connect  ... : %v\n", err)
	}
	fmt.Println("Client connection to address: ", address)

	//Create a grpc client instance to represent local node (this node).
	//In a way we are establishing connection to our own node ... which may seem weird, but it does make sense:
	//The idea is to have a local representation of this node, that can interact with other nodes in the system,
	//since it is a client that can call remote procedures and recieve remote procedure calls through grpc.
	thisNode := pb.NewNodeServiceClient(conn)
	fmt.Println("New client for this node generated:", &thisNode)

	fmt.Println("Attempting to connect to all clients in system...")
	//Now we want to establish connection to all other nodes in the system.
	EstablishConnectionToAllSystemClients(standardPort, port, transportCreds, connectedNodes)

	//We have now established connection to all other nodes in the system, notified them, and they have established connections back in return.

	//Next: We try to inter the critical section
	//generate node
	fmt.Println("Making a node...")
	node := &NodeInfo{port: int32(port), client: thisNode, connectedNodes: connectedNodes, timestamp: int32(timestamp)}

	fmt.Println("Press 1 to start critical system operation, or 2 to exit: ")
	var input int
	fmt.Scan(&input)

	for input != 1 {
		if input == 2 {
			os.Exit(1)
		} else {
			fmt.Println("Unknown command, try again: ")
			fmt.Scan(&input)
		}
	}

	fmt.Println("Attempting to access the critical zone...")
	node.AccessCriticalZone(int32(port))
	select {}
}
