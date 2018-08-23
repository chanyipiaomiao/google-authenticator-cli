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

// Add 增加
func (s *Secret) Add(t *hltool.TOTP) error {
	o, err := hltool.StructToBytes(t)
	if err != nil {
		return err
	}
	err = s.TwoStepDB.Set(map[string][]byte{
		t.Name: o,
	})
	if err != nil {
		return err
	}
	return nil
}

// Delete 数据库中删除指定的name
func (s *Secret) Delete(name string) error {
	err := s.TwoStepDB.Delete([]string{name})
	if err != nil {
		return err
	}
	return nil
}


func SortMapByKey(m map[string][]byte) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}

// formatPrint 格式化输出
func formatPrint(r map[string][]byte)  {
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

// List 列出所有的名称和6位数字
func (s *Secret) List(name string) error {
	fmt.Printf("%-20s %-15s %-5s\n", "Name", "Number", "Remaining time")
	fmt.Printf("%-20s %-15s %-5s\n", "----", "------", "--------------")

	if name != "all" {
		r, err := s.TwoStepDB.Get([]string{name})
		if err != nil {
			return err
		}
		formatPrint(r)

	} else {
		r, err := s.TwoStepDB.GetAll()
		if err != nil {
			return err
		}
		formatPrint(r)
	}

	return nil
}

// Save 保存6位数字到文件
func (s *Secret) Save(name, username, path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	r, err := s.TwoStepDB.Get([]string{name})
	if err != nil {
		return err
	}
	for _, v := range r {
		totp := new(hltool.TOTP)
		err := hltool.BytesToStruct(v, totp)
		if err != nil {
			continue
		}
		n, _, err := hltool.TwoStepAuthGenNumber(totp)
		if err != nil {
			continue
		}
		fmt.Fprintf(f, "%s\n%s", username, n)
	}

	return nil
}

func cli() {
	app := kingpin.New("google-authenticator-cli", "模拟 Google Authenticator 验证器")

	add := app.Command("add", "添加secret")
	addName := add.Flag("name", "名称标识").Required().String()
	secret := add.Flag("secret", "二步验证里面生成的Secret,一般跟二维码一起展示").String()
	qrCode := add.Flag("qrcode", "指定二维码图片路径").String() // 此选项和secret2选1
	algorithm := add.Flag("alg", "指定加密算法 SHA1|SHA256").Default("SHA1").String()

	del := app.Command("delete", "删除secret")
	deleteName := del.Flag("delete-name", "名称标识").Required().String()

	show := app.Command("show", "显示所有的6位数字")
	showSecret := show.Flag("show-name", "显示指定的标识的6位数字").Default("all").String()

	save := app.Command("save", "保存生成的6位数字到文件,文件格式: 第一行: 用户名  第二行: 6位数字")
	saveName := save.Flag("save-name", "指定要保存的名称").Required().String()
	username := save.Flag("username", "用户名,比如连接OPENVPN的用户名").Required().String()
	savePath := save.Flag("path", "文件存储路径").Required().String()

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
	case "save":
		err := newSecret.Save(*saveName, *username, *savePath)
		if err != nil {
			log.Fatalf("newSecret.Save(*saveName, *username, *savePath) error: %s\n", err)
		}
	}

}

func main() {
	cli()
}
