package api

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"drexel.edu/voters/db"
	"github.com/gin-gonic/gin"
)

// The api package creates and maintains a reference to the data handler
// this is a good design practice
type VotersAPI struct {
	db *db.VoterList
}

func New() (*VotersAPI, error) {
	dbHandler, err := db.NewVoterList()
	if err != nil {
		return nil, err
	}

	return &VotersAPI{db: dbHandler}, nil
}

type PollRequest struct {
	PollID   uint      `json:"PollID"`
	VoteDate time.Time `json:"VoteDate"`
}

//Below we implement the API functions.  Some of the framework
//things you will see include:
//   1) How to extract a parameter from the URL, for example
//	  the id parameter in /voters/:id
//   2) How to extract the body of a POST request
//   3) How to return JSON and a correctly formed HTTP status code
//	  for example, 200 for OK, 404 for not found, etc.  This is done
//	  using the c.JSON() function
//   4) How to return an error code and abort the request.  This is
//	  done using the c.AbortWithStatus() function

// implementation for GET /voters
// returns all voters
func (va *VotersAPI) ListAllVoters(c *gin.Context) {

	voterList, err := va.db.GetAllVoters()
	if err != nil {
		log.Println("Error Getting All Voters: ", err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	//Note that the database returns a nil slice if there are no items
	//in the database.  We need to convert this to an empty slice
	//so that the JSON marshalling works correctly.  We want to return
	//an empty slice, not a nil slice. This will result in the json being []
	if voterList == nil {
		voterList = make([]db.Voter, 0)
	}

	c.JSON(http.StatusOK, voterList)
}

// implementation for GET /voters/:id
// returns a single voter
func (va *VotersAPI) GetVoter(c *gin.Context) {

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
		log.Println("VoterID needs to be a positive value")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	voter, err := va.db.GetVoter(numAsUint)
	if err != nil {
		log.Println("Voter not found: ", err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	//Git will automatically convert the struct to JSON
	//and set the content-type header to application/json
	c.JSON(http.StatusOK, voter)
}

// implementation for GET /crash
// This simulates a crash to show some of the benefits of the
// gin framework
func (va *VotersAPI) CrashSim(c *gin.Context) {
	//panic() is go's version of throwing an exception
	panic("Simulating an unexpected crash")
}

// implementation for POST /voters
// adds a new voter
func (va *VotersAPI) AddVoter(c *gin.Context) {
	var voter db.Voter

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
	if err := c.ShouldBindJSON(&voter); err != nil {
		log.Println("Error binding JSON: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := va.db.AddVoter(voter); err != nil {
		log.Println("Error adding voter: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, voter)
}

// implementation for PUT /voters
// Web api standards use PUT for Updates
func (va *VotersAPI) UpdateVoter(c *gin.Context) {
	var voter db.Voter
	if err := c.ShouldBindJSON(&voter); err != nil {
		log.Println("Error binding JSON: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := va.db.UpdateVoter(voter); err != nil {
		log.Println("Error updating voter: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, voter)
}

// implementation for DELETE /voters/:id
// deletes a voter
func (va *VotersAPI) DeleteVoter(c *gin.Context) {
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
		log.Println("VoterID needs to be a positive value")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := va.db.DeleteVoter(numAsUint); err != nil {
		log.Println("Error deleting voter: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

// implementation for DELETE /voters
// deletes all voters
func (va *VotersAPI) DeleteAllVoters(c *gin.Context) {

	if err := va.db.DeleteAllVoters(); err != nil {
		log.Println("Error deleting all voters: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

// implementation for GET /voters/:id/polls
// gets JUST the voter history for the voter with VoterID

func (va *VotersAPI) GetVoterPolls(c *gin.Context) {
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
		log.Println("VoterID needs to be a positive value")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	voterPolls, err := va.db.GetVoterPolls(numAsUint)
	if err != nil {
		log.Println("Error deleting voter: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, voterPolls)
}

// implementation for GET /voters/:id/polls/:pollId
// Gets JUST the single voter poll data with PollID = :pollId and VoterID = :id

func (va *VotersAPI) GetVoterPollByPollId(c *gin.Context) {
	voterIdS := c.Param("id")
	voterId64, err := strconv.ParseInt(voterIdS, 10, 32)

	if err != nil {
		log.Println("Error converting voter id to int64: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	voterNum := int(voterId64)
	var voterNumAsUint uint
	if voterNum >= 0 {
		voterNumAsUint = uint(voterNum)
	} else {
		log.Println("VoterID needs to be a positive value")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	pollIdS := c.Param("pollId")
	pollId64, err := strconv.ParseInt(pollIdS, 10, 32)

	if err != nil {
		log.Println("Error converting poll id to int64: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	pollNum := int(pollId64)
	var pollNumAsUint uint
	if pollNum >= 0 {
		pollNumAsUint = uint(pollNum)
	} else {
		log.Println("PollId needs to be a positive value")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	voterPoll, err := va.db.GetVoterPollByPollId(voterNumAsUint, pollNumAsUint)
	if err != nil {
		log.Println("Error deleting voter: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, voterPoll)

}

// implementation for POST /voters/:id/polls/:pollId
// Puts JUST the single voter poll data for the voter id

func (va *VotersAPI) AddVoterPoll(c *gin.Context){
	voterIdS := c.Param("id")
	voterId64, err := strconv.ParseInt(voterIdS, 10, 32)

	if err != nil {
		log.Println("Error converting voter id to int64: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	voterNum := int(voterId64)
	var voterNumAsUint uint
	if voterNum >= 0 {
		voterNumAsUint = uint(voterNum)
	} else {
		log.Println("VoterID needs to be a positive value")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var pollRequest PollRequest
		
	if err := c.ShouldBindJSON(&pollRequest); err != nil {
		log.Println("Error binding JSON: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	newPoll := db.VoterPoll{
		PollID:   pollRequest.PollID,
		VoteDate: pollRequest.VoteDate,
	}

	if err := va.db.AddVoterPoll(voterNumAsUint, newPoll); err != nil {
		log.Println("Error adding voter: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)

}

// implementation for GET /voters/health
// returns a "health" record indicating that the voter API is functioning properly

func (va *VotersAPI) GetHealthData(c *gin.Context){

	healthData, err := va.db.GetHealthData()
	if err != nil {
		log.Println("Error Getting health data: ", err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, healthData)
}