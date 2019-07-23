package main

import (
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/readpref"
    "testing"
    "os"
    "log"
    "context"
    "regexp"
    "strings"
)

func Prepare() (*log.Logger, *mongo.Client, map[string]*mongo.Collection) {
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
        panic(err.Error())
    }
    collections, _, err := getRepos(logger, client)
    if err != nil {
        panic(err.Error())
    }
    
    return logger, client, collections
}

func TestLogMsg(t *testing.T) {
    if result := logMsg("error", "test_tag", "test_msg");
       result != "[error] [test_tag] test_msg" {
        t.Errorf("Except: %s, got %s", "[error] [test_tag] test_msg", result)
    }
}

func TestCreateLog(t *testing.T) {
    f, err := os.OpenFile("log/test_log.log",
                          os.O_RDWR | os.O_CREATE | os.O_APPEND,
                          0666)
    defer f.Close()
    if err != nil {
        panic(err.Error())
    }
    logger := createLog(f)
    
    logger.Println("Test log")
    
    fi, err := f.Stat()
    if err != nil {
        t.Errorf("File stats error, %s", err.Error())
    }
    
    if fi.Size() <= 0 {
        t.Errorf("Log file writing failed")
    }
}

func TestConnectDB(t *testing.T) {
    ctx := context.Background()
    client, err := connectDB()
    if err != nil {
        t.Errorf("Connect DB error, %s", err.Error())
    }
    
    err = client.Ping(ctx, readpref.Primary())
    if err != nil {
        t.Errorf("Ping error, %s", err.Error())
    }
}

func TestGetRepos(t *testing.T) {
    r, err := regexp.Compile("\\d{4}\\-\\d{2}\\-\\d{2}\\s+\\d{2}:\\d{2}:\\d{2}\\.\\d+\\s+[+-]*\\d{4}\\s+\\w+")
    if err != nil {
        panic(err.Error())
    }
    
    logger, client, _ := Prepare()
    
    collections, time, err := getRepos(logger, client)
    if err != nil {
        t.Errorf("Get repo error, %s", err.Error())
    }
    
    timeStr := time.String()
    if !r.MatchString(timeStr) {
        t.Errorf("Time format error")
    }
    
    if len(collections) <= 0 {
        t.Errorf("Not getting any collection")
    }
}

func TestGetPackagesByName (t *testing.T) {
    logger, _, collections := Prepare()
    
    pkgs, err := getPackagesByName(logger, collections, "Placehold")
    if err != nil {
        t.Errorf("Get packages from all string error")
    }
    if pkgs == nil {
        t.Errorf("Not getting packages")
    }
    for _, pkg := range pkgs {
        if !strings.Contains(pkg.Name, "Placehold") {
            t.Errorf("Getting wrong package: %s", pkg.Name)
        }
    }
    
    pkgs, err = getPackagesByName(logger, collections, "Pl")
    if err != nil {
        t.Errorf("Get packages from start error")
    }
    if pkgs == nil {
        t.Errorf("Not getting packages")
    }
    for _, pkg := range pkgs {
        if !strings.Contains(pkg.Name, "Pl") {
            t.Errorf("Getting wrong package: %s", pkg.Name)
        }
    }
    
    pkgs, err = getPackagesByName(logger, collections, "ce")
    if err != nil {
        t.Errorf("Get packages from midddle error")
    }
    if pkgs == nil {
        t.Errorf("Not getting packages")
    }
    for _, pkg := range pkgs {
        if !strings.Contains(pkg.Name, "ce") {
            t.Errorf("Getting wrong package: %s", pkg.Name)
        }
    }
    
    pkgs, err = getPackagesByName(logger, collections, "ld")
    if err != nil {
        t.Errorf("Get packages from end error")
    }
    if pkgs == nil {
        t.Errorf("Not getting packages")
    }
    for _, pkg := range pkgs {
        if !strings.Contains(pkg.Name, "ld") {
            t.Errorf("Getting wrong package: %s", pkg.Name)
        }
    }
    
    pkgs, err = getPackagesByName(logger, collections, "xx")
    if err != nil {
        t.Errorf("Get no existing package error")
    }
    if pkgs != nil {
        t.Errorf("Getting wrong packages")
    }
}

func TestGetPackagesByPkgName (t *testing.T) {
    logger, _, collections := Prepare()
    pkg, err := getPackagesByPkgName(logger,
                                     collections,
                                     "ph",
                                     "test")
    if err != nil {
        t.Errorf("Get package error")
    }
    if pkg.PackageName == "" {
        t.Errorf("Not getting package")
    }
    if pkg.PackageName != "ph" {
        t.Errorf("Getting wrong package: %s", pkg.PackageName)
    }
                                     
    pkg, err = getPackagesByPkgName(logger,
                                     collections,
                                     "noph",
                                     "test")
    if err != nil {
        t.Errorf("Get no existing package error")
    }
    if pkg.PackageName != "" {
        t.Errorf("Getting wrong package")
    }
}