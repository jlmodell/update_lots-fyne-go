package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

const (
	lotsFile  = "C://Temp//Lots.csv"
	targetUrl = "http://128.1.0.155:8089/update"
	logPath   = "C://Temp//lot-updater.log"
)

var logger *log.Logger

func init() {
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	logger = log.New(logFile, "", log.LstdFlags|log.Lshortfile)
}

type Update struct {
	Lot        string `json:"lot"`        // lot number
	Part       string `json:"part"`       // part number
	Expiration string `json:"expiration"` // expiration date
	OnHand     string `json:"on_hand"`    // on hand quantity
	Allocated  string `json:"allocated"`  // allocated quantity
}

type UpdateResponse struct {
	Updates     []Update `json:"updates"`
	ErrorsCount int      `json:"constraint_errors_count"`
	Count       int      `json:"count"`
}

func main() {
	a := app.New()
	w := a.NewWindow("Busse | Lots Updater")

	title := widget.NewLabel("Lot Updater")

	w.SetContent(container.NewVBox(
		title,
		layout.NewSpacer(),
		widget.NewLabel("Make sure to run UPDATE.LOTS before running this program."),
		layout.NewSpacer(),
		widget.NewButton("Upload Updated Lot File", func() {
			resp, err := uploadFile(lotsFile, targetUrl)
			if err != nil {
				panic(err)
			}
			var updatedLots []string
			for _, update := range resp.Updates {
				// logger
				logger.Printf("%+v", update)

				updatedLots = append(updatedLots, fmt.Sprintf("Lot: %s\tPart: %s\tExpiration: %s\tOn Hand: %s\tAllocated: %s", update.Lot, update.Part, update.Expiration, update.OnHand, update.Allocated))
			}

			updatedLotsString := "Nothing to update."
			if len(updatedLots) > 0 {
				updatedLotsString = strings.Join(updatedLots, "\n")
			}

			w.SetContent(
				container.NewVBox(
					widget.NewLabel(fmt.Sprintf("Updated %d lots", len(resp.Updates))),
					widget.NewLabel(fmt.Sprintf("Found %d errors", resp.ErrorsCount)),
					widget.NewLabel(fmt.Sprintf("Found %d lots", resp.Count)),
					layout.NewSpacer(),
					widget.NewLabel(updatedLotsString),
				),
			)
		}),
	))

	w.ShowAndRun()
}

func uploadFile(filename string, targetUrl string) (UpdateResponse, error) {
	var updateResponse UpdateResponse
	// Open file for uploading
	file, err := os.Open(filename)
	if err != nil {
		return updateResponse, err
	}
	defer file.Close()
	// Create form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	// Add file
	part, err := writer.CreateFormFile("file", path.Base(filename))
	if err != nil {
		return updateResponse, err
	}
	io.Copy(part, file)
	writer.Close()
	// Create request
	req, err := http.NewRequest("POST", targetUrl, body)
	if err != nil {
		return updateResponse, err
	}
	// Set content type
	req.Header.Set("Content-Type", writer.FormDataContentType())
	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return updateResponse, err
	}
	defer resp.Body.Close()
	// Check response
	if resp.StatusCode != http.StatusOK {
		return updateResponse, errors.New("bad status: " + resp.Status)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)

	json.Unmarshal(buf.Bytes(), &updateResponse)

	return updateResponse, nil
}
