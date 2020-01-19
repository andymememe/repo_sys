# Repositary System
A package repository written by Go lang.

## Usage
### Server
#### Requirement
##### Database
- MongoDB
##### Go Packages
- [Gin Server](github.com/gin-gonic/gin)
- [mongo-driver](go.mongodb.org/mongo-driver)
- [yaml.v3](gopkg.in/yaml.v3)
  
#### DB Setting
Copy *config/config_template.yaml* as *config/config_secret.yaml*. Change the DB setting inside the yaml file.</br>
You need a database in MongoDB call *repo*, and collection name is the name of repo.</br>
You can add repo name in config/repos.yaml under *repos*.

#### Command
```bash
cd repo_server
go build repo_server.go
./repo_server
```

### Client
#### Command
```bash
cd repo_client
go build repo_client.go
./repo_client
```
