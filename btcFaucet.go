package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"syscall"
	"time"

	"github.com/Arman92/go-tdlib"
	"github.com/howeyc/gopass"
)

var (
	version, tagName, branch, commitID, buildTime string
)

func main() {
	fmt.Println("比特币水龙头")
	version = fmt.Sprintf("Version: %s, Branch: %s, Build: %s, Build time: %s", tagName, branch, commitID, buildTime)

	tdlib.SetLogVerbosityLevel(0)
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
	var waitTime int

	flag.Usage = func() {
		fmt.Println(version)
		fmt.Println("Usage:")
		flag.PrintDefaults()
	}

	flag.IntVar(&waitTime, "wait", 30, "开启水龙头间隔时间，随机增加0-30秒")
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

	go func() {
		// Create an filter function which will be used to filter out unwanted tdlib messages
		eventFilter := func(msg *tdlib.TdMessage) bool {
			updateMsg := (*msg).(*tdlib.UpdateNewMessage)
			// For example, we want incomming messages from user with below id:
			if updateMsg.Message.SenderUserID == 848714900 {
				return true
			}
			return false
		}

		// Here we can add a receiver to retreive any message type we want
		// We like to get UpdateNewMessage events and with a specific FilterFunc
		receiver := client.AddEventReceiver(&tdlib.UpdateNewMessage{}, eventFilter, 100)
		for newMsg := range receiver.Chan {
			//fmt.Println(newMsg)
			updateMsg := (newMsg).(*tdlib.UpdateNewMessage)
			// We assume the message content is simple text: (should be more sophisticated for general use)
			msgText := updateMsg.Message.Content.(*tdlib.MessageText)
			re := regexp.MustCompile(`^.+฿([0-9.]+).+฿([0-9.]+)$`)
			account := re.FindStringSubmatch(msgText.Text.Text)
			if len(account) > 0 {
				fmt.Printf("%s 得到：%s，账户总额：%s\n", time.Now().Format("2006-01-02 15:04:05"), account[1], account[2])
			}
			//fmt.Printf("MsgText: %s\n\n", msgText.Text.Text)
		}

	}()

	go func() {
		rand.Seed(time.Now().UnixNano())
		// Should get chatID somehow, check out "getChats" example
		chatID := int64(848714900)
		inputMsgTxt := tdlib.NewInputMessageText(tdlib.NewFormattedText("💦 Faucet", nil), true, true)
		var w int
		for {
			client.SendMessage(chatID, 0, false, true, nil, inputMsgTxt)

			w = waitTime + rand.Intn(30)
			//fmt.Printf("sleep %d second\n", w)
			time.Sleep(time.Duration(w) * time.Second)
		}
	}()

	for {
		time.Sleep(1 * time.Second)
	}
	// rawUpdates gets all updates comming from tdlib
	//rawUpdates := client.GetRawUpdatesChannel(100)
	//for update := range rawUpdates {
	// Show all updates
	//	fmt.Println(update.Data)
	//	fmt.Print("\n\n----------\n\n")
	//}
}
