package api

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"drexel.edu/votes/db"
	"github.com/gin-gonic/gin"
)

// The api package creates and maintains a reference to the data handler
// this is a good design practice
type VotesAPI struct {
	db *db.VoteList
}

var bootTime time.Time
var calls uint

func New() (*VotesAPI, error) {
	dbHandler, err := db.NewVoteList()
	if err != nil {
		return nil, err
	}

	bootTime = time.Now()

	return &VotesAPI{db: dbHandler}, nil
}

type VoteRequest struct {
	VoteID		uint	`json:"VoteID"`
	VoterID		uint	`json:"VoterID"`
	PollID		uint	`json:"PollID"`
	VoteValue	uint	`json:"VoteValue"`
}

// implementation for GET /votes
// returns all votes
func (va *VotesAPI) ListAllVotes(c *gin.Context) {

	voteList, err := va.db.GetAllVotes()
	if err != nil {
		log.Println("Error Getting All Votes: ", err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	//Note that the database returns a nil slice if there are no items
	//in the database.  We need to convert this to an empty slice
	//so that the JSON marshalling works correctly.  We want to return
	//an empty slice, not a nil slice. This will result in the json being []
	if voteList == nil {
		voteList = make([]db.Vote, 0)
	}

	calls = calls + 1
	c.JSON(http.StatusOK, voteList)
}

// implementation for GET /votes/:id
// returns a single vote
func (va *VotesAPI) GetVote(c *gin.Context) {

	//Note go is minimalistic, so we have to get the
	//id parameter using the Param() function, and then
	//convert it to an int64 using the strconv package
	idS := c.Param("id")
	id64, err := strconv.ParseInt(idS, 10, 32)
	if err != nil {
		log.Println("Error converting id to int64: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	//Note that ParseInt always returns an int64, so we have to
	//convert it to an int before we can use it.
	num := int(id64)
	var numAsUint uint
	if num >= 0 {
		numAsUint = uint(num)
	} else {
		log.Println("VoteID needs to be a positive value")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	vote, err := va.db.GetVote(numAsUint)
	if err != nil {
		log.Println("Vote not found: ", err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	calls = calls + 1
	//Git will automatically convert the struct to JSON
	//and set the content-type header to application/json
	c.JSON(http.StatusOK, vote)
}

// implementation for GET /crash
// This simulates a crash to show some of the benefits of the
// gin framework
func (va *VotesAPI) CrashSim(c *gin.Context) {
	//panic() is go's version of throwing an exception
	panic("Simulating an unexpected crash")
}

// implementation for POST /votess
// adds a new vote
func (va *VotesAPI) AddVote(c *gin.Context) {
	var vote db.Vote

	//With HTTP based APIs, a POST request will usually
	//have a body that contains the data to be added
	//to the database.  The body is usually JSON, so
	//we need to bind the JSON to a struct that we
	//can use in our code.
	//This framework exposes the raw body via c.Request.Body
	//but it also provides a helper function ShouldBindJSON()
	//that will extract the body, convert it to JSON and
	//bind it to a struct for us.  It will also report an error
	//if the body is not JSON or if the JSON does not match
	//the struct we are binding to.
	if err := c.ShouldBindJSON(&vote); err != nil {
		log.Println("Error binding JSON: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := va.db.AddVote(vote); err != nil {
		log.Println("Error adding vote: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	calls = calls + 1
	c.JSON(http.StatusOK, vote)
}

// implementation for PUT /votes
// Web api standards use PUT for Updates
func (va *VotesAPI) UpdateVote(c *gin.Context) {
	var vote db.Vote
	if err := c.ShouldBindJSON(&vote); err != nil {
		log.Println("Error binding JSON: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := va.db.UpdateVote(vote); err != nil {
		log.Println("Error updating vote: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	calls = calls + 1
	c.JSON(http.StatusOK, vote)
}

// implementation for DELETE /votes/:id
// deletes a vote
func (va *VotesAPI) DeleteVote(c *gin.Context) {
	idS := c.Param("id")
	id64, err := strconv.ParseInt(idS, 10, 32)

	if err != nil {
		log.Println("Error converting id to int64: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	num := int(id64)
	var numAsUint uint
	if num >= 0 {
		numAsUint = uint(num)
	} else {
		log.Println("VoteID needs to be a positive value")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := va.db.DeleteVote(numAsUint); err != nil {
		log.Println("Error deleting vote: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	calls = calls + 1
	c.Status(http.StatusOK)
}

// implementation for DELETE /votes
// deletes all votes
func (va *VotesAPI) DeleteAllVotes(c *gin.Context) {

	if err := va.db.DeleteAllVotes(); err != nil {
		log.Println("Error deleting all votes: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	calls = calls + 1
	c.Status(http.StatusOK)
}

// implementation for GET /votes/health
// returns a "health" record indicating that the votes API is functioning properly

func (va *VotesAPI) GetHealthData(c *gin.Context){

	healthData, err := va.db.GetHealthData(bootTime, calls+1)
	if err != nil {
		log.Println("Error Getting health data: ", err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	
	calls = calls + 1
	c.JSON(http.StatusOK, healthData)
}