package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	repodbcontroller "repo_sys/repo_server/repodb_controller"
)

// Repos saves list of repo
type Repos struct {
	Repos []string `json:"repos"`
}

func sliceContainString(list []string, ele string) bool {
	for _, val := range list {
		if val == ele {
			return true
		}
	}
	return false
}

func createLog(f *os.File) *log.Logger {
	var logger *log.Logger
	logger = log.New(f, "repo_server", log.LstdFlags|log.Lshortfile)

	return logger
}

func logMsg(typeStr string, tag string, msg string) string {
	return fmt.Sprintf("[%s] [%s] %s", typeStr, tag, msg)
}

func getRepos(logger *log.Logger) (Repos, time.Time, error) {
	var repos Repos
	var modifiedDate time.Time

	filename := "repos.json"
	jsonFile, err := os.Open(filename)
	if err != nil {
		return Repos{}, time.Time{}, err
	}

	fileStat, err := os.Stat(filename)
	if err != nil {
		return Repos{}, time.Time{}, err
	}
	modifiedDate = fileStat.ModTime()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return Repos{}, time.Time{}, err
	}
	err = json.Unmarshal(byteValue, &repos)
	if err != nil {
		return Repos{}, time.Time{}, err
	}

	return repos, modifiedDate, err
}

func main() {
	// Create log file
	f, err := os.OpenFile("log/server.log",
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0666)
	defer f.Close()
	if err != nil {
		panic(err.Error())
	}
	logger := createLog(f)

	// Connect to DB
	logger.Println(logMsg("info", "main", "Server starting..."))
	repoDBController := repodbcontroller.NewRepoDBController()
	err = repoDBController.ConnectDB()
	if err != nil {
		logger.Panic(logMsg("err", "connectDB", err.Error()))
	}

	// Get repo list
	repos, modDate, err := getRepos(logger)
	if err != nil {
		logger.Panic(logMsg("err", "getRepos", err.Error()))
	}

	// Open server
	r := gin.New()
	r.Use(func(c *gin.Context) {
		logger.Println(logMsg("info", c.Request.Method, c.Request.URL.Path))
	})

	// Get: Healthy check
	r.GET("/healthy_check", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "ok",
		})
	})

	// Get: Last update time
	r.GET("/last_update", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message":       "ok",
			"modified_date": modDate.String(),
		})
	})

	// Get: Get packages by name
	r.GET("/get_packages", func(c *gin.Context) {
		query := c.Query("q") // Search query for package name
		if query == "" {
			c.Abort()
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Request the parameter q for name.",
			})
			return
		}

		packages, err := repoDBController.GetPackagesByName(repos.Repos, query)
		if err != nil {
			c.Abort()
			c.JSON(http.StatusInternalServerError, gin.H{
				"message":   "error",
				"exception": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"message":  "ok",
			"packages": packages,
		})
	})

	// Get: Get one packages with package name
	r.GET("/get_package", func(c *gin.Context) {
		packageName := c.Query("pkg") // Package Name
		repoName := c.Query("repo")   // Repo name
		if packageName == "" {
			c.Abort()
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Request the parameter pkg for package name.",
			})
			return
		}

		if repoName == "" {
			repoName = "main"
			logger.Println(logMsg("warn",
				c.Request.URL.Path,
				"Use default repo: main"))
		}

		if sliceContainString(repos.Repos, repoName) {
			repoName = "main"
			logger.Println(logMsg("warn",
				c.Request.URL.Path,
				"Use default repo: main"))
		}

		pkg, err := repoDBController.GetPackageByPkgName(packageName, repoName)
		if err != nil {
			c.Abort()
			c.JSON(http.StatusInternalServerError, gin.H{
				"message":   "error",
				"exception": err.Error(),
			})
			return
		}

		if pkg.PackageName == "" {
			c.Abort()
			c.JSON(http.StatusNotFound, gin.H{
				"message": "Package not found",
				"repo":    repoName,
			})
			return
		}

		c.JSON(200, gin.H{
			"message": "ok",
			"repo":    repoName,
			"package": pkg,
		})
	})

	r.Run() // listen and serve on 0.0.0.0:8080
}
