package main

import (
	"flag"
	"fmt"
	"os"

	"drexel.edu/voters/api"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Global variables to hold the command line flags to drive the voters CLI
// application
var (
	hostFlag string
	portFlag uint
)

func processCmdLineFlags() {

	//Note some networking lingo, some frameworks start the server on localhost
	//this is a local-only interface and is fine for testing but its not accessible
	//from other machines.  To make the server accessible from other machines, we
	//need to listen on an interface, that could be an IP address, but modern
	//cloud servers may have multiple network interfaces for scale.  With TCP/IP
	//the address 0.0.0.0 instructs the network stack to listen on all interfaces
	//We set this up as a flag so that we can overwrite it on the command line if
	//needed
	flag.StringVar(&hostFlag, "h", "0.0.0.0", "Listen on all interfaces")
	flag.UintVar(&portFlag, "p", 1080, "Default Port")

	flag.Parse()
}

// main is the entry point for our voters API application.  It processes
// the command line flags and then uses the db package to perform the
// requested operation
func main() {
	processCmdLineFlags()
	r := gin.Default()
	r.Use(cors.Default())

	apiHandler, err := api.New()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	r.GET("/voters", apiHandler.ListAllVoters)
	r.POST("/voters", apiHandler.AddVoter)
	r.PUT("/voters", apiHandler.UpdateVoter)
	r.DELETE("/voters", apiHandler.DeleteAllVoters)
	r.DELETE("/voters/:id", apiHandler.DeleteVoter)
	r.GET("/voters/:id", apiHandler.GetVoter)
	r.GET("/voters/:id/polls", apiHandler.GetVoterPolls)
	r.GET("/voters/:id/polls/:pollId", apiHandler.GetVoterPoll)
	r.POST("/voters/:id/polls", apiHandler.AddVoterPoll)
	r.DELETE("/voters/:id/polls/:pollId", apiHandler.DeleteVoterPoll)
	r.PUT("/voters/:id/polls", apiHandler.UpdateVoterPoll)
	r.GET("/voters/health", apiHandler.GetHealthData)
	r.GET("/crash", apiHandler.CrashSim)

	serverPath := fmt.Sprintf("%s:%d", hostFlag, portFlag)
	r.Run(serverPath)
}
