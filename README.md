# Simple Web Crawler

## System Design

The designed proposed for the crawler consists of:  
- A Vue client: Allows users to enter a url for a website to crawl, and see the results of the crawling process. 
- A Go API server: Gets the url to crawl, process it and send the results to the frontend using WebSockets.



## Requirements:

### Docker
Download and install Docker.
If you are on a Mac (see https://docs.docker.com/docker-for-mac/install).
If you are on Ubuntu (https://docs.docker.com/install/linux/docker-ce/ubuntu/).

## Running the development environment

The local development environment consists of 2 docker containers:

- `srv`: Go http server. Handles the API requests.
- `vue`: Vue.js frontend application.

#### Running the application
The recommended way to start all the services together is executing `docker-compose up` in the root project folder.

- Access the front-end site with the following URLs: `http://localhost:3000`.
- Enter the url of the website you want to crawl and wait for the results from the websocket connection. 

#### Running Separately

To run the `srv` services locally (outside of docker)

Stop the container
```
docker stop simplecrawler-srv-1
```


Run the local server (you should have Go 1.20 installed):
```
cd server (or cd worker)
go run main.go
```

To run the `vue` client service locally (outside of docker), stop the container and run (you should have Node.js v.16.17):
```
docker stop simplecrawler-vue-1
cd client
npm install
npm run dev
```

#### Stopping the application
To stop all the services execute `docker-compose down` in the root project folder.


#### Running the test suite
To run the test suite execute:
```
cd server (or cd worker)
go test ./...
```