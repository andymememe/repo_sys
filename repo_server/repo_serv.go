package main

import (
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "github.com/gin-gonic/gin"
    "log"
    "fmt"
    "os"
    "time"
    "context"
    "io/ioutil"
    "encoding/json"
    "net/http"
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
              client *mongo.Client) (map[string]*mongo.Collection,
                                     time.Time, error) {
    var repos Repos
    var modifiedDate time.Time
    
    collections := make(map[string]*mongo.Collection)
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
        collections[element] = client.Database("repo").Collection(element)
    }
    
    return collections, modifiedDate, err
}

func getPackagesByName(logger *log.Logger,
                       collections map[string]*mongo.Collection,
                       name string) ([]Package, error) {
    var err error
    var packages []Package
    var cur *mongo.Cursor
    
    ctx := context.Background()
    for key, coll := range collections {
        logger.Println(logMsg("info",
                              "getPackagesByName",
                              "Check repo: " + key))
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

func getPackagesByPkgName(logger *log.Logger,
                          collections map[string]*mongo.Collection,
                          pkg_name string,
                          repo_name string) (Package, error) {
    var err error
    var pkg Package
    var cur *mongo.Cursor
    
    ctx := context.Background()
    logger.Println(logMsg("info",
                          "getPackagesByPkgName",
                          "Check repo: " + repo_name))
    cur, err = collections[repo_name].Find(ctx, bson.M{
        "package_name": pkg_name,
    })
    if err != nil {
        return pkg, err
    }
    defer cur.Close(ctx)
    
    for cur.Next(ctx) {
        pkg = Package{}
        err = cur.Decode(&pkg)
        if err != nil {
            return pkg, err
        }
        logger.Println(logMsg("debug",
                              "getPackagesByPkgName",
                              pkg.Name))
    }
    
    err = cur.Err()
    if err != nil {
        return pkg, err
    }
    
    return pkg, err
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
        query := c.Query("q")
        if query == "" {
            c.Abort()
            c.JSON(http.StatusBadRequest, gin.H{
                "message": "Request the parameter q for name.",
            })
            return
        }
        
        packages, err := getPackagesByName(logger, collections, query)
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
    
	r.GET("/get_package", func(c *gin.Context) {
        package_name := c.Query("pkg")
        repo_name := c.Query("repo")
        if package_name == "" {
            c.Abort()
            c.JSON(http.StatusBadRequest, gin.H{
                "message": "Request the parameter pkg for package name.",
            })
            return
        }
        
        if repo_name == "" {
            repo_name = "main"
            logger.Println(logMsg("warn",
                                  c.Request.URL.Path,
                                  "Use default repo: main"))
        }
        
        if _, ok := collections[repo_name]; !ok {
            repo_name = "main"
            logger.Println(logMsg("warn",
                                  c.Request.URL.Path,
                                  "Use default repo: main"))
        }
        
        pkg, err := getPackagesByPkgName(logger,
                                         collections,
                                         package_name,
                                         repo_name)
        if err != nil {
            c.Abort()
            c.JSON(http.StatusInternalServerError, gin.H{
                "message": "error",
                "exception": err.Error(),
            })
            return
        }
        
        if pkg.PackageName == "" {
            c.Abort()
            c.JSON(http.StatusNotFound, gin.H{
                "message": "Package not found",
                "repo": repo_name,
            })
            return
        }
        
		c.JSON(200, gin.H{
			"message": "ok",
            "repo": repo_name,
            "package": pkg,
		})
	})
    
	r.Run() // listen and serve on 0.0.0.0:8080
}