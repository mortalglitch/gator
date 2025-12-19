*Gator RSS Aggregator and Feed Reader*

Requirements: 
- Go (language)
- Postgres (Database)

You can build gator using either 'go build gator' to build the package locally or 'go install gator' to install it to the go bin folder.

A config file will need to be made in the users home directory. 
Example config: ~/.gatorconfig.json
{"db_url":"postgres://username:@localhost:5432/gator?sslmode=disable","current_user_name":"username"}
Where the "postgres://username" will be the username you use with your local postgres install 

A few commands:
- gator register  - Register a new user
- gator login (username) - Log into a specific user.
- gator reset  - resets and drops tables from the current database.
- gator users  - lists all users from database.
- gator agg [optional: time 1s, 1m, 1hr]   - starts the aggregation process based on the time interval 15s for example would refresh every 15 seconds.
- gator addfeed ("name") ("url") - adds a feed to the current login users follow lists
- gator follow ("url") - follows a feed based on URL



