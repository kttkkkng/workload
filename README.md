# Workload Invoker
Workload Invoker is an application which is used to invoke faas application.
## Each File Responsibility
- `Checker.go`: To check if config file is valid.
- `EventGenerator.go`: To generate the time scedule which is used for sending request.
- `Invoker.go`: To send request and receive response.
- `client.go`, `kvslib`, `pb`, `proto`: Only used for KVS.
## Usage Guide
### The Test Configuration
Before run the test, you need to prepare a workload configuration file which define the scenario that you want to test. Parameter detail:
1. Primary fields:
    i. `duration`: Determines the length of the test in seconds
    ii. `instances`: A collection of the invocation instances. Each instance describes the invocation behavior for an application. Moreover, Each instance run independently with others.
    iii. `FrontEndAddr`: The address of the frontend service (must reimplement)
2. Each invocation instance:
    i. `application`: The action of this instance. In KVS, we only use Get and Put (must reimplement)
    ii. `distribution`: The distribution of this instance. Support only `Unifrom` distribution.
    iii. `rate`: In `Uniform` distribution, rate is the amount of request per second. (rate: 100 mean 100 requests per second)
    iv. `activity_window`: If set to null, the application is invoked during the entire test. By setting a time window, one can limit the activity of the application to a sub-interval of the test.
    v: `dataFile`: The path of the data file which is used to send request.
### Run the Test
```bash
./Invoker <config_file_path>
```
If there no Invoker file, run
```bash
make build
```
## Before the Usage
This application only has been made for KVS application.
- Please reimplement the send request and receive response part which are `HTTPInstanceGenerator`, `post` and `ReceiveResponse` function at `Invoker.go` file and remove `kvslib`, `pb`, `proto` and `clent.go`
## PS
I want to recheck the code once. Please use this application after 10 August 2021.