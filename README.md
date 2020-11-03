# bitburst

**dataservice**		- interface for database

**onlineservice**	- service for getting a user's online status

**userservice**		-service for getting online status of many users

**Steps to run the program**

**1.Setting database**

a.docker run -d -p 5432:5432 --name bitburst-postgres -e POSTGRES_PASSWORD=secret  postgres

b.docker exec -it bitburst-postgres psql -U postgres

c.CREATE DATABASE bitburst;

**2.Running application**

a. git clone https://github.com/tebrizetayi/bitburst/ and 

b. After downloading is completed run  "cd bitburst"

c. Open terminal and run command: "go mod init ."

d. Go to the onlineuser folder : "cd onlineservice"

e. Run the onlineservice: "go run main.go"

f. Open new terminal in folder "bitburst"

g. Go to the userservice folder : "cd userservice" 

h. Run the userservice: "go run main.go"

In the terminal for onlineservice you can see the id of the online users





