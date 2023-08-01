## Voters API 

This is an application for storing and managing voter data using an API.

It keeps `voter` items in memory right now.

The API supports the following actions

	GET "/voters"                      Returns a list of all voter data
	POST"/voters"                      Adds new voter with JSON payload
	PUT "/voters"                      Updates voters wiht JSON payload
	DELETE "/voters"                   Deletes all voters
	DELETE "/voters/:id"               Deletes voter by Voter ID
	GET "/voters/:id"                  Returns voter by Voter ID
	GET "/voters/:id/polls"            Returns all polls for a Voter ID
	GET "/voters/:id/polls/:pollId"    Returns a poll by Poll ID for a Voter ID
	POST "/voters/:id/polls"           Adds a new poll for a Voter ID with a JSON payload
	DELETE "/voters/:id/polls/:pollId" Deletes a poll by Poll ID for a Voter ID
	PUT "/voters/:id/polls"            Updates a poll by Poll ID for a Voter ID with a JSON payload
 	GET "/voters/health"               Returns total runtime and API calls since runtime

JSON Payloads should be provided in the following format:
  
 	POST and PUT "/voters"            {"VoterID":1,"FirstName":"John","LastName":"Doe","VoteHistory":[{"PollID":1,"VoteDate":"2023-07-25T23:36:24.820414-04:00"},{"PollID":2,"VoteDate":"2023-07-25T23:36:24.820414-04:00"}]}
	POST and PUT "/voters/:id/polls"  {"VoterID":1,"VoteHistory":[{"PollID":5,"VoteDate":"2023-07-25T23:36:24.820414-04:00"}]}
 
