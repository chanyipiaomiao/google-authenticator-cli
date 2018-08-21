package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"sort"

	"github.com/chanyipiaomiao/hltool"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	secretDBName    = "twostep.db"
	secretTableName = "secret"
)

func SortMapByKey(m map[string][]byte) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}

type Secret struct {
	TwoStepDB *hltool.BoltDB
}

func NewSecret() (*Secret, error) {
	dbPath := path.Join(path.Dir(os.Args[0]), secretDBName)
	twostepDB, err := hltool.NewBoltDB(dbPath, secretTableName)
	if err != nil {
		return nil, err
	}
	return &Secret{TwoStepDB: twostepDB}, nil
}

func (s *Secret) Add(t *hltool.TOTP) error {
	o, err := hltool.StructToBytes(t)
	if err != nil {
		return err
	}
	err = s.TwoStepDB.Set(map[string][]byte{
		t.SecretKey: o,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Secret) Delete(secret string) error {
	err := s.TwoStepDB.Delete([]string{secret})
	if err != nil {
		return err
	}
	return nil
}

func (s *Secret) List(name string) error {
	fmt.Printf("%-20s %-15s %-5s\n", "Name", "Number", "Remaining time")
	fmt.Printf("%-20s %-15s %-5s\n", "----", "------", "--------------")

	if name != "all" {
		r, err := s.TwoStepDB.Get([]string{name})
		if err != nil {
			return err
		}

		for _, v := range SortMapByKey(r) {
			var totp *hltool.TOTP
			err := hltool.BytesToStruct(r[v], totp)
			if err != nil {
				continue
			}
			n, t, err := hltool.TwoStepAuthGenNumber(totp)
			if err != nil {
				continue
			}
			fmt.Printf("%-20s %-15s %-5d\n", r[v], n, t)
		}
	} else {
		r, err := s.TwoStepDB.GetAll()
		if err != nil {
			return err
		}

		for _, v := range SortMapByKey(r) {
			totp := new(hltool.TOTP)
			err := hltool.BytesToStruct(r[v], totp)
			if err != nil {
				continue
			}
			n, t, err := hltool.TwoStepAuthGenNumber(totp)
			if err != nil {
				continue
			}
			fmt.Printf("%-20s %-15s %-5d\n", totp.Name, n, t)
		}
	}

	return nil
}

var (
	newSecret *Secret
)

func init() {
	var err error
	newSecret, err = NewSecret()
	if err != nil {
		log.Fatalf(" NewSecret() error: %s\n", err)
	}
}

func cli() {
	app := kingpin.New("google-authenticator-cli", "模拟 Google Authenticator 验证器")

	add := app.Command("add", "添加secret")
	addName := add.Flag("name", "名称标识").Required().String()
	secret := add.Flag("secret", "二步验证里面生成的Secret,一般跟二维码一起展示").String()
	qrCode := add.Flag("qrcode", "指定二维码图片路径").String() // 此选项和secret2选1
	algorithm := add.Flag("alg", "指定加密算法 SHA1|SHA256").Default("SHA1").String()

	del := app.Command("delete", "删除secret")
	deleteName := del.Flag("delete-secret", "名称标识").Required().String()

	show := app.Command("show", "显示所有的6位数字")
	showSecret := show.Flag("show-secret", "显示指定的标识的6位数字").Default("all").String()

	c, err := app.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("parse cli args error: %s\n", err)
	}

	switch c {
	case "add":
		if *secret != "" {
			err := newSecret.Add(&hltool.TOTP{SecretKey: *secret, Name: *addName, Algorithm: *algorithm})
			if err != nil {
				log.Fatalf("s.Add(*addName, *secret) error: %s\n", err)
			}
		} else if *qrCode != "" {
			t, err := hltool.TwoStepAuthParseQRCode(*qrCode)
			if err != nil {
				log.Fatalf("hltool.TwoStepAuthParseQRCode(*qrCode) error: %s\n", err)
			}
			err = newSecret.Add(&hltool.TOTP{SecretKey: t.SecretKey, Name: *addName, Algorithm: t.Algorithm})
			if err != nil {
				log.Fatalf("s.Add(*addName, *secret) error: %s\n", err)
			}
		} else {
			log.Fatalf("--secret | --qrcode You must choose one\n")
		}
		fmt.Println("add ok.")
	case "delete":
		err := newSecret.Delete(*deleteName)
		if err != nil {
			log.Fatalf("s.Delete(*deleteName) error: %s\n", err)
		}
		fmt.Println("delete ok.")
	case "show":
		err := newSecret.List(*showSecret)
		if err != nil {
			log.Fatalf("s.List(*showName) error: %s\n", err)
		}
	}

}

func main() {
	cli()
}
