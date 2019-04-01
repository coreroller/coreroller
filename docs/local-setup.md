# Local setup

### Database (postgress)
Make sure you have installed [docker](https://docs.docker.com/install/)  

Run:  
```sh
docker pull coreroller/postgres
docker run docker run -d -p 5432:5432 coreroller/postgres
```

Run sql queries
```
docker exec -it {{containerID}} sh -c "su - postgres"
psql
\connect coreroller
```

### Backend
Make sure you have installed [go](https://golang.org/)  

Beside that you will need a view other go binaries:  
```sh
go get -u github.com/constabulary/gb/... github.com/go-bindata/go-bindata/...
```

When installed the above you can build and run the project
```sh
cd backend

# You probebly only have to run this once, though when chaning .sql files you need to re-run this
gb generate

# Run this every time you have changed some code
gb build
./bin/rollerd -bind :8000
```

### Frontend
Make sure you have installed [nodejs](https://nodejs.org/en/)  

Install the needed packages
```
cd frontend
npm i
```

After that you a view command you can run
```sh
# Build the project for production
npm run build

# Watch the files and transpile when files change
npm run watch
```

