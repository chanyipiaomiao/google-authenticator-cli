package main

import (
	"fmt"
	"github.com/chanyipiaomiao/hltool"
	"github.com/gizak/termui"
	"github.com/gosuri/uilive"
	"gopkg.in/alecthomas/kingpin.v2"
	"log"
	"os"
	"path"
	"runtime"
	"sort"
	"time"
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
	for _, v := range keys {
		n, t, err := hltool.TwoStepAuthGenByKey(v)
		if err != nil {
			continue
		}
		fmt.Printf("%s %s %d\n", m[v], n, t)
	}
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

func (s *Secret) Add(name, secret string) error {
	err := s.TwoStepDB.Set(map[string][]byte{
		secret: []byte(name),
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
	if name != "all" {
		r, err := s.TwoStepDB.Get([]string{name})
		if err != nil {
			return err
		}
		SortMapByKey(r)
	} else {
		r, err := s.TwoStepDB.GetAll()
		if err != nil {
			return err
		}
		SortMapByKey(r)
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

func ShowUI() {
	err := termui.Init()
	if err != nil {
		log.Fatalf("termui.Init() error: %s\n", err)
	}
	defer termui.Close()
	rows1 := [][]string{
		[]string{"header1", "header2", "header3"},
		[]string{"你好吗", "Go-lang is so cool", "Im working on Ruby"},
		[]string{"2016", "10", "11"},
	}

	table1 := termui.NewTable()
	table1.Rows = rows1
	table1.FgColor = termui.ColorWhite
	table1.BgColor = termui.ColorDefault
	table1.Y = 0
	table1.X = 0
	table1.Width = 62
	table1.Height = 7

	termui.Render(table1)
}

func ShowWindowUI() {

}

func cli() {
	app := kingpin.New("google-authenticator-cli", "模拟 Google Authenticator 验证器")

	add := app.Command("add", "添加secret")
	addName := add.Flag("name", "名称标识").Required().String()
	secret := add.Flag("secret", "二步验证里面生成的Secret,一般跟二维码一起展示").Required().String()

	del := app.Command("delete", "删除secret")
	deleteName := del.Flag("delete-secret", "名称标识").Required().String()

	show := app.Command("show", "显示所有的6位数字")
	showName := show.Flag("show-name", "显示指定的标识的6位数字").Default("all").String()

	ui := app.Command("ui", "打开终端")
	showUI := ui.Flag("show-ui", "显示终端UI").Default("false").Bool()

	c, err := app.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("parse cli args error: %s\n", err)
	}

	switch c {
	case "add":
		err := newSecret.Add(*addName, *secret)
		if err != nil {
			log.Fatalf("s.Add(*addName, *secret) error: %s\n", err)
		}
		fmt.Println("add ok.")
	case "delete":
		err := newSecret.Delete(*deleteName)
		if err != nil {
			log.Fatalf("s.Delete(*deleteName) error: %s\n", err)
		}
		fmt.Println("delete ok.")
	case "show":
		err := newSecret.List(*showName)
		if err != nil {
			log.Fatalf("s.List(*showName) error: %s\n", err)
		}
	}

	if *showUI {

		if runtime.GOOS == "window" {
			ShowWindowUI()
		} else {
			ShowUI()
		}

	}

}

func ui() {
	writer := uilive.New()
	// start listening for updates and render
	writer.Start()

	for i := 0; i <= 1000; i++ {
		fmt.Fprintf(writer, "Downloading.. (%d/%d) GB\n", i, 1000)
		time.Sleep(time.Millisecond * 5)
	}

	fmt.Fprintln(writer, "Finished: Downloaded 100GB")
	writer.Stop() // flush and stop rendering
}

func main() {
	//cli()
	ui()
}
