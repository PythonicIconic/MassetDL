package main

import (
	"bufio"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	assetURI = regexp.MustCompile(`(data-mediathumb-url=")(.*)(")`)
	mainwg   = &sync.WaitGroup{}
	inputwg  = &sync.WaitGroup{}
)

type Asset struct {
	XMLName xml.Name `xml:"roblox"`
	Item    Item     `xml:"Item"`
}

type Item struct {
	XMLName    xml.Name   `xml:"Item"`
	NestedItem NestedItem `xml:"Item"`
	Class      string     `xml:"class,attr"`
	Properties Property   `xml:"Properties"`
}

type NestedItem struct {
	XMLName    xml.Name `xml:"Item"`
	AnimItem   AnimItem `xml:"Item"`
	Class      string   `xml:"class,attr"`
	Properties Property `xml:"Properties"`
}

type AnimItem struct {
	XMLName    xml.Name `xml:"Item"`
	Class      string   `xml:"class,attr"`
	Properties Property `xml:"Properties"`
}

type Property struct {
	XMLName    xml.Name `xml:"Properties"`
	Content    Content  `xml:"Content"`
	Name       string   `xml:"string"`
	Archivable bool     `xml:"bool"`
}

type Content struct {
	XMLName xml.Name `xml:"Content"`
	URL     string   `xml:"url"`
}

func GetAssetType(id string, assetChan chan [2]string) {
	defer inputwg.Done()
	resp, err := http.Get(fmt.Sprintf("https://assetdelivery.roblox.com/v1/asset?id=%v", id))
	if err != nil {
		fmt.Println("[ERROR] Failed to make request:", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("[ERROR] Failed to read response body:", err)
		return
	}
	var asset Asset

	err = xml.Unmarshal(body, &asset)
	if err != nil {
		if err.Error() == "EOF" {
			fmt.Println("[WARNING] Given ID was not found, assuming it is audio")
			mainwg.Add(1)
			go Audio(id)

			return
		} else if err.Error() == "XML syntax error on line 3: invalid character entity & (no semicolon)" {
			fmt.Println("[ERROR] Old format detected, skipping")
			return
		} else {
			fmt.Println("[ERROR] Unknown error occurred getting asset info:", err, "ID:", id)
			return
		}
	}

	switch asset.Item.Class {
	case "Accessory":
		mainwg.Add(1)
		go Accessory(id, asset.Item.Properties.Name)

		return
	case "Folder":
		mainwg.Add(1)
		go Accessory(strings.Split(asset.Item.NestedItem.AnimItem.Properties.Content.URL, "id=")[1], "UNKNOWN")

		return
	case "Lighting":
		mainwg.Add(1)
		go Accessory(strings.Split(asset.Item.NestedItem.Properties.Content.URL, "id=")[1], "UNKNOWN")

		return
	case "ControllerService", "Model", "Teams", "Timer", "Part":
		return
	}

	if len(strings.Split(asset.Item.Properties.Content.URL, "id=")) < 2 {
		fmt.Println("[ERROR] Unknown error occurred when retrieving URL. Unsupported asset?")
		return
	}

	assetChan <- [2]string{asset.Item.Class, strings.Split(asset.Item.Properties.Content.URL, "id=")[1]}
}

func DownloadAsset(id string) (data []byte) {
	resp, err := http.Get(fmt.Sprintf("https://assetdelivery.roblox.com/v1/asset?id=%v", id))
	if err != nil {
		fmt.Println("[ERROR] Failed to make request:", err)
		return
	}
	defer resp.Body.Close()

	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("[ERROR] Failed to read response body:", err)
		return
	}

	return
}

func Accessory(id string, assetName string) {
	defer mainwg.Done()
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://assetdelivery.roblox.com/v1/asset?id=%v", id), nil)
	if err != nil {
		fmt.Println("[ERROR] Failed to create request:", err)
		return
	}

	req.Header.Set("accept-encoding", "*")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("[ERROR] Failed to make request:", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("[ERROR] Failed to read response body:", err)
		return
	}

	err = ioutil.WriteFile(fmt.Sprintf("accessories/%v_%v.rbxm", assetName, id), body, 0777)
	if err != nil {
		fmt.Println("[ERROR] Failed to save accessory:", err)
		return
	}

	fmt.Println("[SUCCESS] Saved accessory", id)
}

func Audio(id string) {
	defer mainwg.Done()
	resp, err := http.Get(fmt.Sprintf("https://roblox.com/library/%v/#", id))
	if err != nil {
		fmt.Println("[ERROR] Failed to create request:", err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("[ERROR] Failed to read response body:", err)
		return
	}
	resp.Body.Close()

	elem := assetURI.Find(body)
	if elem == nil {
		fmt.Println("[ERROR] Failed to find audio URL, invalid ID")
		return
	}

	url := strings.Split(string(elem), "\"")[1]
	resp, err = http.Get(url)
	if err != nil {
		fmt.Println("[ERROR] Failed to create request:", err)
		return
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("[ERROR] Failed to read response body:", err)
		return
	}

	err = ioutil.WriteFile(fmt.Sprintf("audio/%v.mp3", id), body, 0644)
	if err != nil {
		fmt.Println("[ERROR] Failed to save audio:", err)
		return
	}

	fmt.Println("[SUCCESS] Saved audio", id)
}

func Shirt(id string) {
	defer mainwg.Done()
	body := DownloadAsset(id)

	err := ioutil.WriteFile(fmt.Sprintf("shirts/%v.png", id), body, 0777)
	if err != nil {
		fmt.Println("[ERROR] Failed to save shirt:", err)
		return
	}

	fmt.Println("[SUCCESS] Saved shirt", id)
}

func Pants(id string) {
	defer mainwg.Done()
	body := DownloadAsset(id)

	err := ioutil.WriteFile(fmt.Sprintf("pants/%v.png", id), body, 0777)
	if err != nil {
		fmt.Println("[ERROR] Failed to save pants:", err)
		return
	}

	fmt.Println("[SUCCESS] Saved pants", id)
}

func Tshirt(id string) {
	defer mainwg.Done()
	body := DownloadAsset(id)

	err := ioutil.WriteFile(fmt.Sprintf("tshirts/%v.png", id), body, 0777)
	if err != nil {
		fmt.Println("[ERROR] Failed to save tshirt:", err)
		return
	}

	fmt.Println("[SUCCESS] Saved tshirt", id)
}

func Face(id string) {
	defer mainwg.Done()
	body := DownloadAsset(id)

	err := ioutil.WriteFile(fmt.Sprintf("faces/%v.png", id), body, 0777)
	if err != nil {
		fmt.Println("[ERROR] Failed to save face:", err)
		return
	}

	fmt.Println("[SUCCESS] Saved face", id)
}

func Scrape(filter string) {
	if filter == "Tshirt" {
		filter = "ShirtGraphic"
	} else if filter == "Face" {
		filter = "Decal"
	}

	assetChan := make(chan [2]string, 200)
	go func() {
		for id := 1000000; id < 5999999999; id++ {
			inputwg.Add(1)
			go GetAssetType(strconv.Itoa(id), assetChan)
			time.Sleep(50 * time.Millisecond)
		}

		inputwg.Wait()
		close(assetChan)
	}()

	for assetInfo := range assetChan {
		if assetInfo[0] == filter || filter == "" {
			switch assetInfo[0] {
			case "Shirt":
				mainwg.Add(1)
				go Shirt(assetInfo[1])
			case "Pants":
				mainwg.Add(1)
				go Pants(assetInfo[1])
			case "ShirtGraphic":
				mainwg.Add(1)
				go Tshirt(assetInfo[1])
			case "Decal":
				mainwg.Add(1)
				go Face(assetInfo[1])
			}
		} else {
			continue
		}
	}

	mainwg.Wait()
}

func File() {
	file, err := os.Open("assets.txt")
	if err != nil {
		fmt.Println("[ERROR] Failed to open assets.txt. Are you in the right folder?")
		return
	}
	defer file.Close()

	assetChan := make(chan [2]string, 20)
	go func() {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			inputwg.Add(1)
			go GetAssetType(scanner.Text(), assetChan)
		}

		inputwg.Wait()
		close(assetChan)
	}()

	for assetInfo := range assetChan {
		switch assetInfo[0] {
		case "Shirt":
			mainwg.Add(1)
			go Shirt(assetInfo[1])
		case "Pants":
			mainwg.Add(1)
			go Pants(assetInfo[1])
		case "ShirtGraphic":
			mainwg.Add(1)
			go Tshirt(assetInfo[1])
		case "Decal":
			mainwg.Add(1)
			go Face(assetInfo[1])
		}
	}

	mainwg.Wait()
}

func main() {
	_ = os.Mkdir("shirts", 0777)
	_ = os.Mkdir("pants", 0777)
	_ = os.Mkdir("audio", 0777)
	_ = os.Mkdir("tshirts", 0777)
	_ = os.Mkdir("faces", 0777)
	_ = os.Mkdir("accessories", 0777)

	var (
		scrape bool
		file   bool
		filter string
	)

	flag.StringVar(&filter, "filter", "", "Set a specific asset type to scrape for. Only works with scrape.\nSupported filters: Shirt, Pants, Tshirt, Face, Audio, Accessory")
	flag.BoolVar(&scrape, "scrape", false, "Set whether to scrape assets or not.")
	flag.BoolVar(&file, "file", false, "Set whether to use file or not.")
	flag.Parse()

	if scrape {
		if filter != "" {
			var match bool
			filter = strings.Title(strings.ToLower(filter))
			for _, asset := range []string{"Shirt", "Pants", "Tshirt", "Face", "Audio", "Accessory"} {
				if filter == asset {
					match = true
				}
			}

			if match {
				Scrape(filter)
			} else {
				fmt.Println("Invalid filter! Supported filters: Shirt, Pants, Tshirt, Face, Audio, Accessory")
			}
		} else {
			Scrape("")
		}
	} else if file {
		File()
	} else {
		fmt.Println("Oops! No option selected. Please use `massetdl -scrape` or `massetdl -file`!")
	}
}
