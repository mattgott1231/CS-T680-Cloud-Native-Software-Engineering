package api

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"drexel.edu/polls/db"
	"github.com/gin-gonic/gin"
)

// The api package creates and maintains a reference to the data handler
// this is a good design practice
type PollsAPI struct {
	db *db.PollList
}

var bootTime time.Time
var calls uint

func New() (*PollsAPI, error) {
	dbHandler, err := db.NewPollList()
	if err != nil {
		return nil, err
	}

	bootTime = time.Now()

	return &PollsAPI{db: dbHandler}, nil
}

type PollRequest struct {
	PollID			uint	`json:"PollID"`
	PollTitle		string	`json:"Polltitle"`
	PollQuestion	string	`json:"PollQuestion"`
	PollOptions		string	`json:"PollOptions"`
}

// implementation for GET /polls
// returns all polls
func (pa *PollsAPI) ListAllPolls(c *gin.Context) {

	pollList, err := pa.db.GetAllPolls()
	if err != nil {
		log.Println("Error Getting All Polls: ", err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	//Note that the database returns a nil slice if there are no items
	//in the database.  We need to convert this to an empty slice
	//so that the JSON marshalling works correctly.  We want to return
	//an empty slice, not a nil slice. This will result in the json being []
	if pollList == nil {
		pollList = make([]db.Poll, 0)
	}

	calls = calls + 1
	c.JSON(http.StatusOK, pollList)
}

// implementation for GET /polls/:id
// returns a single poll
func (pa *PollsAPI) GetPoll(c *gin.Context) {

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
		log.Println("PollID needs to be a positive value")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	poll, err := pa.db.GetPoll(numAsUint)
	if err != nil {
		log.Println("Poll not found: ", err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	calls = calls + 1
	//Git will automatically convert the struct to JSON
	//and set the content-type header to application/json
	c.JSON(http.StatusOK, poll)
}

// implementation for GET /crash
// This simulates a crash to show some of the benefits of the
// gin framework
func (pa *PollsAPI) CrashSim(c *gin.Context) {
	//panic() is go's version of throwing an exception
	panic("Simulating an unexpected crash")
}

// implementation for POST /polls
// adds a new poll
func (pa *PollsAPI) AddPoll(c *gin.Context) {
	var poll db.Poll

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
	if err := c.ShouldBindJSON(&poll); err != nil {
		log.Println("Error binding JSON: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := pa.db.AddPoll(poll); err != nil {
		log.Println("Error adding poll: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	calls = calls + 1
	c.JSON(http.StatusOK, poll)
}

// implementation for PUT /polls
// Web api standards use PUT for Updates
func (pa *PollsAPI) UpdatePoll(c *gin.Context) {
	var poll db.Poll
	if err := c.ShouldBindJSON(&poll); err != nil {
		log.Println("Error binding JSON: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := pa.db.UpdatePoll(poll); err != nil {
		log.Println("Error updating poll: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	calls = calls + 1
	c.JSON(http.StatusOK, poll)
}

// implementation for DELETE /polls/:id
// deletes a poll
func (pa *PollsAPI) DeletePoll(c *gin.Context) {
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
		log.Println("PollID needs to be a positive value")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := pa.db.DeletePoll(numAsUint); err != nil {
		log.Println("Error deleting poll: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	calls = calls + 1
	c.Status(http.StatusOK)
}

// implementation for DELETE /polls
// deletes all polls
func (pa *PollsAPI) DeleteAllPolls(c *gin.Context) {

	if err := pa.db.DeleteAllPolls(); err != nil {
		log.Println("Error deleting all polls: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	calls = calls + 1
	c.Status(http.StatusOK)
}

// implementation for GET /polls/health
// returns a "health" record indicating that the polls API is functioning properly

func (pa *PollsAPI) GetHealthData(c *gin.Context){

	healthData, err := pa.db.GetHealthData(bootTime, calls+1)
	if err != nil {
		log.Println("Error Getting health data: ", err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	
	calls = calls + 1
	c.JSON(http.StatusOK, healthData)
}