package chatserver

import (
	context "context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type messageUnit struct {
	Pnumber           string
	Room              string
	MessageBody       string
	MessageUniqueCode int
}

type messageHandle struct {
	MQue []messageUnit
	mu   sync.Mutex
}

var messageHandleObject = messageHandle{}

type ChatServer struct {
	UnimplementedServicesServer
}

type Config struct {
	Pnumber string
	Room    string
}

type UserInfo struct {
	User []Config
}

var Users = UserInfo{}

// key : pNumber, value : stream
var StreamInfo = map[string]Services_ChatServiceServer{}

// key : room, value : []pNumber
var RoomInfo = map[string][]string{}

// key : pNumber, value : Addr
var AddrInfo = map[string]string{}

//define ChatService
func (*ChatServer) ChatService(csi Services_ChatServiceServer) error {

	var pNumber string
	var room string

	md, ok := metadata.FromIncomingContext(csi.Context())
	if ok {
		if len(md["pnumber"]) > 0 {
			pNumber = md["pnumber"][0]
		}

		if len(md["room"]) > 0 {
			room = md["room"][0]
		}
		log.Print(md)
	}

	errch := make(chan error)

	ListFeatures(pNumber, room, csi)

	// p, _ := peer.FromContext(csi.Context())

	log.Printf("All Streams : %v", StreamInfo)
	log.Printf("All Rooms : %v", RoomInfo)
	log.Printf("All Addr : %v", AddrInfo)

	// receive messages - init a go routine
	go receiveFromStream(csi, errch)

	// send messages - init a go routine
	go sendToStream(errch)

	// fmt.Println(p.Addr.String())

	return <-errch

}

func makeSliceUnique(s []string) []string {
	keys := make(map[string]struct{})
	res := make([]string, 0)
	for _, val := range s {
		if _, ok := keys[val]; ok {
			continue
		} else {
			keys[val] = struct{}{}
			res = append(res, val)
		}
	}
	return res
}

func ListFeatures(pNumber string, room string, stream Services_ChatServiceServer) error {
	// Save this stream instance in the server on a map or other suitable data structure
	// so that you can query for this stream instance later
	// This will act same like your websocket session

	RoomInfo[room] = append(RoomInfo[room], pNumber)
	StreamInfo[pNumber] = stream

	p, _ := peer.FromContext(stream.Context())

	AddrInfo[p.Addr.String()] = pNumber

	RoomInfo[room] = makeSliceUnique(RoomInfo[room])

	return nil
}

func (*ChatServer) PingPong(ctx context.Context, req *FromClient) (*FromServer, error) {
	log.Println(req)
	pNumber := req.GetPnumber()
	body := req.GetBody()

	response := &FromServer{
		Pnumber: pNumber,
		Body:    body,
	}

	return response, nil
}

//receive messages
func receiveFromStream(csi_ Services_ChatServiceServer, errch_ chan error) error {

	//implement a loop
	for {
		mssg, err := csi_.Recv()
		if err != nil {
			log.Printf("Error in receiving message from client :: %v", err)
			p, _ := peer.FromContext(csi_.Context())
			pNumber := AddrInfo[p.Addr.String()]
			delete(StreamInfo, pNumber)
			delete(AddrInfo, p.Addr.String())
			errch_ <- err
			// return err
		} else {

			messageHandleObject.mu.Lock()

			messageHandleObject.MQue = append(messageHandleObject.MQue, messageUnit{
				Pnumber:           mssg.Pnumber,
				Room:              mssg.Room,
				MessageBody:       mssg.Body,
				MessageUniqueCode: rand.Intn(1e8),
			})

			log.Printf("%v", messageHandleObject.MQue[len(messageHandleObject.MQue)-1])

			messageHandleObject.mu.Unlock()

		}

		time.Sleep(500 * time.Millisecond)
	}
}

//send message
func sendToStream(errch_ chan error) {

	fmt.Println("*******go start******")

	//implement a loop
	for {

		//loop through messages in MQue
		for {

			time.Sleep(500 * time.Millisecond)

			messageHandleObject.mu.Lock()

			if len(messageHandleObject.MQue) == 0 {
				messageHandleObject.mu.Unlock()
				break
			}

			messagePnumber := messageHandleObject.MQue[0].Pnumber
			messageRoom := messageHandleObject.MQue[0].Room
			messageBody := messageHandleObject.MQue[0].MessageBody

			messageHandleObject.mu.Unlock()

			//send message to designated client (do not send to the same client)
			// if (messageRoom == room) && (messagePnumber != pNumber) {

			pNumbers := RoomInfo[messageRoom]

			for _, pNumber := range pNumbers {
				val, exists := StreamInfo[pNumber]
				if !exists {
					continue
				}
				err := val.Send(&FromServer{Pnumber: messagePnumber, Body: messageBody, Room: messageRoom})

				if err != nil {
					errch_ <- err
				}

				messageHandleObject.mu.Lock()

				if len(messageHandleObject.MQue) > 1 {
					messageHandleObject.MQue = messageHandleObject.MQue[1:] // delete the message at index 0 after sending to receiver
				} else {
					messageHandleObject.MQue = []messageUnit{}
				}

				messageHandleObject.mu.Unlock()
			}

			// for _, csi_ := range StreamInfo {

			// 	if room != messageRoom {
			// 		continue
			// 	}

			// 	err := csi_.Send(&FromServer{Pnumber: messagePnumber, Body: messageBody, Room: messageRoom})

			// 	if err != nil {
			// 		errch_ <- err
			// 	}

			// 	messageHandleObject.mu.Lock()

			// 	if len(messageHandleObject.MQue) > 1 {
			// 		messageHandleObject.MQue = messageHandleObject.MQue[1:] // delete the message at index 0 after sending to receiver
			// 	} else {
			// 		messageHandleObject.MQue = []messageUnit{}
			// 	}

			// 	messageHandleObject.mu.Unlock()

			// }

			// }

		}

		time.Sleep(100 * time.Millisecond)
	}
}
