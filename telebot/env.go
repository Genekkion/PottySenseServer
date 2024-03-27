package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

// Sets the environmental variables from the specified
// env file. If there are any required variables not set,
// the function panics. For optional variables, the default
// value will be used according to the input map. Returns the
// env map.
func SetEnv(required []string, optional map[string]string) map[string]string {

	readFile, err := os.Open("./.env")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("The following environmental variables have been set")
	log.Print()
	scanner := bufio.NewScanner(readFile)
	scanner.Split(bufio.ScanLines)
	envMap := map[string]string{}
	//i := 1
	for scanner.Scan() {
		// Trims leading, 
		strArr := strings.Split(
			strings.TrimSpace(scanner.Text()),
			"=")
		if len(strArr) != 2 {
			continue
		}
		fmt.Printf("%s ", strArr[0])
		os.Setenv(strArr[0], strArr[1])
		envMap[strArr[0]] = strArr[1]
	}
	fmt.Println()
	log.Println("Environmental variables have been set.")
	return envMap
}
