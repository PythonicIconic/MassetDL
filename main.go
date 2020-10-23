package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
)

var (
	assetURI = regexp.MustCompile(`(data-mediathumb-url=")(.*)(")`)
	wg       = &sync.WaitGroup{}
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

func GetAssetType(id string) (assetType string, trueID string) {
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
			wg.Add(1)
			go Audio(id)

			return
		} else {
			fmt.Println("[ERROR] Unknown error occurred getting asset info:", err)

			return
		}
	}

	if asset.Item.Class == "Accessory" {
		wg.Add(1)
		go Accessory(id, asset.Item.Properties.Name)

		return
	} else if asset.Item.Class == "Folder" {
		wg.Add(1)
		go Accessory(strings.Split(asset.Item.NestedItem.AnimItem.Properties.Content.URL, "id=")[1], "UNKNOWN")

		return
	}

	assetType = asset.Item.Class
	trueID = strings.Split(asset.Item.Properties.Content.URL, "id=")[1]

	return
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
	defer wg.Done()
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
	defer wg.Done()
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
	defer wg.Done()
	body := DownloadAsset(id)

	err := ioutil.WriteFile(fmt.Sprintf("shirts/%v.png", id), body, 0777)
	if err != nil {
		fmt.Println("[ERROR] Failed to save shirt:", err)
		return
	}

	fmt.Println("[SUCCESS] Saved shirt", id)
}

func Pants(id string) {
	defer wg.Done()
	body := DownloadAsset(id)

	err := ioutil.WriteFile(fmt.Sprintf("pants/%v.png", id), body, 0777)
	if err != nil {
		fmt.Println("[ERROR] Failed to save pants:", err)
		return
	}

	fmt.Println("[SUCCESS] Saved pants", id)
}

func Tshirt(id string) {
	defer wg.Done()
	body := DownloadAsset(id)

	err := ioutil.WriteFile(fmt.Sprintf("tshirts/%v.png", id), body, 0777)
	if err != nil {
		fmt.Println("[ERROR] Failed to save tshirt:", err)
		return
	}

	fmt.Println("[SUCCESS] Saved tshirt", id)
}

func Face(id string) {
	defer wg.Done()
	body := DownloadAsset(id)

	err := ioutil.WriteFile(fmt.Sprintf("faces/%v.png", id), body, 0777)
	if err != nil {
		fmt.Println("[ERROR] Failed to save face:", err)
		return
	}

	fmt.Println("[SUCCESS] Saved face", id)
}

func main() {
	file, err := os.Open("assets.txt")
	if err != nil {
		fmt.Println("[ERROR] Failed to open assets.txt. Are you in the right folder?")
		return
	}
	defer file.Close()

	_ = os.Mkdir("shirts", 0777)
	_ = os.Mkdir("pants", 0777)
	_ = os.Mkdir("audio", 0777)
	_ = os.Mkdir("tshirts", 0777)
	_ = os.Mkdir("faces", 0777)
	_ = os.Mkdir("accessories", 0777)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		assetType, trueID := GetAssetType(scanner.Text())
		switch assetType {
		case "Shirt":
			wg.Add(1)
			go Shirt(trueID)
		case "Pants":
			wg.Add(1)
			go Pants(trueID)
		case "ShirtGraphic":
			wg.Add(1)
			go Tshirt(trueID)
		case "Decal":
			wg.Add(1)
			go Face(trueID)
		}
	}

	wg.Wait()
}
