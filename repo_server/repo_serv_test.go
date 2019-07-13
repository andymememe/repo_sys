package main

import (
    "testing";
    "os";
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