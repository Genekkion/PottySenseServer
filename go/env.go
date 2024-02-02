package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func setEnv() {
    readFile, err := os.Open("./.env")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("< setting environmental variables >")
    scanner := bufio.NewScanner(readFile)
    scanner.Split(bufio.ScanLines)
    for scanner.Scan() {
        strArr := strings.Split(scanner.Text(), "=")
        os.Setenv(strArr[0], strArr[1])
        fmt.Println(strArr[0], ": ", os.Getenv(strArr[0]))
    }
    fmt.Println()
}
