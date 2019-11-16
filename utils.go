package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strconv"
)

func readFileString(name string) (string, error) {
	b, err := ioutil.ReadFile(name)
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSpace(b)), nil
}

func readFileInt64(name string) (int64, error) {
	strv, err := readFileString(name)
	if err != nil {
		return 0, fmt.Errorf("failed to read file: %v", err)
	}

	intv, err := strconv.ParseInt(strv, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to convert string: %v", err)
	}

	return intv, nil
}
