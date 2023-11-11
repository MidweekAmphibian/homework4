package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	pb "homework4/proto"

	"github.com/briandowns/spinner"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var mu sync.Mutex

type NodeInfo struct {
	port                          int32
	client                        pb.NodeServiceClient
	connectedNodes                []pb.NodeServiceClient
	inCriticalSection             bool
	tryingToAccessCriticalSection bool
	timestamp                     int32
	localQueue                    []pb.AccessRequest
}
type Server struct {
	pb.UnimplementedNodeServiceServer
	node NodeInfo
}

var connectedNodesMapPort = make(map[int32]NodeInfo)
var connectedNodesMapClient = make(map[pb.NodeServiceClient]NodeInfo)

func (s *Server) IExist(context.Context, *pb.Confirmation) (*pb.Confirmation, error) {
	return &pb.Confirmation{}, nil
}

func (s *Server) AnnounceQueuedRequest(ctx context.Context, queuedRequest *pb.AccessRequest) (*pb.Confirmation, error) {
	//This method is called when a node is not granted access by all the nodes it requested access from.
	//In order to maintain a consistent queue, we check if the queued request from the node is already in the queue, and if it isn't we add it.
	//fmt.Println("I HAVE RECIEVED A CALL TO UPDATE MY QUEUE")
	var isAlreadyQueued = false
	for _, request := range s.node.localQueue {
		if request.NodeID == queuedRequest.NodeID {
			isAlreadyQueued = true
		}
	}
	if !isAlreadyQueued {
		s.node.localQueue = append(s.node.localQueue, pb.AccessRequest{NodeID: queuedRequest.NodeID, Timestamp: queuedRequest.Timestamp})
	}
	return &pb.Confirmation{}, nil
}

func (s *Server) AnnounceConnection(ctx context.Context, announcement *pb.ConnectionAnnouncement) (*pb.Confirmation, error) {
	//We have recieved a connection announcement, which means that a new node has established a connection to this client.
	//We must also establish a connection to this client in return. We have the information we need from the ConnectionAnnouncement
	transportCreds := insecure.NewCredentials()
	//Establish a grpc connection to the other node using addres and transport credentials
	address := ":" + strconv.Itoa(int(announcement.NodeID))
	fmt.Println("NEW NODE JOINED.")
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(transportCreds))
	if err != nil {
		log.Fatalf("Failed to connect  ... : %v\n", err)
	}
	//we have establised a new connection to the new node. We add it to the list of node connections
	node := pb.NewNodeServiceClient(conn)
	//Add the node we have connected to our list of nodes in the system.
	//We also maintain a map which lets us find the node from its NodeID.
	nodeInfo := NodeInfo{port: announcement.NodeID, client: node}
	s.node.connectedNodes = append(s.node.connectedNodes, nodeInfo.client)
	connectedNodesMapPort[announcement.NodeID] = nodeInfo
	connectedNodesMapClient[node] = nodeInfo
	//We send back a confirmation message to indicate that the connection was esablished
	return &pb.Confirmation{}, nil
}

func (s *Server) LeaveCriticalSection() {

	log.Printf("\n\n * * * LEAVING THE CRITICAL ZONE... (PORT %v) * * * \n\n\n", s.node.port)
	s.node.inCriticalSection = false

	// Next we need to inform all the connected nodes that we are releasing the critical section so that another node may be granted access.
	// We do this by calling the method AnnounceLeave on all connected nodes, which handles the logic of getting the next request from the queued requests and handling it.

	//We notify other connected nodes that this node is leaving the critical section.
	//We do this so that they remove the node from their local queue, if it is there.
	for _, connectedNode := range s.node.connectedNodes {
		node := connectedNodesMapClient[connectedNode]
		if node.port != s.node.port {
			// Node announces that it is leaving
			_, err := connectedNode.AnnounceLeave(context.Background(), &pb.LeaveAnnouncement{
				NodeID: s.node.port, // Identify this node
			})
			if err != nil {
				log.Printf("Failed to send LeaveAnnouncement to node with port %d: %v\n", node.port, err)
			}
		}
	}

	//Next we need to tell the node that sent the next request in the queue that it is allowed to enter the critical section.
	//We do that here to avoid every node recieving a leaveAnnouncement telling the next node what to do at once.

	//We ensure the localQueue is sorted by timestamp in ascending order
	//(We can sort the queue like this, where we specify a function which is used to sort the contents of the queue. Clever!)
	//
	sort.Slice(s.node.localQueue, func(i, j int) bool {
		return s.node.localQueue[i].Timestamp < s.node.localQueue[j].Timestamp
	})
	//Find the next request (request with the smallest timestamp, which is at the first index after sorting)
	//If there are no requests queued we do nothing.
	if len(s.node.localQueue) > 0 {
		nextRequest := s.node.localQueue[0]

		//find sender ID
		nextRequestSenderNodeID := nextRequest.NodeID
		// Find the corresponding connected node
		nextRequestSenderNode := connectedNodesMapPort[nextRequestSenderNodeID]
		//we assume that all nodes maintain a queue that is consistent with the queue in all other nodes. This means we don't need to go through
		//a whole lot of figuring out if all the nodes agree when we get out the next request from the queue. Instead, we can just tell the node
		//which sent the request we got out from the queue that it is free to go to the critical section directly (using an rpc call).

		go nextRequestSenderNode.client.EnterCriticalSectionDirectly(context.Background(), &pb.AccessRequest{NodeID: s.node.port, Timestamp: s.node.timestamp})

		// We then remove this request from the local queue, as it has been processed
		// We remove it only from the local queue, as the request has already been removed from all other node's queues when node
		//announced that it is leaving the critical section.
		s.node.removeRequest(nextRequest.NodeID)
		s.node.ListRequestsInQueue()

		//OBS OBS OBS: Right now I am confused and can't tell why we don't just remove this request by announcing to ourselves we are leaving?
	}

}

func (s *Server) RequestAccess(ctx context.Context, accessRequest *pb.AccessRequest) (*pb.AccessRequestResponse, error) {

	senderTimestamp := accessRequest.Timestamp
	localTimestamp := s.node.timestamp

	//fmt.Println("I HAVE RECIEVED AN ACCESS REQUEST")
	//If the current node is not in the critical section, and if the requesting node has a lower time stamp, grant access, otherwise deny access.
	granted := false
	if !s.node.tryingToAccessCriticalSection && !s.node.inCriticalSection || s.node.inCriticalSection && senderTimestamp < localTimestamp || (!s.node.inCriticalSection && senderTimestamp == localTimestamp && accessRequest.NodeID > s.node.port) {
		granted = true
		//fmt.Printf("I HAVE DECIDED TO GRANT ACCESS! MY TIMESTAMP IS: %v! AM I IN THE CRITICAL SECTION? %t\n}", s.node.timestamp, s.node.inCriticalSection)
	} else {
		granted = false
		//if the above conditions are not met we queue the request until it becomes its turn to enter the critical section
		s.node.localQueue = append(s.node.localQueue, pb.AccessRequest{NodeID: accessRequest.NodeID, Timestamp: accessRequest.Timestamp})
		//fmt.Printf("I HAVE DECIDED TO DENY ACCESS! MY TIMESTAMP IS: %v! AM I IN THE CRITICAL SECTION? %t\n", s.node.timestamp, s.node.inCriticalSection)
		//fmt.Printf("I HAVE ADDED THE REQUEST TO MY LOCAL QUEUE. THE NUMBER OF REQUESTS IN MY QUEUE IS: %v\n", len(s.node.localQueue))
		//s.node.ListRequestsInQueue()
	}
	//We return a response with information on whether or not the request has been granted.
	return &pb.AccessRequestResponse{Granted: granted, Timestamp: localTimestamp}, nil
}
func (s *Server) AnnounceLeave(ctx context.Context, announcement *pb.LeaveAnnouncement) (*pb.LeaveAnnouncementResponse, error) {
	//This method is called to indicate that a node has left the critical section, which means we need to update the queue.
	//We remove from the queue the request from the node which we have just learned has left the critical section, if it is there.
	s.node.removeRequest(announcement.NodeID)
	return &pb.LeaveAnnouncementResponse{}, nil
}

// This instance of node tries to access the critical zone by sending requests and waiting for responses
func (s *Server) AccessCriticalZone(port int32) {
	// Spinner for a more interactive application
	spinner := spinner.New(spinner.CharSets[14], 100*time.Millisecond) // Build our new spinner
	for {
		//if already in the critical section, have a wait and try again.
		if s.node.inCriticalSection {
			if !spinner.Active() {
				fmt.Println("* In Critical Section... *")
				// Spinner for a more interactive application
				spinner.Start()
			}
			time.Sleep(time.Millisecond * time.Duration(500))
			continue
		}
		// We check if the request is already in the local queue. If so, we don't want to send another request.
		if s.IsAccessRequestInQueue(port) {
			if !spinner.Active() {
				fmt.Println("* Waiting... *")
				// Spinner for a more interactive application
				spinner.Start()
			}
			time.Sleep(time.Millisecond * time.Duration(500))
			continue
		}

		s.node.tryingToAccessCriticalSection = false

		randSrc := rand.NewSource(time.Now().UnixNano())
		randGen := rand.New(randSrc)
		wait := int32(randGen.Intn(5000)) // Generate a random integer and cast to int32
		time.Sleep(time.Millisecond * time.Duration(wait))
		if !spinner.Active() {
			fmt.Println(" * * * TRYING TO ENTER THE CRITICAL SECTION * * * ")
			// Spinner for a more interactive application
			spinner.Start()
		}

		s.node.tryingToAccessCriticalSection = true

		var accessGrantedCount = 0

		if !s.node.inCriticalSection {
			// Sending access requests/reponses to other peer nodes
			// (important part of the Ricart-Agrawala Algorithm)

			//We update the timestamp before we send out a request to the connected nodes.
			s.node.timestamp++
			for _, connectedNode := range s.node.connectedNodes {
				node := connectedNodesMapClient[connectedNode]

				//We don't send an access request to ourself, therefore skip node's own port.
				if node.port != s.node.port {

					accessRequestResponse, err := connectedNode.RequestAccess(context.Background(), &pb.AccessRequest{NodeID: port, Timestamp: s.node.timestamp})

					if err != nil {
						log.Fatalf("Oh no! Something went wrong while requesting access to enter critical section. The error is: %v", err)
					}

					if accessRequestResponse.Granted {
						accessGrantedCount++
					}

					s.node.timestamp++
				}
			}

			//We have now cycled through every node, sent out a request and recieved a response.
			//We check if the number of responses granting access matches the number of connected nodes (minus one, since we don't send the request to this node!)
			if accessGrantedCount == len(s.node.connectedNodes)-1 {

				if spinner.Active() {
					spinner.Stop()
				}

				//If it matches we enter (and leave) the critical section
				s.EnterCriticalSection()

				//we reset the inCriticalSection and accesGrantedCount variables.
				accessGrantedCount = 0
			} else {
				//If all nodes did not grant access, we instead need to queue our request.
				//We will then handle the request in due time, the same way we handle the other requests in the queue.
				s.node.localQueue = append(s.node.localQueue, pb.AccessRequest{NodeID: port, Timestamp: s.node.timestamp})

				//Queuing this request is already handled in some nodes by the call to RequestAccess.
				//However: If some nodes did grant access, they did not queue this request already. Therefore we cycle through all connected nodes and
				//Tell them to add this request to their queue if they didn't already.
				for _, connectedNode := range s.node.connectedNodes {
					node := connectedNodesMapClient[connectedNode]
					node.client.AnnounceQueuedRequest(context.Background(), &pb.AccessRequest{NodeID: port, Timestamp: s.node.timestamp})
				}

				if spinner.Active() {
					spinner.Stop()
				}
			}
		}
	}
}

func (s *Server) EnterCriticalSectionDirectly(ctx context.Context, accessRequest *pb.AccessRequest) (*pb.Confirmation, error) {
	//fmt.Printf("THE NODE WITH THE ID %v HAS JUST ASKED ME THE ENTER THE CRITICAL SECTION DIRECTLY!\n", accessRequest.NodeID)
	s.EnterCriticalSection()
	return &pb.Confirmation{}, nil
}

func (s *Server) EnterCriticalSection() {
	log.Printf("\n\n* * * ENTERED THE CRITICAL ZONE ON PORT %v, AT TIMESTAMP: %v * * * \n\n\n", s.node.port, s.node.timestamp)
	s.node.inCriticalSection = true
	s.node.tryingToAccessCriticalSection = false
	time.Sleep(time.Millisecond * time.Duration(5000))

	file, err := os.OpenFile("criticalsection.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	mu.Lock()

	if err != nil {
		log.Fatalf("Access to critical section denied!")
	} else {
		fmt.Println("Access to critical section granted!")
		_, err := file.WriteString("Last edit by port " + strconv.Itoa(int(s.node.port)) + " with timestamp " + strconv.Itoa(int(s.node.timestamp)) + "\n")
		if err != nil {
			log.Fatalf("Failed to modify critical section: %s", err)
		}
	}
	defer file.Close()
	defer mu.Unlock()

	//We update the timestamp ... to ensure that future requests reflect the most recent state. I am not 100 % sure this is necessary...
	s.node.timestamp++
	//We remove the request to enter the section from the queue if it is there.
	s.node.removeRequest(s.node.port)
	s.LeaveCriticalSection()
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

func FindAnAvailablePort(standardPort int) (int, error) {
	for port := standardPort; port < standardPort+100; port++ {
		addr := "localhost:" + strconv.Itoa(port)
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			//The port is in use, increment and try the next one
			continue
		}
		//if no error, the port is free. Return the port.
		listener.Close()
		return port, nil
	}
	return 0, fmt.Errorf("no free port found in the range")
}

func (s *Server) EstablishConnectionToAllOtherNodes(standardPort int, thisPort int, transportCreds credentials.TransportCredentials, connectedNodes []pb.NodeServiceClient) {
	//We cycle through the available ports in order to find the other nodes in the system and establish connections.
	for port := standardPort; port < standardPort+100; port++ {
		if port != thisPort {
			address := "localhost:" + strconv.Itoa(port)
			//fmt.Println("Hello. Checking this port: " + address)
			conn, err := grpc.Dial(address, grpc.WithTransportCredentials(transportCreds))
			if err == nil {
				//OBS OBS OBS: THE PROBLEM IS THIS: THIS METHOD DOES NOT RETURN AN ERROR JUST BECAUSE THERE IS NO NODE ON THE PORT ...
				//SO THE CODE THINKS ITS FOUND A NEW NODE EVEN WHEN THERE IS NONE. AND THE PROGRAM ONLY CRASHES ONCE
				//THE CODE CALLS THE ANNOUNCECONNECTION METHOD ON THE NODE. SO HOW CAN WE CHECK IF THERE IS ACTUALLY A NODE ON THE PORT?

				//We make a node with the connection to check if there is anything there...
				node := pb.NewNodeServiceClient(conn)
				_, err1 := node.IExist(context.Background(), &pb.Confirmation{})
				if err1 != nil {
					//There is no node on this port. We move onto the next and try again.
					continue
				}
				//We didn't get an error which means that there is a node on this port. We have established a connection!
				//First we add the node to our own list of connected nodes as well as the relevant maps.....
				s.node.connectedNodes = append(s.node.connectedNodes, node)
				connectedNodes = append(connectedNodes, node)                                                                                     //OBS OBS OBS - why do we have two of these? Can we just have the connected nodes on the server? Why is it on the server anyway?
				nodeInfo := &NodeInfo{port: int32(port), client: node, connectedNodes: s.node.connectedNodes, timestamp: int32(s.node.timestamp)} //OBS OBS OBS IS THIS RIGHT REGARDING TIMESTAMP?
				connectedNodesMapPort[int32(port)] = *nodeInfo
				connectedNodesMapClient[node] = *nodeInfo

				//Then we send an announcement to inform the node
				//in order to inform it that we have connected to it (and that it should connect to this node in return.)
				_, err := node.AnnounceConnection(context.Background(), &pb.ConnectionAnnouncement{NodeID: int32(s.node.port)})
				if err != nil {
					log.Fatalf("Oh no! The node sent an announcement of a new connection but did not recieve a confirmation in return. Error: %v", err)
				}
				//fmt.Println("SENDING THE ANNOUNCEMENT SEEMS TO HAVE GONE OK?")
			}
		}
	}
}

// Function to check if a request is already in the local queue
func (s *Server) IsAccessRequestInQueue(nodeID int32) bool {
	for _, req := range s.node.localQueue {
		if req.NodeID == nodeID {
			return true
		}
	}
	return false
}

func (n *NodeInfo) ListRequestsInQueue() {
	fmt.Println("ELEMENTS IN MY QUEUE: ")
	i := 1
	for _, req := range n.localQueue {
		fmt.Printf(" - %v: Node with port %v", i, req.NodeID)
		i++
	}
	fmt.Print("\n")
}

func main() {

	timestamp := 0
	connectedNodes := []pb.NodeServiceClient{}

	//First we need to establish connection to the other nodes in the system.
	//In order to enble us to make remote procedure calls to other nodes with grpc
	//For the purposes of this excersize we decided to simply configure the IP addresses and ports of the nodes manually.
	//To do this, we just specify a standard port and increment the port if the port is already in use and repeat until we find a free port.
	//It might have been better to do some more dynamic node discovery system.

	//we need to establish connection
	//First we find a port
	standardPort := 8000
	port, err := FindAnAvailablePort(standardPort)
	fmt.Printf("THE PORT IS: %v\n ", port)
	if err != nil {
		log.Fatalf("Oh no! Failed to find a port")
	}

	// Create a gRPC server
	grpcServer := grpc.NewServer()
	server := Server{}
	// Register your gRPC service with the server
	pb.RegisterNodeServiceServer(grpcServer, &server)

	//initialize the listener on the specified port. net.Listen listens for incoming connections with tcp socket
	listen, err := net.Listen("tcp", "localhost:"+strconv.Itoa(port))
	if err != nil {
		log.Fatalf("Could not listen at port: %d : %v", port, err)
	}
	go func() {
		// Start gRPC server in a goroutine
		err := grpcServer.Serve(listen)
		if err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	//When a port is found, next we need to establish connections to the existing nodes in the system.
	//because the ports and address are configured manually in this solution we already know the range
	//in which we expect to find those other nodes. We cycle through them

	//we create insecure transport credentials (in the context of this assignment we choose not to worry about security):
	transportCreds := insecure.NewCredentials()

	//Establish a grpc connection to the other node using addres and transport credentials
	address := ":" + strconv.Itoa(port)
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(transportCreds))
	if err != nil {
		log.Fatalf("Failed to connect  ... : %v\n", err)
	}

	//Create a grpc client instance to represent local node (this node).
	//In a way we are establishing connection to our own node ... which may seem weird, but it does make sense:
	//The idea is to have a local representation of this node, that can interact with other nodes in the system,
	//since it is a client that can call remote procedures and recieve remote procedure calls through grpc.
	thisNodeClient := pb.NewNodeServiceClient(conn)

	//Add the node we have connected to our list of nodes in the system.
	//We also maintain a map which lets us find the node from its NodeID.
	//Add node to list of connected nodes and make NodeInfo node.
	//We also need to add our own node to the list of connected nodes ... this makes some of the other logic easier even if it seems odd
	connectedNodes = append(connectedNodes, thisNodeClient)

	// Generate node
	node := &NodeInfo{port: int32(port), client: thisNodeClient, connectedNodes: connectedNodes, timestamp: int32(timestamp)}
	connectedNodesMapPort[int32(port)] = *node
	connectedNodesMapClient[thisNodeClient] = *node
	server.node = *node

	//Now we want to establish connection to all other nodes in the system.
	server.EstablishConnectionToAllOtherNodes(standardPort, port, transportCreds, connectedNodes)
	//We have now established connection to all other nodes in the system, notified them, and they have established connections back in return.

	//Next: We try to inter the critical section
	go server.AccessCriticalZone(int32(port))

	select {}
}
