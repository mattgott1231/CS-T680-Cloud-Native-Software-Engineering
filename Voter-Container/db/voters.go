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

type voterPoll struct{
	PollID uint
	VoteDate time.Time
}
  
type Voter struct{
	VoterID uint
	FirstName string
	LastName string
	VoteHistory []voterPoll
}

const (
	RedisNilError        = "redis: nil"
	RedisDefaultLocation = "0.0.0.0:6379"
	RedisKeyPrefix       = "voters:"
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

type VoterList struct {
	healthInfo healthData
	cache
}

//constructor for VoterList struct
func NewVoterList() (*VoterList, error) {
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
// ToDo struct.  It accepts a string that represents the location of the redis
// cache.
func NewWithCacheInstance(location string) (*VoterList, error) {

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
	voterList := &VoterList{
		healthInfo: healthData{},
		cache: cache{
			cacheClient: client,
			jsonHelper:  jsonHelper,
			context:     ctx,
		},
	}
	return voterList, nil
}

//------------------------------------------------------------
// REDIS HELPERS
//------------------------------------------------------------

// In redis, our keys will be strings, they will look like
// voter:<number>.  This function will take an integer and
// return a string that can be used as a key in redis
func redisKeyFromId(id uint) string {
	return fmt.Sprintf("%s%d", RedisKeyPrefix, id)
}

// Helper to return a VoterList from redis provided a key
func (v *VoterList) getItemFromRedis(key string, voter *Voter) error {

	//Lets query redis for the voter, note we can return parts of the
	//json structure, the second parameter "." means return the entire
	//json structure
	voterObject, err := v.jsonHelper.JSONGet(key, ".")
	if err != nil {
		return err
	}

	//JSONGet returns an "any" object, or empty interface,
	//we need to convert it to a byte array, which is the
	//underlying type of the object, then we can unmarshal
	//it into our voter struct
	err = json.Unmarshal(voterObject.([]byte), voter)
	if err != nil {
		return err
	}

	return nil
}

//------------------------------------------------------------
// THESE ARE THE PUBLIC FUNCTIONS THAT SUPPORT OUR VOTER APP
//------------------------------------------------------------

// AddVoter accepts a Voter and adds it to the DB.
// Preconditions:   (1) The database file must exist and be a valid
//
//					(2) The voter must not already exist in the DB
//	    				because we use the voter.VoterId as the key, this
//						function must check if the voter already
//	    				exists in the DB, if so, return an error
//
// Postconditions:
//
//	    (1) The voter will be added to the DB
//		(2) The DB file will be saved with the voter added
//		(3) If there is an error, it will be returned
func (v *VoterList) AddVoter(voter Voter) error {

	//Before we add an voter to the DB, lets make sure
	//it does not exist, if it does, return an error
	redisKey := redisKeyFromId(voter.VoterID)
	var existingVoter Voter
	if err := v.getItemFromRedis(redisKey, &existingVoter); err == nil {
		return errors.New("voter already exists")
	}

	//Add voter to database with JSON Set
	if _, err := v.jsonHelper.JSONSet(redisKey, ".", voter); err != nil {
		return err
	}

	//If everything is ok, return nil for the error
	return nil
}

// DeleteVoter accepts a voter id and removes it from the DB.
// Preconditions:   (1) The database file must exist and be a valid
//
//					(2) The voter must exist in the DB
//	    				because we use the voter.VoterId as the key, this
//						function must check if the voter already
//	    				exists in the DB, if not, return an error
//
// Postconditions:
//
//	    (1) The voter will be removed from the DB
//		(2) The DB file will be saved with the voter removed
//		(3) If there is an error, it will be returned
func (v *VoterList) DeleteVoter(id uint) error {

	pattern := redisKeyFromId(id)
	numDeleted, err := v.cacheClient.Del(v.context, pattern).Result()
	if err != nil {
		return err
	}
	if numDeleted == 0 {
		return errors.New("voter does not exist")
	}

	return nil
}

// DeleteAllVoters removes all voters from the DB.
// It will be exposed via a DELETE /voters endpoint
func (v *VoterList) DeleteAllVoters() error {

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
		return errors.New("one or more voters could not be deleted")
	}

	return nil
}

// UpdateVoter accepts a voter and updates it in the DB.
// Preconditions:   (1) The database file must exist and be a valid
//
//					(2) The voter must exist in the DB
//	    				because we use the voter.VoterId as the key, this
//						function must check if the voter already
//	    				exists in the DB, if not, return an error
//
// Postconditions:
//
//	    (1) The voter will be updated in the DB
//		(2) The DB file will be saved with the voter updated
//		(3) If there is an error, it will be returned
func (v *VoterList) UpdateVoter(voter Voter) error {

	// Check if voter exists before trying to update it
	// this is a good practice, return an error if the
	// voter does not exist
	redisKey := redisKeyFromId(voter.VoterID)
	var existingVoter Voter
	if err := v.getItemFromRedis(redisKey, &existingVoter); err != nil {
		return errors.New("voter does not exist")
	}

	//Add voter to database with JSON Set.  Note there is no update
	//functionality, so we just overwrite the existing voter
	if _, err := v.jsonHelper.JSONSet(redisKey, ".", voter); err != nil {
		return err
	}

	return nil
}

// GetVoter accepts a voter id and returns the voter from the DB.
// Preconditions:   (1) The database file must exist and be a valid
//
//					(2) The voter must exist in the DB
//	    				because we use the voter.VoterId as the key, this
//						function must check if the voter already
//	    				exists in the DB, if not, return an error
//
// Postconditions:
//
//	    (1) The voter will be returned, if it exists
//		(2) If there is an error, it will be returned
//			along with an empty Voter
//		(3) The database file will not be modified
func (v *VoterList) GetVoter(id uint) (Voter, error) {

	// Check if voter exists before trying to get it
	// this is a good practice, return an error if the
	// voter does not exist
	var voter Voter
	pattern := redisKeyFromId(id)
	err := v.getItemFromRedis(pattern, &voter)
	if err != nil {
		return Voter{}, errors.New("voter does not exist")
	}

	return voter, nil
}

// GetAllVoters returns all voters from the DB.  If successful it
// returns a slice of all of the voters to the caller
// Preconditions:   (1) The database file must exist and be a valid
//
// Postconditions:
//
//	    (1) All voters will be returned, if any exist
//		(2) If there is an error, it will be returned
//			along with an empty slice
//		(3) The database file will not be modified
func (v *VoterList) GetAllVoters() ([]Voter, error) {

	//Now that we have the DB loaded, lets crate a slice
	var voterList []Voter
	var voter Voter

	//Lets query redis for all of the items
	pattern := RedisKeyPrefix + "*"
	ks, _ := v.cacheClient.Keys(v.context, pattern).Result()
	for _, key := range ks {
		err := v.getItemFromRedis(key, &voter)
		if err != nil {
			return nil, err
		}
		voterList = append(voterList, voter)
	}

	//Now that we have all of our voters in a slice, return it
	return voterList, nil
}

// PrintVoter accepts a Voter and prints it to the console
// in a JSON pretty format. As some help, look at the
// json.MarshalIndent() function from our in class go tutorial.
func (v *VoterList) PrintVoter(voter Voter) {
	jsonBytes, _ := json.MarshalIndent(voter, "", "  ")
	fmt.Println(string(jsonBytes))
}

// PrintAllVoters accepts a slice of Voters and prints them to the console
// in a JSON pretty format.  It should call PrintVoter() to print each voter
// versus repeating the code.
func (v *VoterList) PrintAllVoters(voterList []Voter) {
	for _, voter := range voterList {
		v.PrintVoter(voter)
	}
}

// JsonToVoter accepts a json string and returns a Voter
// This is helpful because the CLI accepts voters for insertion
// and updates in JSON format.  We need to convert it to a Voter
// struct to perform any operations on it.
func (v *VoterList) JsonToVoter(jsonString string) (Voter, error) {
	var voter Voter
	err := json.Unmarshal([]byte(jsonString), &voter)
	if err != nil {
		return Voter{}, err
	}

	return voter, nil
}

// GetVoterPolls accepts a voter id and returns polls from that voter.
// Preconditions:   (1) The database file must exist and be a valid
//
//					(2) The voter must exist in the DB
//	    				because we use the voter.VoterId as the key, this
//						function must check if the voter already
//	    				exists in the DB, if not, return an error
//
// Postconditions:
//
//      (1) All polls will be returned, if any exist
//		(2) If there is an error, it will be returned
//			along with an empty slice
//		(3) The database file will not be modified
func (v *VoterList) GetVoterPolls(id uint) ([]voterPoll, error) {

	// we should check if voter exists before trying to retriece polls
	// this is a good practice, return an error if the
	// voter does not exist

	var voter Voter
	pattern := redisKeyFromId(id)
	err := v.getItemFromRedis(pattern, &voter)
	if err != nil {
		return nil, errors.New("voter does not exist")
	}

	return voter.VoteHistory, nil
}


// GetVoterPoll accepts a voter id and poll id and returns the requested poll.
// Preconditions:   (1) The database file must exist and be a valid
//
//					(2) The voter and poll must exist in the DB
//	    				because we use the voter.VoterId as the key, this
//						function must check if the voter already
//	    				exists in the DB, if not, return an error
//
// Postconditions:
//
//	    (1) The poll will be returned, if it exists
//		(2) If there is an error, it will be returned
//			along with an empty poll
//		(3) The database file will not be modified
func (v *VoterList) GetVoterPoll(voterId, pollId uint) (voterPoll , error) {

    // we should if voter exists before trying to retrieve polls
    // this is a good practice, return an error if the
    // voter does not exist

	var voter Voter
	pattern := redisKeyFromId(voterId)
	err := v.getItemFromRedis(pattern, &voter)
	if err != nil {
		return voterPoll{}, errors.New("voter does not exist")
	}

    for _, poll := range voter.VoteHistory {
        if poll.PollID == pollId{
			return poll, nil
        }
    }

    return voterPoll{}, errors.New("poll not found for given voter")
}

// AddVoterPoll accepts a voter id and new poll to add to the voter.
// Preconditions:   (1) The database file must exist and be a valid
//
//					(2) The voter must exist in the DB
//	    				because we use the voter.VoterId as the key, this
//						function must check if the voter already
//	    				exists in the DB, if not, return an error
//
// Postconditions:
//
//	    (1) The poll will be added to the DB
//		(2) The DB file will be saved with the poll added
//		(3) If there is an error, it will be returned
func (v *VoterList) AddVoterPoll(voterId uint, requestVoter Voter) error {

	// Check if voter exists before trying to update it
	// this is a good practice, return an error if the
	// voter does not exist

	var voter Voter
	pattern := redisKeyFromId(voterId)
	err := v.getItemFromRedis(pattern, &voter)
	if err != nil {
		return errors.New("voter does not exist")
	}
	
	requestPoll := requestVoter.VoteHistory[0]

	for _, poll := range voter.VoteHistory {
        if poll.PollID == requestPoll.PollID{
			return errors.New("poll already exists in voter")
        }
    }

	voter.VoteHistory = append(voter.VoteHistory, requestPoll)
	v.UpdateVoter(voter)

	return nil
}

// DeleteVoterPoll accepts a voter id and a poll to add to the voter.
// Preconditions:   (1) The database file must exist and be a valid
//
//					(2) The voter must exist in the DB
//	    				because we use the voter.VoterId as the key, this
//						function must check if the voter already
//	    				exists in the DB, if not, return an error
//
// Postconditions:
//
//	    (1) The poll will be deleted from the DB
//		(2) The DB file will be saved with the poll deleted
//		(3) If there is an error, it will be returned
func (v *VoterList) DeleteVoterPoll(voterId uint, pollId uint) error {

	// Check if voter exists before trying to update it
	// this is a good practice, return an error if the
	// voter does not exist

	var voter Voter
	pattern := redisKeyFromId(voterId)
	err := v.getItemFromRedis(pattern, &voter)
	if err != nil {
		return errors.New("voter does not exist")
	}

	index := -1
    for i, poll := range voter.VoteHistory {
        if poll.PollID == pollId{
            index = i
            break
        }
    }	

	if index == -1{
		return errors.New("poll does not exist in voter")
	}
	
	voter.VoteHistory[index] = voter.VoteHistory[len(voter.VoteHistory)-1]
	voter.VoteHistory = voter.VoteHistory[:len(voter.VoteHistory)-1]
	v.UpdateVoter(voter)

	return nil
}

// UpdateVoterPoll accepts a voter id and poll to update fpr the voter.
// Preconditions:   (1) The database file must exist and be a valid
//
//					(2) The voter must exist in the DB
//	    				because we use the voter.VoterId as the key, this
//						function must check if the voter already
//	    				exists in the DB, if not, return an error
//
// Postconditions:
//
//	    (1) The poll will be updated in the DB
//		(2) The DB file will be saved with the poll updated
//		(3) If there is an error, it will be returned
func (v *VoterList) UpdateVoterPoll(voterId uint, requestVoter Voter) error {

	// Check if voter exists before trying to update it
	// this is a good practice, return an error if the
	// voter does not exist

	var voter Voter
	pattern := redisKeyFromId(voterId)
	err := v.getItemFromRedis(pattern, &voter)
	if err != nil {
		return errors.New("voter does not exist")
	}

	requestPoll := requestVoter.VoteHistory[0]

	index := -1
    for i, poll := range voter.VoteHistory {
        if poll.PollID == requestPoll.PollID{
            index = i
            break
        }
    }	

    if index == -1 {
        return errors.New("poll does not exist in voter")
    } 
	
	voter.VoteHistory[index] = requestPoll
	v.UpdateVoter(voter)

	return nil
}

func (v *VoterList) GetHealthData(bootTime time.Time, calls uint) (healthData, error){

	v.healthInfo = healthData{Uptime: time.Now().Sub(bootTime), APIcalls: calls}

	return v.healthInfo, nil
}