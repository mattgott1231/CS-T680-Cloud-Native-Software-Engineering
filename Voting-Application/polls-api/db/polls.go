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

type pollOption struct {
	PollOptionID    uint
	PollOptionText string
}
  
type Poll struct {
	PollID			uint
	PollTitle		string
	PollQuestion	string
	PollOptions		[]pollOption
	Links 			[]string
}

const (
	RedisNilError        = "redis: nil"
	RedisDefaultLocation = "0.0.0.0:6379"
	RedisKeyPrefix       = "polls:"
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

type PollList struct {
	healthInfo healthData
	cache
}

//constructor for PollList struct
func NewPollList() (*PollList, error) {
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
// Poll struct.  It accepts a string that represents the location of the redis
// cache.
func NewWithCacheInstance(location string) (*PollList, error) {

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

	//Return a pointer to a new voterList struct
	pollList := &PollList{
		healthInfo: healthData{},
		cache: cache{
			cacheClient: client,
			jsonHelper:  jsonHelper,
			context:     ctx,
		},
	}
	return pollList, nil
}

//------------------------------------------------------------
// REDIS HELPERS
//------------------------------------------------------------

// In redis, our keys will be strings, they will look like
// polls:<number>.  This function will take an integer and
// return a string that can be used as a key in redis
func redisKeyFromId(id uint) string {
	return fmt.Sprintf("%s%d", RedisKeyPrefix, id)
}

// Helper to return a VoterList from redis provided a key
func (p *PollList) getItemFromRedis(key string, poll *Poll) error {

	//Lets query redis for the poll, note we can return parts of the
	//json structure, the second parameter "." means return the entire
	//json structure
	pollObject, err := p.jsonHelper.JSONGet(key, ".")
	if err != nil {
		return err
	}

	//JSONGet returns an "any" object, or empty interface,
	//we need to convert it to a byte array, which is the
	//underlying type of the object, then we can unmarshal
	//it into our voter struct
	err = json.Unmarshal(pollObject.([]byte), poll)
	if err != nil {
		return err
	}

	return nil
}

//------------------------------------------------------------
// THESE ARE THE PUBLIC FUNCTIONS THAT SUPPORT OUR POLL APP
//------------------------------------------------------------

// AddPoll accepts a Poll and adds it to the DB.
// Preconditions:   (1) The database file must exist and be a valid
//
//					(2) The poll must not already exist in the DB
//	    				because we use the poll.PollID as the key, this
//						function must check if the poll already
//	    				exists in the DB, if so, return an error
//
// Postconditions:
//
//	    (1) The poll will be added to the DB
//		(2) The DB file will be saved with the poll added
//		(3) If there is an error, it will be returned
func (p *PollList) AddPoll(poll Poll) error {

	//Before we add an poll to the DB, lets make sure
	//it does not exist, if it does, return an error
	redisKey := redisKeyFromId(poll.PollID)
	var existingPoll Poll
	if err := p.getItemFromRedis(redisKey, &existingPoll); err == nil {
		return errors.New("poll already exists")
	}

	//Add poll to database with JSON Set
	poll.Links = []string{"GET All Polls: 1090/polls/", "POST Poll: 1090/polls/:id", "DELETE All Polls: 1090/polls", "DELETE Poll: 1090/polls/:id","GET All Votes: 1100/votes/","POST Vote: 1100/votes/:id","GET All Voters: 1080/voters/","POST Voter: 1080/voters/:id"}
	if _, err := p.jsonHelper.JSONSet(redisKey, ".", poll); err != nil {
		return err
	}

	//If everything is ok, return nil for the error
	return nil
}

// DeletePoll accepts a poll id and removes it from the DB.
// Preconditions:   (1) The database file must exist and be a valid
//
//					(2) The poll must exist in the DB
//	    				because we use the poll.PollID as the key, this
//						function must check if the poll already
//	    				exists in the DB, if not, return an error
//
// Postconditions:
//
//	    (1) The poll will be removed from the DB
//		(2) The DB file will be saved with the poll removed
//		(3) If there is an error, it will be returned
func (p *PollList) DeletePoll(id uint) error {

	pattern := redisKeyFromId(id)
	numDeleted, err := p.cacheClient.Del(p.context, pattern).Result()
	if err != nil {
		return err
	}
	if numDeleted == 0 {
		return errors.New("poll does not exist")
	}

	return nil
}

// DeleteAllPolls removes all polls from the DB.
// It will be exposed via a DELETE /polls endpoint
func (p *PollList) DeleteAllPolls() error {

	pattern := RedisKeyPrefix + "*"
	ks, _ := p.cacheClient.Keys(p.context, pattern).Result()
	//Note delete can take a collection of keys.  In go we can
	//expand a slice into individual arguments by using the ...
	//operator
	numDeleted, err := p.cacheClient.Del(p.context, ks...).Result()
	if err != nil {
		return err
	}

	if numDeleted != int64(len(ks)) {
		return errors.New("one or more polls could not be deleted")
	}

	return nil
}

// UpdatePoll accepts a Poll and updates it in the DB.
// Preconditions:   (1) The database file must exist and be a valid
//
//					(2) The poll must exist in the DB
//	    				because we use the poll.PollID as the key, this
//						function must check if the poll already
//	    				exists in the DB, if not, return an error
//
// Postconditions:
//
//	    (1) The poll will be updated in the DB
//		(2) The DB file will be saved with the poll updated
//		(3) If there is an error, it will be returned
func (p *PollList) UpdatePoll(poll Poll) error {

	// Check if poll exists before trying to update it
	// this is a good practice, return an error if the
	// poll does not exist
	redisKey := redisKeyFromId(poll.PollID)
	var existingPoll Poll
	if err := p.getItemFromRedis(redisKey, &existingPoll); err != nil {
		return errors.New("poll does not exist")
	}

	//Add poll to database with JSON Set.  Note there is no update
	//functionality, so we just overwrite the existing poll
	poll.Links = []string{"GET All Polls: 1090/polls/", "POST Poll: 1090/polls/:id", "DELETE All Polls: 1090/polls", "DELETE Poll: 1090/polls/:id","GET All Votes: 1100/votes/","POST Vote: 1100/votes/:id","GET All Voters: 1080/voters/","POST Voter: 1080/voters/:id"}
	if _, err := p.jsonHelper.JSONSet(redisKey, ".", poll); err != nil {
		return err
	}

	return nil
}

// GetPoll accepts a poll id and returns the poll from the DB.
// Preconditions:   (1) The database file must exist and be a valid
//
//					(2) The poll must exist in the DB
//	    				because we use the poll.PollID as the key, this
//						function must check if the poll already
//	    				exists in the DB, if not, return an error
//
// Postconditions:
//
//	    (1) The poll will be returned, if it exists
//		(2) If there is an error, it will be returned
//			along with an empty Poll
//		(3) The database file will not be modified
func (p *PollList) GetPoll(id uint) (Poll, error) {

	// Check if poll exists before trying to get it
	// this is a good practice, return an error if the
	// poll does not exist
	var poll Poll
	pattern := redisKeyFromId(id)
	err := p.getItemFromRedis(pattern, &poll)
	if err != nil {
		return Poll{}, errors.New("poll does not exist")
	}

	return poll, nil
}

// GetAllPolls returns all polls from the DB.  If successful it
// returns a slice of all of the polls to the caller
// Preconditions:   (1) The database file must exist and be a valid
//
// Postconditions:
//
//	    (1) All polls will be returned, if any exist
//		(2) If there is an error, it will be returned
//			along with an empty slice
//		(3) The database file will not be modified
func (p *PollList) GetAllPolls() ([]Poll, error) {

	//Now that we have the DB loaded, lets crate a slice
	var pollList []Poll
	var poll Poll

	//Lets query redis for all of the items
	pattern := RedisKeyPrefix + "*"
	ks, _ := p.cacheClient.Keys(p.context, pattern).Result()
	for _, key := range ks {
		err := p.getItemFromRedis(key, &poll)
		if err != nil {
			return nil, err
		}
		pollList = append(pollList, poll)
	}

	if len(pollList) < 1 {
		pollList = append(pollList, Poll{
			PollID: 0,
			PollTitle: "",
			PollQuestion: "",
			PollOptions: []pollOption{},
			Links: []string{"GET All Polls: 1090/polls/", "POST Poll: 1090/polls/:id", "DELETE All Polls: 1090/polls", "DELETE Poll: 1090/polls/:id","GET All Votes: 1100/votes/","POST Vote: 1100/votes/:id","GET All Voters: 1080/voters/","POST Voter: 1080/voters/:id"},
		})
	}

	//Now that we have all of our polls in a slice, return it
	return pollList, nil
}

// PrintPoll accepts a Poll and prints it to the console
// in a JSON pretty format. As some help, look at the
// json.MarshalIndent() function from our in class go tutorial.
func (p *PollList) PrintPoll(poll Poll) {
	jsonBytes, _ := json.MarshalIndent(poll, "", "  ")
	fmt.Println(string(jsonBytes))
}

// PrintAllPolls accepts a slice of Polls and prints them to the console
// in a JSON pretty format.  It should call PrintPoll() to print each poll
// versus repeating the code.
func (p *PollList) PrintAllPolls(pollList []Poll) {
	for _, poll := range pollList {
		p.PrintPoll(poll)
	}
}

// JsonToPoll accepts a json string and returns a Poll
// This is helpful because the CLI accepts polls for insertion
// and updates in JSON format.  We need to convert it to a Poll
// struct to perform any operations on it.
func (p *PollList) JsonToPoll(jsonString string) (Poll, error) {
	var poll Poll
	err := json.Unmarshal([]byte(jsonString), &poll)
	if err != nil {
		return Poll{}, err
	}

	return poll, nil
}

func (p *PollList) GetHealthData(bootTime time.Time, calls uint) (healthData, error){

	p.healthInfo = healthData{Uptime: time.Now().Sub(bootTime), APIcalls: calls}

	return p.healthInfo, nil
}