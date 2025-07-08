# KubeCloud API Tests

This directory contains test files that demonstrate and validate the KubeCloud API functionality.

## Test Files

### 1. `register_test.go`
- **Purpose**: Tests user registration functionality
- **Function**: `TestRegister`
- **Actions**: Registers a new user with the API
- **Note**: Will log a message if user already exists

### 2. `deploy_test.go` 
- **Purpose**: Tests deployment workflow with SSE monitoring
- **Function**: `TestDeployment`
- **Actions**: 
  - Logs in with existing user credentials
  - Deploys a Kubernetes cluster 
  - Listens to Server-Sent Events (SSE) for deployment progress with detailed logging
- **Timeout**: 5 minutes for SSE listening
- **Output**: Logs all SSE messages received during deployment

### 3. `getter_test.go`
- **Purpose**: Tests listing and retrieving deployments and kubeconfig
- **Function**: `TestGetters`
- **Actions**:
  - Logs in with existing user credentials
  - Lists all deployments for the user
  - Retrieves details for the first deployment (if any exist)
  - Downloads kubeconfig for the first deployment (if any exist)
- **Output**: Prints deployment information in JSON format and kubeconfig preview

### 4. `kubeconfig_test.go`
- **Purpose**: Tests kubeconfig retrieval functionality
- **Function**: `TestKubeconfig`
- **Actions**:
  - Logs in with existing user credentials
  - Lists deployments to find available clusters
  - Downloads and validates kubeconfig for the first deployment
- **Output**: Validates kubeconfig structure and logs preview

### 5. `client.go`
- **Purpose**: Shared HTTP client implementation
- **Contains**: All API interaction methods used by the test files
  - `Register()`, `Login()`, `DeployCluster()`, `ListenToSSE()`
  - `ListDeployments()`, `GetDeployment()`, `GetKubeconfig()`

## Running the Tests

### Prerequisites
- KubeCloud backend server running on `localhost:8080`
- A registered user with credentials:
  - Email: `testuser@example.com`
  - Password: `testpassword123`

### Run Individual Tests

```bash
# Run registration test
go test -run TestRegister

# Run deployment test  
go test -run TestDeployment

# Run getter tests
go test -run TestGetters

# Run kubeconfig test
go test -run TestKubeconfig
```

### Run All Tests
```bash
go test ./...
```

## Test Configuration

The tests are configured to:
- Connect to the API at `http://localhost:8080/api/v1`
- Use a 30-second timeout for HTTP requests
- Use a 5-minute timeout for SSE connections
- Deploy clusters with 3 nodes (leader, master, worker)
- SSH to cluster nodes to retrieve kubeconfig files

## Notes

- Tests are designed to work as examples and don't clean up after completion
- Tests log progress information during execution
- SSE test will show real-time deployment progress updates
- Getter tests display full deployment information in JSON format
- Tests will exit gracefully when completed, regardless of success/failure
- Kubeconfig test downloads the kubectl configuration file for cluster access
