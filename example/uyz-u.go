package main

import (
	"context"
	"fmt"
	"log"

	user "github.com/phrynus/go-utils/uyz-u"
)

func TestUyzUser() {
	client, err := user.New(user.ClientConfig{
		BaseURL:          "http://47.98.194.112:8008/api/user",
		AppID:            1000,
		AppKey:           "9c83f734ba94de6331d2d0a7d67579dd",
		Version:          "1.0.0",
		VersionIndex:     "test",
		ClientPrivateKey: "",
		ServerPublicKey:  "",
		EncryptionMode:   user.EncryptionAES,
		DisableSignature: false,              // flip to true if the backend skips MD5 signatures
		EncodingMode:     user.EncodingHex,   // used when selecting AES/DES/RC4
		SymmetricKey:     "QKR6SwTSfWTvidy1", // required for AES/DES/RC4 modes
	})
	if err != nil {
		log.Fatalf("create client: %v", err)
	}

	login, err := client.NewLogin().
		Account("51154393").
		Password("123456").
		UDID("SN-1234-4514").
		Do(context.Background())
	if err != nil {
		log.Fatalf("login request failed: %v", err)
	}

	fmt.Println(login)
	fmt.Println(client.NewVIP().Do())
	// fmt.Println(client.NewHeartbeat().Do())
	// extend, err := client.NewSetExtend().Key("test").Value("test").Do()
	// if err != nil {
	// 	log.Fatalf("set extend request failed: %v", err)
	// }
	// fmt.Println(extend)
	goods, err := client.NewGoods().Page(1).Do(context.Background())
	if err != nil {
		log.Println(err)
	}
	fmt.Println(goods)

	fmt.Println(client.NewPay().Type("ali").GID(14).Mode("qr").Do())
}
