# Konstruktor Coding Challenge

## Build
To build the project within a Docker environment, use the following command:
```bash
docker build -t konstruktor-test .
```
This creates a Docker image that includes all necessary environments and dependencies required to run the project.

## Run
To run the created Docker container, execute the following command:
```bash
docker run konstruktor-test
```

## Test Coverage Report
To generate a test coverage report and copy it to your local machine, follow these steps:


```bash
docker cp <containerID>:/app/coverage.out ./coverage.out
```
**Note**: Replace `<containerID>` with the actual ID of your running container. You can find your container ID by using the docker ps command.

Then, you can use Go tools to view the test coverage report, for example:
```bash
go tool cover -html=coverage.out
```