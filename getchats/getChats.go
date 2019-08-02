package main

import (
	"flag"
	"fmt"
	"math"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Arman92/go-tdlib"
	"github.com/howeyc/gopass"
)

var (
	version, tagName, branch, commitID, buildTime string
)

var allChats []*tdlib.Chat
var haveFullChatList bool

func main() {

	version = fmt.Sprintf("Version: %s, Branch: %s, Build: %s, Build time: %s", tagName, branch, commitID, buildTime)

	tdlib.SetLogVerbosityLevel(1)
	tdlib.SetFilePath("./errors.txt")

	// Create new instance of client
	client := tdlib.NewClient(tdlib.Config{
		APIID:               "952817",
		APIHash:             "217d192bf3884ee374dd742eb2ddeba8",
		SystemLanguageCode:  "en",
		DeviceModel:         "Server",
		SystemVersion:       "1.0.0",
		ApplicationVersion:  version,
		UseMessageDatabase:  true,
		UseFileDatabase:     true,
		UseChatInfoDatabase: true,
		UseTestDataCenter:   false,
		DatabaseDirectory:   "./db",
		FileDirectory:       "./files",
		IgnoreFileNames:     false,
	})

	var proxyHost, proxyPort, proxyType string
	var proxyUser, proxyPass, mtprotoSecret string

	flag.Usage = func() {
		fmt.Println(version)
		fmt.Println("Usage:")

		flag.PrintDefaults()
	}

	flag.StringVar(&proxyHost, "host", "", "代理服务器IP")
	flag.StringVar(&proxyPort, "port", "", "代理服务器端口号")
	flag.StringVar(&proxyType, "type", "", "代理服务器类型（http, socks5, mtproto）")
	flag.StringVar(&proxyUser, "user", "", "代理服务器用户名（http, socks5）")
	flag.StringVar(&proxyPass, "password", "", "代理服务器密码（http, socks5）")
	flag.StringVar(&mtprotoSecret, "secret", "", "Mtproto 代理密钥（Mtproto）")
	flag.CommandLine.SetOutput(os.Stdout)
	flag.Parse()

	if proxyHost != "" && proxyPort != "" && proxyType != "" {
		if ip := net.ParseIP(proxyHost); ip == nil {
			fmt.Println("代理服务器IP错误！")
			os.Exit(1)
		}

		i, err := strconv.ParseInt(proxyPort, 10, 32)
		if err != nil {
			fmt.Println("代理服务器端口号错误")
			os.Exit(1)
		}
		port := int32(i)

		switch proxyType {
		case "http":
			fmt.Printf("设置HTTP代理，%s:%s\n", proxyHost, proxyPort)
			client.AddProxy(proxyHost, port, true, tdlib.NewProxyTypeHttp(proxyUser, proxyPass, false))
		case "socks5":
			fmt.Printf("设置Socks5代理，%s:%s\n", proxyHost, proxyPort)
			client.AddProxy(proxyHost, port, true, tdlib.NewProxyTypeSocks5(proxyUser, proxyPass))
		case "mtproto":
			fmt.Printf("设置Mtproto代理，%s:%s\n", proxyHost, proxyPort)
			client.AddProxy(proxyHost, port, true, tdlib.NewProxyTypeMtproto(mtprotoSecret))
		default:
			fmt.Println("未知代理类型")
			flag.Usage()
			os.Exit(1)
		}
	}

	// Handle Ctrl+C , Gracefully exit and shutdown tdlib
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		client.DestroyInstance()
		os.Exit(1)
	}()

	// Wait while we get AuthorizationReady!
	// Note: See authorization example for complete auhtorization sequence example
	for {
		currentState, _ := client.Authorize()
		if currentState.GetAuthorizationStateEnum() == tdlib.AuthorizationStateWaitPhoneNumberType {
			fmt.Print("电话号码：")
			var number string
			fmt.Scanln(&number)
			_, err := client.SendPhoneNumber(number)
			if err != nil {
				fmt.Printf("错误电话号码：%v\n", err)
			}
		} else if currentState.GetAuthorizationStateEnum() == tdlib.AuthorizationStateWaitCodeType {
			fmt.Print("验证码：")
			var code string
			fmt.Scanln(&code)
			_, err := client.SendAuthCode(code)
			if err != nil {
				fmt.Printf("错误验证码：%v\n", err)
			}
		} else if currentState.GetAuthorizationStateEnum() == tdlib.AuthorizationStateWaitPasswordType {
			fmt.Print("密码：")
			var password string
			var maskPassword []byte
			for len(maskPassword) < 1 {
				maskPassword, _ = gopass.GetPasswdMasked()
			}
			password = string(maskPassword)
			//fmt.Scanln(&password)
			_, err := client.SendAuthPassword(password)
			if err != nil {
				fmt.Printf("密码错误：%v\n", err)
			}
		} else if currentState.GetAuthorizationStateEnum() == tdlib.AuthorizationStateReadyType {
			fmt.Println("已认证，开启水龙头！")
			break
		}
	}

	// get at most 1000 chats list
	getChatList(client, 1000)
	fmt.Printf("共有 %d 个会话\n", len(allChats))

	for _, chat := range allChats {
		fmt.Printf("会话标题：%s，会话 ID：%v\n", chat.Title, chat.ID)
	}

	for {
		time.Sleep(1 * time.Second)
	}
}

// see https://stackoverflow.com/questions/37782348/how-to-use-getchats-in-tdlib
func getChatList(client *tdlib.Client, limit int) error {

	if !haveFullChatList && limit > len(allChats) {
		offsetOrder := int64(math.MaxInt64)
		offsetChatID := int64(0)
		var lastChat *tdlib.Chat

		if len(allChats) > 0 {
			lastChat = allChats[len(allChats)-1]
			offsetOrder = int64(lastChat.Order)
			offsetChatID = lastChat.ID
		}

		// get chats (ids) from tdlib
		chats, err := client.GetChats(tdlib.JSONInt64(offsetOrder),
			offsetChatID, int32(limit-len(allChats)))
		if err != nil {
			return err
		}
		if len(chats.ChatIDs) == 0 {
			haveFullChatList = true
			return nil
		}

		for _, chatID := range chats.ChatIDs {
			// get chat info from tdlib
			chat, err := client.GetChat(chatID)
			if err == nil {
				allChats = append(allChats, chat)
			} else {
				return err
			}
		}
		return getChatList(client, limit)
	}
	return nil
}
