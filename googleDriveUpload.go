package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3" //v2跟v3不合用要注意
)

// refs https://developers.google.com/drive/v3/web/quickstart/go

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile := tokenCacheFile() //取得token的存放位置
	//if err != nil {
	//	log.Fatalf("Unable to get path to cached credential file. %v", err)
	//}
	tok, err := tokenFromFile(cacheFile) //呼叫tokenFromFile取得token檔
	if err != nil { //注意這邊的err不一定是沒有token，也有可能是token的Decode錯誤
		tok = getTokenFromWeb(config) //呼叫getTokenFromWeb重新產生一個網址要求複製token並貼上
		saveToken(cacheFile, tok)     //取得新的token就存檔到cacheFile的路徑
	}
	return config.Client(ctx, tok) //成功取得token並return
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string) {
	//此為另一範例，會在當前user資料夾裡建立資料夾
	//usr, err := user.Current()
	//if err != nil {
	//	return "", err
	//}
	//tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	//os.MkdirAll(tokenCacheDir, 0700)
	//return filepath.Join(tokenCacheDir, url.QueryEscape("drive-go-quickstart.json")), err

	//這邊定義token存放的路徑及資料夾，並且建立資料夾
	tokenCacheDir := filepath.Join("./", "googleDriveToken")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir, url.QueryEscape("drive-go-quickstart.json"))
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	//搜尋路徑，有token檔就開啟並Decode，沒有就回傳nil
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t) //Decode錯誤也會回傳err
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.Create(file)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func main() {
	//這邊是用參數的方式開啟，輸入要上傳的檔案名稱
	//if len(os.Args) != 2 {
	//	fmt.Fprintln(os.Stderr, "Usage: drive filename (to upload a file)")
	//	return
	//}
	//filename := os.Args[1]

	filename := "滑鼠.JPG"

	ctx := context.Background()

	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/drive-go-quickstart.json
	config, err := google.ConfigFromJSON(b, drive.DriveFileScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(ctx, config)

	srv, err := drive.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve drive Client %v", err)
	}

	f, err := os.Open(filename)
	if err != nil {
		log.Fatalf("error opening %q: %v", filename, err)
	}
	defer f.Close()

	//建立資料夾，給定Name跟MimeType，最後Do會回傳一些資料，型態是dict，資料內容可參考API
	createFolder, err := srv.Files.Create(&drive.File{Name: "testFolder", MimeType: "application/vnd.google-apps.folder"}).Do()
	if err != nil {
		log.Fatalf("Unable to create folder: %v", err)
	}

	//建立array存放資料夾ID，上面建立資料夾Do的回傳內容就包括資料夾ID
	var folderIDList []string
	folderIDList = append(folderIDList, createFolder.Id)
	//上傳檔案，create要給定檔案名稱，要傳進資料夾就加上Parents參數給定folderID的array，media傳入我們要上傳的檔案，最後Do
	driveFile, err := srv.Files.Create(&drive.File{Name: filename, Parents: folderIDList}).Media(f).Do()
	if err != nil {
		log.Fatalf("Unable to create file: %v", err)
	}

	log.Printf("file: %+v", driveFile)
	log.Println("done")
}
