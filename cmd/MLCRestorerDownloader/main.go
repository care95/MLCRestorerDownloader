package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"
	"strings"

	downloader "github.com/Xpl0itU/MLCRestorerDownloader"
)

func main() {
	fmt.Println("Menu:")
	fmt.Println("1. Download MLC titles")
	fmt.Println("2. Download SLC titles")
	fmt.Println("3. Exit")

	fmt.Print("Select an option: ")
	var inputKey string
	fmt.Scanln(&inputKey)

	switch inputKey {
	case "1":
		showSubmenu("MLC")
	case "2":
		showSubmenu("SLC")
	case "3":
		fmt.Println("Exiting...")
		return
	default:
		fmt.Println("Invalid option")
		return
	}
}

func showSubmenu(titleType string) {
	if titleType != "MLC" && titleType != "SLC" {
		fmt.Println("Invalid title type")
		return
	}
	titles, err := readTitleInfoFromFile("titles.json")
	if err != nil {
		fmt.Println("[Error]", err)
		return
	}

	var chosenTitles map[string][]string
	switch titleType {
	case "MLC":
		chosenTitles = titles.MLC
	case "SLC":
		chosenTitles = titles.SLC
	default:
		fmt.Println("Invalid option")
		return
	}

	fmt.Println("Menu:")
	fmt.Printf("1. Download EUR %s titles\n", titleType)
	fmt.Printf("2. Download USA %s titles\n", titleType)
	fmt.Printf("3. Download JPN %s titles\n", titleType)
	fmt.Println("4. Back to main menu")

	fmt.Print("Select an option: ")
	var inputKey string
	fmt.Scanln(&inputKey)

	switch inputKey {
	case "1":
		downloadTitles("EUR", chosenTitles, titleType)
	case "2":
		downloadTitles("USA", chosenTitles, titleType)
	case "3":
		downloadTitles("JPN", chosenTitles, titleType)
	case "4":
		fmt.Println("Going back to the main menu...")
		main()
	default:
		fmt.Println("Invalid option")
		return
	}
}

func downloadTitles(region string, titles map[string][]string, titleType string) {
	selectedRegionTitles := titles[region]
	allRegionTitles := titles["All"]

	allTitles := append(selectedRegionTitles, allRegionTitles...)

	var failedTitles []string

	commonKey, err := getCommonKey()
	if err != nil {
		fmt.Println("[Error]", err)
		return
	}

	client := &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   100,
			MaxConnsPerHost:       100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	progressReporter := NewProgressReporterCLI()

	for _, titleID := range allTitles {
		if titleID == "dummy" {
			continue
		}

		var version string
		if parts := strings.Split(titleID, ":v"); len(parts) == 2 {
			titleID = parts[0]
			version = parts[1]
			fmt.Printf("\n[Info] Downloading version %s for title %s on region %s for type %s\n\n", version, titleID, region, titleType)
		} else {
			fmt.Printf("\n[Info] Downloading the latest version for title %s on region %s for type %s\n\n", titleID, region, titleType)
		}

		if err := downloader.DownloadTitle(titleID, fmt.Sprintf("output/%s/%s/%s", titleType, region, titleID), progressReporter, client, commonKey, version); err != nil {
			fmt.Println("[Error]", err)
			failedTitles = append(failedTitles, titleID)
			continue
		}
		fmt.Printf("\n[Info] Download files for title %s on region %s for type %s done\n\n", titleID, region, titleType)
	}

	if len(failedTitles) == 0 {
		fmt.Println("[Info] All done!")
		fmt.Println("Press Enter to exit...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		os.Exit(1)
	} else {
		fmt.Println("[Error] Some titles were not able to be downloaded.  Would you like to retry?")
		fmt.Println("1. Yes, retry ")
		fmt.Println("2. No, not right now")
		fmt.Print("Select an option: ")
		var inputKey string
		fmt.Scanln(&inputKey)

		switch inputKey {
		case "1":
			downloadTitles(region, titles, titleType)
		default:
			os.Exit(1)
		}
	}
}
