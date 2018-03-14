package main

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"proj2_f5w9a_h6v9a_q7w9a_r8u8_w1c0b/serverpb"
	"strings"
	"time"

	"google.golang.org/grpc"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Not enough arguments.")
		return
	}
	// Set up RPC connection to client
	ctx, _ := context.WithTimeout(context.TODO(), 2*time.Second)
	conn, err := grpc.DialContext(ctx, os.Args[1], grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	client := serverpb.NewClientClient(conn)
	fmt.Println("\n 🌠  Welcome to the Ivan Planetary File System. Type 'help' to list all options. 🌠 \n")
	start(client, ctx)
}

func start(client serverpb.ClientClient, ctx context.Context) {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("ipfs> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		cmd := strings.Split(input, " ")
		switch cmd[0] {
		case "get":
			get(cmd, client, ctx)
		case "add":
			add(cmd, client, ctx)
		case "peers":
			peers(cmd, client, ctx)
		case "reference":
			reference(cmd, client, ctx)
		case "help":
			fmt.Println("\n 🚀  List of options: \n")
			fmt.Println("	get <file_id>		Fetch a document")
			fmt.Println("	add <path/to/file>	Add a document to this node")
			fmt.Println("	peers list		List this node's peers")
			fmt.Println("	peers add <node_id>	Add a peer to this node")
			fmt.Println("	reference get		Fetch the document that this reference points to")
			fmt.Println("	reference add <foo>	Add or update a reference\n")
		case "quit":
			fmt.Println("Exiting program... Goodbye. 🌙")
			os.Exit(1)
		default:
			fmt.Println("Invalid command. Type 'help' to list all options.")
		}
	}
}

func get(cmd []string, client serverpb.ClientClient, ctx context.Context) {
	if len(cmd) != 2 {
		fmt.Println("Incorrect number of arguments. Please specify a file ID.")
	} else {
		args := &serverpb.GetRequest{
			FileId: cmd[1],
		}
		resp, err := client.Get(ctx, args)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("File name: " + resp.File.GetName())
			fmt.Println(resp.File.GetData())
		}
	}
}

func add(cmd []string, client serverpb.ClientClient, ctx context.Context) {
	if len(cmd) != 2 {
		fmt.Println("Incorrect number of arguments. Please specify the path to the file you wish to add.")
	} else {
		data, err := ioutil.ReadFile(cmd[1])
		if err != nil {
			fmt.Println(err)
		} else {
			args := &serverpb.AddRequest{
				File: &serverpb.File{
					Data: data,
				},
			}
			resp, err := client.Add(ctx, args)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("File ID: " + resp.GetFileId())
			}
		}
	}
}

func peers(cmd []string, client serverpb.ClientClient, ctx context.Context) {
	if len(cmd) < 2 {
		fmt.Println("Incorrect number of arguments.")
	} else if cmd[1] == "list" {
		// TODO: List peers
	} else if cmd[1] == "add" && len(cmd) == 3 {
		// TODO: Add peer
	} else if cmd[1] == "add" && len(cmd) != 3 {
		fmt.Println("Please specify a peer ID.")
	} else {
		fmt.Println("Invalid command.")
	}
}

func reference(cmd []string, client serverpb.ClientClient, ctx context.Context) {
	if len(cmd) < 2 {
		fmt.Println("Incorrect number of arguments.")
	} else if cmd[1] == "get" {
		// TODO: Fetch reference
	} else if cmd[1] == "add" && len(cmd) == 3 {
		// TODO: Add reference
	} else if cmd[1] == "add" && len(cmd) != 3 {
		// TODO: WTF is a reference
		fmt.Println("Please specify a reference.")
	} else {
		fmt.Println("Invalid command.")
	}
}
