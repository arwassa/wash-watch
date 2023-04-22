package main

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"time"

	utils "github.com/arwassa/wash-watch/pkg/common"
	"github.com/arwassa/wash-watch/pkg/washd"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestWashDaemon(client washd.WashServiceClient) []utils.Machine {

	list, err := client.ListMachines(context.Background(), &washd.MachineListRequest{})
	if err != nil {
		panic(err)
	}
	machineList := make([]utils.Machine, 0, len(list.GetMachines()))
	for _, m := range list.GetMachines() {
		machineList = append(machineList, utils.Machine{
			Name:       m.Name,
			StateField: m.Status,
		})
	}
	return machineList
}

func main() {

	watchFlag := flag.String("watch", "", "Wait til specified machines have finished")

	flag.Parse()

	var watchList []string

	if *watchFlag != "" {
		watchList = strings.Split(strings.Trim(*watchFlag, ""), ",")
	}

	conn, err := grpc.Dial("passthrough:///unix://test_sock.sock", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	defer conn.Close()

	client := washd.NewWashServiceClient(conn)

	if watchList == nil {
		for _, machine := range TestWashDaemon(client) {
			fmt.Println(machine)
		}
	}

	for {
		machines := TestWashDaemon(client)
		maxTimeLeft := time.Duration(0)
		for _, id := range watchList {
			for _, m := range machines {
				if id == m.Id() {
					fmt.Println(m)
					if m.TimeLeft() > maxTimeLeft {
						maxTimeLeft = m.TimeLeft()
					}
				}
			}
		}
		if maxTimeLeft <= 0 {
			return
		} else {

			minWait := time.Minute
			if (maxTimeLeft / 2) < minWait {
				fmt.Printf("Wait for %s\n", minWait)
				time.Sleep(minWait)
			} else {
				fmt.Printf("Wait for %s\n", maxTimeLeft/2)
				time.Sleep(maxTimeLeft / 2)
			}
		}
	}
}
