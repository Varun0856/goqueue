package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const baseURL = "http://localhost:8080"

type createJobRequest struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

type createJobResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: goqueue-cli <submit|get> [arguement]")
		os.Exit(1)
	}

	var err error

	switch os.Args[1] {
	case "submit":
		if len(os.Args) != 4 {
			err = fmt.Errorf("usage: goqueue-cli submit <type> <payload>")
		} else {
			err = submitJob(os.Args[2], os.Args[3])
		}
	case "get":
		if len(os.Args) != 3 {
			err = fmt.Errorf("usage: goqueue-cli get <job-id>")
		} else {
			err = getJob(os.Args[2])
		}
	default:
		err = fmt.Errorf("unknown command: %s", os.Args[1])
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func submitJob(jobType, payload string) error {
	reqBody := createJobRequest{
		Type:    jobType,
		Payload: payload,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := http.Post(
		baseURL+"/jobs",
		"application/json",
		bytes.NewReader(data),
	)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %s: %s", resp.Status, body)
	}

	var result createJobResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	fmt.Printf("Job submitted successfully\nID: %s\nStatus: %s\n", result.ID, result.Status)
	return nil
}

func getJob(id string) error {
	resp, err := http.Get(baseURL + "/jobs/" + id)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %s: %s", resp.Status, body)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	data, err := json.MarshalIndent(result, "", " ")
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}
