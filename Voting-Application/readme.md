## Voting Application API

This application keeps `voter`, `poll`, and `vote` items in a redis cashe using 3 APIs (Voters, Polls, Votes).

There are two options for running this application:

1) build locally and run:
- run ./build-better-docker.sh shell script to build the three API components
- 'docker compose -f docker-compose-better.yaml up' to start running the containers
- 'docker compose -f docker-compose-better.yaml down' to stop running the containers

2) run from builds already on docker hub: 
- 'docker compose up' to start running the containers
- 'docker compose down' to stop running the containers

Once containers are running access the main API endpoint at http://localhost:1100/votes.  Before creating a vote, there must first be an existing voter and existing poll, otherwise an error will be logged.

A Collection in Postman has been made to test the functionality of this application: https://www.postman.com/orbital-module-administrator-57603215/workspace/mattgott1231/collection/23094790-c689f8c7-5099-4cba-9b7f-f1ecf2acc7bb?action=share&creator=23094790

This application uses HATEOS hypermedia to provide the user with the available actions to seccesfully use and navigate the program.  A few actions are listed below for each API endpoint:


GET All Votes: 1100/votes/

POST Vote: 1100/votes/:id

DELETE All Votes: 1100/votes/

DELETE Vote: 1100/votes/:id

PUT Vote: 1100/votes/:id

GET All Voters: 1080/voters/

POST Voter: 1080/voters/:id

DELETE All Voters: 1080/voters/

DELETE Voter: 1080/voters/:id

PUT Voter: 1080/voters/:id

GET Voter Polls: 1080/voters/:id/polls

POST Voter Poll: 1080/voters/:id/polls/:pollId

DELETE Voter Polls: 1080/voters/:id/polls

DELETE Voter Poll: 1080/voters/:id/polls/:pollId

GET All Polls: 1090/Polls/

POST Poll: 1090/polls/:id

DELETE All Polls: 1090/polls/

DELETE Poll: 1090/polls/:id

PUT Poll: 1090/polls/:id



JSON formats for POST/PUT requests:

Voters: 

{

  "VoterID": uint,
  
  "FirstName": string,
  
  "LastName": string,
  
  "VoteHistory": []string
  
}

Polls:

{

  "PollID": uint,
  
  "PollTitle": string,
  
  "PollQuestion": string,
  
  "PollOptions": []string
  
}

Votes:

{

  "VoteID": uint,
  
  "VoterID": uint,
  
  "PollID": uint,
  
  "VoteValue": uint
  
}
