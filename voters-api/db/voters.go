package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
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

type DbMap map[uint]Voter

type VoterList struct {
	voterMap DbMap //A map of VoterIDs as keys and Voter structs as values
	healthInfo healthData
}

//constructor for VoterList struct
func NewVoterList() (*VoterList, error) {
	voterList := &VoterList{
		voterMap: make(DbMap),
		healthInfo: healthData{},
	}
	return voterList, nil
}

type healthData struct{
	Uptime time.Duration
	APIcalls uint
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
	_, ok := v.voterMap[voter.VoterID]
	if ok {
		return errors.New("voter already exists")
	}

	//Now that we know the vpter doesn't exist, lets add it to our map
	v.voterMap[voter.VoterID] = voter

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

	// we should if voter exists before trying to delete it
	// this is a good practice, return an error if the
	// voter does not exist

	_, ok := v.voterMap[id]
	if !ok {
		return errors.New("voter not found")
	}

	//Now lets use the built-in go delete() function to remove
	//the voter from our map
	delete(v.voterMap, id)

	return nil
}

// DeleteAllVoters removes all voters from the DB.
// It will be exposed via a DELETE /voters endpoint
func (v *VoterList) DeleteAllVoters() error {
	//To delete everything, we can just create a new map
	//and assign it to our existing map.  The garbage collector
	//will clean up the old map for us
	v.voterMap = make(DbMap)

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
	_, ok := v.voterMap[voter.VoterID]
	if !ok {
		return errors.New("voter does not exist")
	}

	//Now that we know the voter exists, lets update it
	v.voterMap[voter.VoterID] = voter

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
	existingVoter, ok := v.voterMap[id]
	if !ok {
		return Voter{}, errors.New("voter does not exist")
	}

	return existingVoter, nil
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
	var localVoterList []Voter

	//Now lets iterate over our map and add each voter to our slice
	for _, voter := range v.voterMap {
		localVoterList = append(localVoterList, voter)
	}

	//Now that we have all of our voters in a slice, return it
	return localVoterList, nil
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
func (v *VoterList) GetVoterPolls(id uint) ([]voterPoll , error) {

	// we should if voter exists before trying to retriece polls
	// this is a good practice, return an error if the
	// voter does not exist

	_, ok := v.voterMap[id]
	if !ok {
		return []voterPoll{}, errors.New("voter not found")
	}

	//Now lets use the built-in go delete() function to remove
	//the voter from our map

	return v.voterMap[id].VoteHistory, nil
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

    voter, voterExists := v.voterMap[voterId]
    if !voterExists {
        return voterPoll{}, errors.New("voter not found")
    }

    tempPoll := voterPoll{}
    for _, poll := range voter.VoteHistory {
        if poll.PollID == pollId{
            tempPoll = poll
            break
        }
    }

    emptyPoll := voterPoll{}

    if tempPoll == emptyPoll {
        return emptyPoll, errors.New("poll not found for given voter")
    } 

    return tempPoll, nil
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

	voter, voterExists := v.voterMap[voterId]
    if !voterExists {
        return errors.New("voter not found")
    }

	emptyPoll := voterPoll{}
	requestPoll := requestVoter.VoteHistory[0]

	tempPoll := emptyPoll
    for _, poll := range voter.VoteHistory {
        if poll.PollID == requestPoll.PollID{
            tempPoll = poll
            break
        }
    }	

    if tempPoll != emptyPoll {
        return errors.New("poll already exists in voter")
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

	voter, voterExists := v.voterMap[voterId]
    if !voterExists {
        return errors.New("voter not found")
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

	voter, voterExists := v.voterMap[voterId]
    if !voterExists {
        return errors.New("voter not found")
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