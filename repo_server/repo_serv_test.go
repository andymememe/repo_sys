package main

import (
    "go.mongodb.org/mongo-driver/mongo/readpref";
    "testing";
    "os";
    "context";
    "regexp";
)

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

    f, err := os.OpenFile("log/test_log.log",
                          os.O_RDWR | os.O_CREATE | os.O_APPEND,
                          0666)
    if err != nil {
        panic(err.Error())
    }
    logger := createLog(f)
    
    client, err := connectDB()
    if err != nil {
        panic(err.Error())
    }
    
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