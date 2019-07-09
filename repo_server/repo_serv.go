package main

import (
    "go.mongodb.org/mongo-driver/mongo";
    "go.mongodb.org/mongo-driver/mongo/options";
    "go.mongodb.org/mongo-driver/bson";
    "go.mongodb.org/mongo-driver/bson/primitive";
    "github.com/gin-gonic/gin";
    "log";
    "fmt";
    "os";
    "time";
    "context";
    "io/ioutil";
    "encoding/json";
    "net/http";
)

type Repos struct {
    Repos []string `json:"repos"`
}

type Package struct {
    ID primitive.ObjectID `bson:"_id, omitempty" json:"_id, omitempty"`
    Name string `bson:"name" json:"name"`
    Version string `bson:"version" json:"version"`
    PackageName string `bson:"package_name" json:"package_name"`
    LastUpdate time.Time `bson:"last_update" json:"last_update"`
    Status string `bson:"status" json:"status"`
}

func logMsg(type_str string,
            tag string,
            msg string) string {
    return fmt.Sprintf("[%s] [%s] %s", type_str, tag, msg)
}

func createLog(f *os.File) *log.Logger {
    var logger *log.Logger
    logger = log.New(f, "repo_server", log.LstdFlags | log.Lshortfile)
    
    return logger
}

func connectDB() (*mongo.Client, error) {
    ctx := context.Background()
    client, err := mongo.NewClient(
        options.
            Client().
            ApplyURI("mongodb://repo:reporepo@localhost:27017/repo"),
    )
    err = client.Connect(ctx)
    
    return client, err
}

func getRepos(logger *log.Logger,
              client *mongo.Client) ([]*mongo.Collection, time.Time, error) {
    var repos Repos
    var collections []*mongo.Collection
    var modifiedDate time.Time
    
    filename := "repos.json"
    jsonFile, err := os.Open(filename)
    if err != nil {
        return collections, modifiedDate, err
    }
    
    fileStat, err := os.Stat(filename)
    if err != nil {
        return collections, modifiedDate, err
    }
    modifiedDate = fileStat.ModTime()
    
    byteValue, err := ioutil.ReadAll(jsonFile)
    if err != nil {
        return collections, modifiedDate, err
    }
    json.Unmarshal(byteValue, &repos)
    for _, element := range repos.Repos {
        logger.Println(logMsg("info", "getRepos", "Get repo: " + element))
        collections = append(collections,
                             client.Database("repo").Collection(element))
    }
    
    return collections, modifiedDate, err
}

func getPackagesByName(logger *log.Logger,
                 collections []*mongo.Collection,
                 name string) ([]Package, error) {
    var err error
    var packages []Package
    var cur *mongo.Cursor
    
    ctx := context.Background()
    for _, coll := range collections {
        logger.Println(logMsg("info",
                              "getPackages",
                              "Check repo: " + coll.Name()))
        cur, err = coll.Find(ctx, bson.M{
            "name": bson.M{
                "$regex": ".*" + name + ".*",
            },
        })
        if err != nil {
            return packages, err
        }
        defer cur.Close(ctx)
        
        for cur.Next(ctx) {
            elem := &Package{}
            err = cur.Decode(elem)
            if err != nil {
                return packages, err
            }
            packages = append(packages, *elem)
        }
        
        err = cur.Err()
        if err != nil {
            return packages, err
        }
    }
    
    return packages, err
}

func main() {
    f, err := os.OpenFile("log/server.log",
                          os.O_RDWR | os.O_CREATE | os.O_APPEND,
                          0666)
    defer f.Close()
    if err != nil {
        panic(err.Error())
    }
    logger := createLog(f)
    
    logger.Println(logMsg("info", "main", "Server starting..."))
    client, err := connectDB()
    if err != nil {
        logger.Panic(logMsg("err", "connectDB", err.Error()))
    }
    collections, modDate, err := getRepos(logger, client)
    if err != nil {
        logger.Panic(logMsg("err", "getRepos", err.Error()))
    }
    
	r := gin.New()
    r.Use(func(c *gin.Context) {
        logger.Println(logMsg("info", c.Request.Method, c.Request.URL.Path))
    })
    
	r.GET("/healthy_check", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "ok",
		})
	})
    
	r.GET("/last_update", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "ok",
            "modified_date": modDate.String(),
		})
	})
    
	r.GET("/get_packages", func(c *gin.Context) {
        package_name := c.Query("q")
        if package_name == "" {
            c.Abort()
            c.JSON(http.StatusBadRequest, gin.H{
                "message": "Request the parameter q for package name.",
            })
            return
        }
        packages, err := getPackagesByName(logger, collections, package_name)
        if err != nil {
            c.Abort()
            c.JSON(http.StatusInternalServerError, gin.H{
                "message": "error",
                "exception": err.Error(),
            })
            return
        }
        
		c.JSON(200, gin.H{
			"message": "ok",
            "packages": packages,
		})
	})
    
	r.Run() // listen and serve on 0.0.0.0:8080
}