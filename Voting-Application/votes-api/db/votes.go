package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/nitishm/go-rejson/v4"
)
  
type Vote struct {
	VoteID		uint
	VoterID		uint
	PollID		uint
	VoteValue	uint
	Links		[]string
}

const (
	RedisNilError        = "redis: nil"
	RedisDefaultLocation = "0.0.0.0:6379"
	RedisKeyPrefix       = "votes:"
)

type cache struct {
	cacheClient *redis.Client
	jsonHelper  *rejson.Handler
	context     context.Context
}

type healthData struct{
	Uptime time.Duration
	APIcalls uint
}

type VoteList struct {
	healthInfo healthData
	cache
}

//constructor for VoteList struct
func NewVoteList() (*VoteList, error) {
	//We will use an override if the REDIS_URL is provided as an environment
	//variable, which is the preferred way to wire up a docker container
	redisUrl := os.Getenv("REDIS_URL")
	//This handles the default condition
	if redisUrl == "" {
		redisUrl = RedisDefaultLocation
	}
	return NewWithCacheInstance(redisUrl)
}

// NewWithCacheInstance is a constructor function that returns a pointer to a new
// Vote struct.  It accepts a string that represents the location of the redis
// cache.
func NewWithCacheInstance(location string) (*VoteList, error) {

	//Connect to redis.  Other options can be provided, but the
	//defaults are OK
	client := redis.NewClient(&redis.Options{
		Addr: location,
	})

	//We use this context to coordinate betwen our go code and
	//the redis operaitons
	ctx := context.Background()

	//This is the reccomended way to ensure that our redis connection
	//is working
	err := client.Ping(ctx).Err()
	if err != nil {
		log.Println("Error connecting to redis" + err.Error())
		return nil, err
	}

	//By default, redis manages keys and values, where the values
	//are either strings, sets, maps, etc.  Redis has an extension
	//module called ReJSON that allows us to store JSON objects
	//however, we need a companion library in order to work with it
	//Below we create an instance of the JSON helper and associate
	//it with our redis connnection
	jsonHelper := rejson.NewReJSONHandler()
	jsonHelper.SetGoRedisClientWithContext(ctx, client)

	//Return a pointer to a new voteList struct
	voteList := &VoteList{
		healthInfo: healthData{},
		cache: cache{
			cacheClient: client,
			jsonHelper:  jsonHelper,
			context:     ctx,
		},
	}
	return voteList, nil
}

//------------------------------------------------------------
// REDIS HELPERS
//------------------------------------------------------------

// In redis, our keys will be strings, they will look like
// votes:<number>.  This function will take an integer and
// return a string that can be used as a key in redis
func redisKeyFromId(id uint) string {
	return fmt.Sprintf("%s%d", RedisKeyPrefix, id)
}

// Helper to return a VoteList from redis provided a key
func (v *VoteList) getItemFromRedis(key string, vote *Vote) error {

	//Lets query redis for the vote, note we can return parts of the
	//json structure, the second parameter "." means return the entire
	//json structure
	voteObject, err := v.jsonHelper.JSONGet(key, ".")
	if err != nil {
		return err
	}

	//JSONGet returns an "any" object, or empty interface,
	//we need to convert it to a byte array, which is the
	//underlying type of the object, then we can unmarshal
	//it into our voter struct
	err = json.Unmarshal(voteObject.([]byte), vote)
	if err != nil {
		return err
	}

	return nil
}

//------------------------------------------------------------
// THESE ARE THE PUBLIC FUNCTIONS THAT SUPPORT OUR VOTE APP
//------------------------------------------------------------

// AddVote accepts a Vote and adds it to the DB.
// Preconditions:   (1) The database file must exist and be a valid
//
//					(2) The vote must not already exist in the DB
//	    				because we use the vote.VoteID as the key, this
//						function must check if the vote already
//	    				exists in the DB, if so, return an error
//
// Postconditions:
//
//	    (1) The vote will be added to the DB
//		(2) The DB file will be saved with the vote added
//		(3) If there is an error, it will be returned
func (v *VoteList) AddVote(vote Vote) error {

	//Before we add an vote to the DB, lets make sure
	//it does not exist, if it does, return an error
	redisKey := redisKeyFromId(vote.VoteID)
	var existingVote Vote
	if err := v.getItemFromRedis(redisKey, &existingVote); err == nil {
		return errors.New("vote already exists")
	}
	var checkVoter Vote
	if err := v.getItemFromRedis(fmt.Sprintf("%s%d", "voters:", vote.VoterID), &checkVoter); err != nil {
		return errors.New("voter does not exists")
	}
	var checkPoll Vote
	if err := v.getItemFromRedis(fmt.Sprintf("%s%d", "polls:", vote.PollID), &checkPoll); err != nil {
		return errors.New("poll does not exists")
	}

	//Add vote to database with JSON Set
	vote.Links = []string{"GET All Votes: 1100/votes/", "POST Vote: 1100/votes/:id", "DELETE All Votes: 1100/votes", "DELETE Vote: 1100/votes/:id","GET All Voters: 1080/voters/","POST Voter: 1080/voters/:id","GET All Polls: 1090/Polls/","POST Poll: 1090/polls/:id"}
	if _, err := v.jsonHelper.JSONSet(redisKey, ".", vote); err != nil {
		return err
	}

	//If everything is ok, return nil for the error
	return nil
}

// DeleteVote accepts a vote id and removes it from the DB.
// Preconditions:   (1) The database file must exist and be a valid
//
//					(2) The vote must exist in the DB
//	    				because we use the vote.VoteID as the key, this
//						function must check if the vote already
//	    				exists in the DB, if not, return an error
//
// Postconditions:
//
//	    (1) The vote will be removed from the DB
//		(2) The DB file will be saved with the vote removed
//		(3) If there is an error, it will be returned
func (v *VoteList) DeleteVote(id uint) error {

	pattern := redisKeyFromId(id)
	numDeleted, err := v.cacheClient.Del(v.context, pattern).Result()
	if err != nil {
		return err
	}
	if numDeleted == 0 {
		return errors.New("vote does not exist")
	}

	return nil
}

// DeleteAllVotes removes all votes from the DB.
// It will be exposed via a DELETE /votes endpoint
func (v *VoteList) DeleteAllVotes() error {

	pattern := RedisKeyPrefix + "*"
	ks, _ := v.cacheClient.Keys(v.context, pattern).Result()
	//Note delete can take a collection of keys.  In go we can
	//expand a slice into individual arguments by using the ...
	//operator
	numDeleted, err := v.cacheClient.Del(v.context, ks...).Result()
	if err != nil {
		return err
	}

	if numDeleted != int64(len(ks)) {
		return errors.New("one or more votes could not be deleted")
	}

	return nil
}

// UpdateVote accepts a Vote and updates it in the DB.
// Preconditions:   (1) The database file must exist and be a valid
//
//					(2) The vote must exist in the DB
//	    				because we use the vote.VoteID as the key, this
//						function must check if the vote already
//	    				exists in the DB, if not, return an error
//
// Postconditions:
//
//	    (1) The vote will be updated in the DB
//		(2) The DB file will be saved with the vote updated
//		(3) If there is an error, it will be returned
func (v *VoteList) UpdateVote(vote Vote) error {

	// Check if vote exists before trying to update it
	// this is a good practice, return an error if the
	// vote does not exist
	redisKey := redisKeyFromId(vote.VoteID)
	var existingVote Vote
	if err := v.getItemFromRedis(redisKey, &existingVote); err != nil {
		return errors.New("vote does not exist")
	}

	//Add vote to database with JSON Set.  Note there is no update
	//functionality, so we just overwrite the existing vote
	vote.Links = []string{"GET All Votes: 1100/votes/", "POST Vote: 1100/votes/:id", "DELETE All Votes: 1100/votes", "DELETE Vote: 1100/votes/:id","GET All Voters: 1080/voters/","POST Voter: 1080/voters/:id","GET All Polls: 1090/Polls/","POST Poll: 1090/polls/:id"}
	if _, err := v.jsonHelper.JSONSet(redisKey, ".", vote); err != nil {
		return err
	}

	return nil
}

// GetVote accepts a Vote id and returns the vote from the DB.
// Preconditions:   (1) The database file must exist and be a valid
//
//					(2) The vote must exist in the DB
//	    				because we use the vote.VoteID as the key, this
//						function must check if the vote already
//	    				exists in the DB, if not, return an error
//
// Postconditions:
//
//	    (1) The vote will be returned, if it exists
//		(2) If there is an error, it will be returned
//			along with an empty vote
//		(3) The database file will not be modified
func (v *VoteList) GetVote(id uint) (Vote, error) {

	// Check if vote exists before trying to get it
	// this is a good practice, return an error if the
	// vote does not exist
	var vote Vote
	pattern := redisKeyFromId(id)
	err := v.getItemFromRedis(pattern, &vote)
	if err != nil {
		return Vote{}, errors.New("vote does not exist")
	}

	return vote, nil
}

// GetAllVotes returns all votes from the DB.  If successful it
// returns a slice of all of the votes to the caller
// Preconditions:   (1) The database file must exist and be a valid
//
// Postconditions:
//
//	    (1) All votes will be returned, if any exist
//		(2) If there is an error, it will be returned
//			along with an empty slice
//		(3) The database file will not be modified
func (v *VoteList) GetAllVotes() ([]Vote, error) {

	//Now that we have the DB loaded, lets crate a slice
	var voteList []Vote
	var vote Vote

	//Lets query redis for all of the items
	pattern := RedisKeyPrefix + "*"
	ks, _ := v.cacheClient.Keys(v.context, pattern).Result()
	for _, key := range ks {
		err := v.getItemFromRedis(key, &vote)
		if err != nil {
			return nil, err
		}
		voteList = append(voteList, vote)
	}

	if len(voteList) < 1 {
		voteList = append(voteList, Vote{
			VoteID: 0,
			VoterID: 0,
			PollID: 0,
			VoteValue: 0,
			Links: []string{"GET All Votes: 1100/votes/", "POST Vote: 1100/votes/:id", "DELETE All Votes: 1100/votes", "DELETE Vote: 1100/votes/:id","GET All Voters: 1080/voters/","POST Voter: 1080/voters/:id","GET All Polls: 1090/Polls/","POST Poll: 1090/polls/:id"},
		})
	}

	//Now that we have all of our votes in a slice, return it
	return voteList, nil
}

// PrintVote accepts a Vote and prints it to the console
// in a JSON pretty format. As some help, look at the
// json.MarshalIndent() function from our in class go tutorial.
func (v *VoteList) PrintVote(vote Vote) {
	jsonBytes, _ := json.MarshalIndent(vote, "", "  ")
	fmt.Println(string(jsonBytes))
}

// PrintAllVotes accepts a slice of Votes and prints them to the console
// in a JSON pretty format.  It should call PrintVote() to print each vote
// versus repeating the code.
func (v *VoteList) PrintAllVotes(voteList []Vote) {
	for _, vote := range voteList {
		v.PrintVote(vote)
	}
}

// JsonToVote accepts a json string and returns a Vote
// This is helpful because the CLI accepts votes for insertion
// and updates in JSON format.  We need to convert it to a Vote
// struct to perform any operations on it.
func (v *VoteList) JsonToVote(jsonString string) (Vote, error) {
	var vote Vote
	err := json.Unmarshal([]byte(jsonString), &vote)
	if err != nil {
		return Vote{}, err
	}

	return vote, nil
}

func (v *VoteList) GetHealthData(bootTime time.Time, calls uint) (healthData, error){

	v.healthInfo = healthData{Uptime: time.Now().Sub(bootTime), APIcalls: calls}

	return v.healthInfo, nil
}