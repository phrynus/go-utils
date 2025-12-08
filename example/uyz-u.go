package main

import (
	"context"
	"fmt"
	"log"

	user "github.com/phrynus/go-utils/uyz-u"
)

const rsaClientPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEoQIBAAKCAQBm/1vYhEeZrDQo33I99sNEq2dL3mn+J62sdp1YtOT7J8MbR0/c
DCJsZ5LVP4m9Psqdn0w9+MRL+8M8jmOritmbP0oJcpMK1k2VVaBJa19XRRm5ElOl
p25WlQar/HybbUYmu0JnWhoUxYc7hZcfb1rozuyarl9aOznKzi8G7KZaLSB+q/Ky
C6NizOhbUMWKxiK8ib3veOhv6mkzoiJm9zZ9JFyzSocEXZ/V4wir2Q5uib1JVsBX
eR/xz1JPcGAv69ZBF0ozXnH1AK+OeTKp/c3Gvo8j0RhqucMuAdmH3CECuBwVeHhh
7z9BOaiTrU5DcE9KULnUQfdNQCrMEM0T2ASxAgMBAAECggEADmtbmWFTgNEZ8Erv
/HrKdZele3qkzh8R8l4cwyl4ES1M5EnEhWLxgsmxRceCagsbZJvDmb11Bco2WAj7
LS3gLxraK55ttYuxCuIU8ZJlo5sZ8c3+Bef9y4NLjtJnQ3813gBDWKLFmfjZaNzX
/l+hk4w8lZ8p3EEEYz8gWFDTLTaMkFY4bPXnGylCVonedTWQSqizl7yph8YD6p7L
zEBL8Dv8lIA46Lwwh5W2UVyUC0aEpZLLU3+nMKfiUTs5790Fcz6aFPjfFST5txMX
iAjV/Cy79cnZ3x50jM653Bxf1LDN6nOpn9xJ8bcfXZonAU2pLWLePZMJRuYl7pYm
2+gNwQKBgQCtdTKC3COrHYHwc99oyhQOvYwlt1eRg50PogHV5RavLd/yaxV3mqKm
hYsTe91giyMPsCBsKY4OzpdUa+fVS+8DXoojsaoz6U1WJ9d0Nk2gm5I/k0pJX4az
CmG2Qxt8dShWOCFWNNld6ynGvdm3wmEqiJokk36GiE++WzxHqjkUGQKBgQCYAp3T
294zyJ3vSbxHbvhno8eKpY3tIoAVjd4P4AsaAMgBeAlh3UT4JfQu6hd8cz9k9562
cRQlsVIxQMQxLvYQEQGwBOZvte2yLvHqyNWl7YSIaf/ULbR7OCdU8EJN+LcR3L0A
UQaKJIGi0AIqykFTYSS45jbKiBK/ZogybkFIWQKBgGhpCGueyVWiIGo1xYAVS7eH
v0mgr/Rmbe9QDJzNFjeCfLA2Zyiki02DSzECOUJ43jT+RrX02Y7uKkdl4JoS6B92
E97ifdpbj/LRbq6EVXvcyU69gVTjTHiPQjvs7ymeeBZWGTMEAue2u2HnO5uSRNzO
d0KXCe0/NgkWcBWPUGZZAoGAb1KDORMs4GmMWCB81SeMnYHQ8VWa4c5BVQenV6Lx
HmLyFjlNTbFZAA3zjKP8/TP9ejjpr5ySb6QzmERhKc/FfjCmNrzv8WGfqL0+h337
EOAoDirqov2xzgdqroahWC7MCzXH6EJucp6XnZ+N5r5mJuTemtZly97pM+in157t
CkkCgYARUacE8AdrVebkzTDKvohzKhKzYpHQ4p8rHb585Dzmh5H3ew2+XOroAkHC
CXbW/M3HZ5mgsb4JJAdRXIClA4dL9yZMUeIoBRb5Hd+D3jAXPU+dCj9O0j9FyyWe
g9yaAi9Hk42khdQgdpe9BNND3XBQ6X9rlsLd0zTDf+MPBRfHog==
-----END RSA PRIVATE KEY-----`

const rsaServerPublicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAhm3phfz4rrM98MJoilYc
sY9ikNGEim9xIPtWsJSDja+ttm9VhjOzxN+fJta/Jb2TXAF9z7sT1pdy8sagupNi
n1SBF/ejgGuw093xAkpaSrzDB7R6oHsPa/+6qnSuleFalGiyOFylyhPGBZwtJwXG
eSGCLFmFEIMHlgEwyHR/qQUoEFDMs9JdVqSwKeMvbCAN+ea4PUq2u1zcrtxjP83d
5Buxx4/BRQ2ycLFeUIqj2LzlTBWZRnPerCr3vcQnRZWXenSqhlLO/Lwi9/rV2xYA
w27psNA0J6l3awDer/bb+Jo9JX1dJiRm1nRcjg1791DwsCutHqx6oxdgCssyyfXF
xQIDAQAB
-----END PUBLIC KEY-----`

func TestUyzUser() {
	client, err := user.New(user.ClientConfig{
		BaseURL:          "https://eyes.phrynus.cn/api/user",
		AppID:            1003,
		AppKey:           "II1PFsYQFMrNPXeAKXRR4COiVrlPKxXj",
		Version:          "1.0.0",
		VersionIndex:     "web",
		ClientPrivateKey: rsaClientPrivateKey,
		ServerPublicKey:  rsaServerPublicKey,
		EncryptionMode:   user.EncryptionRSA,
		DisableSignature: false, // flip to true if the backend skips MD5 signatures
		// EncodingMode:     user.EncodingHex, // used when selecting AES/DES/RC4
		// SymmetricKey:     "UOCljhqNxLYHQLqNKB8sGy9O355hU1zy", // required for AES/DES/RC4 modes
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
