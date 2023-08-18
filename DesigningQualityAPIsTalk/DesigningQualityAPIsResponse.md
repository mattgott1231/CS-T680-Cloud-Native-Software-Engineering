## Assignment: Response to Designing Quality APIs by Martin Nally
The video lecture titled Designing Quality APIs by Martin Nally provides strong arguments and techniques for generating REST, or HTPP focused, APIs.  I learned that REST means to use HTTP as the universal API, the concept of which is described as an entity based API, while RPC APIs are procedural focused.  I think the most insightful point that Martin makes is about identity and understanding relationships when it comes to APIs.  The biggest problem with engineering a well-designed API is the relationship between entities and connecting them.  Understanding identity, more than just a name, is the key to understanding relationships.  The best method for API formatting it to put all of the identity and how to use the API within the API itself, and not in the documentation (especially regarding JSON payloads).  One big takeaway that I learned is to include URL type links in the JSON fields so not only does the user know right away what each means, but how the relational API calls can and should be used.  Martin describes his best methodology for simple and effective JSON formatting for data in web APIs in this way, and I will definitely be using provided documentation in my future work.

The discussion focuses mostly on REST APIs rather than RPC APIs, and even includes some pros and cons.  I have not even heard of RPC APIs, so I’m curious what are the scenarios that would promote the use of RPC APIs over REST APIs?  Is one better than the other, or is there a time and place for both?  I’m leaning towards the latter following the video, but I would like to understand more.  Also, how would a JSON payload be structured to use an API service from a separate/third-party provider?  Would a field just include a URL to an outside developer’s API service?  