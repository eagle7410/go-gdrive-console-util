package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

var (
	googleSecretPath string = "../gsecret/"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := googleSecretPath + "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func main() {
	var command, file, id string
	params := os.Args[1:]

	for inx := range params {

		switch params[inx] {
		case "-id":
			id = params[inx+1]
		case "-f":
			file = params[inx+1]
		case "-c":
			command = params[inx+1]
		}
	}

	switch command {
	case "fileGet":
		fileGet(&file, &id)
	case "fileUpdate":
		fileUpdateCloud(&file, &id)
	case "fileCreate":
		fileCreateCloud(&file)
	case "fileList":
		filesList()
	default:
		log.Fatalf("Command %v not allowed \n", command)
	}

}
func initCloud() *drive.Service {
	b, err := ioutil.ReadFile(googleSecretPath + "credentials.json")

	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.

	config, err := google.ConfigFromJSON(b, drive.DriveScope)

	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	client := getClient(config)

	srv, err := drive.New(client)

	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	return srv
}

func fileGet(fileName, id *string) {

	if len(*fileName) < 1 {
		log.Fatalf("Unable to retrieve Drive client: %v", "File name is empty")
	}

	if len(*id) < 1 {
		log.Fatalf("Unable to retrieve Drive client: %v", "File Id is empty")
	}

	srv := initCloud()

	res, err := srv.Files.Get(*id).Download()

	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	out, err := os.Create(*fileName)

	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	defer out.Close()

	_, err = io.Copy(out, res.Body)

	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println("Download is ok")
}

func fileUpdateCloud(fileName, id *string) {

	if len(*fileName) < 1 {
		log.Fatalf("Unable to retrieve Drive client: %v", "File name is empty")
	}

	if len(*id) < 1 {
		log.Fatalf("Unable to retrieve Drive client: %v", "File Id is empty")
	}

	srv := initCloud()

	file, err := os.Open(*fileName)

	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	defer file.Close()

	f := &drive.File{Name: filepath.Base(*fileName)}

	res, err := srv.Files.Update(*id, f).Media(file).Do()

	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("%s, %s, %s\n", res.Name, res.Id, res.MimeType)
}

func fileCreateCloud(fileName *string) {
	if len(*fileName) < 1 {
		log.Fatalf("Unable to retrieve Drive client: %v", "File name is empty")
	}

	srv := initCloud()

	file, err := os.Open(*fileName)

	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer file.Close()

	f := &drive.File{Name: filepath.Base(*fileName)}

	res, err := srv.Files.Create(f).Media(file).Do()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("%s, %s, %s\n", res.Name, res.Id, res.MimeType)
}

func filesList() {

	srv := initCloud()

	r, err := srv.Files.List().
		Do()

	if err != nil {
		log.Fatalf("Unable to retrieve files: %v", err)
	}
	fmt.Println("Files:")
	if len(r.Files) == 0 {
		fmt.Println("No files found.")
	} else {
		for _, i := range r.Files {
			fmt.Printf("  name -> %s id -> %s \n", i.Name, i.Id)
		}
	}
}
