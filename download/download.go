package download

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/xmdhs/gomclauncher/lang"
	"github.com/xmdhs/gomclauncher/launcher"
)

func (l Libraries) Downassets(i int, c chan int) error {
	e, done, ch := creatch(len(l.assetIndex.Objects), i)
	cxt, cancel := context.WithCancel(l.cxt)
	defer cancel()
	go func() {
		for _, v := range l.assetIndex.Objects {
			v := v
			ok := ver(launcher.Minecraft+`/assets/objects/`+v.Hash[:2]+`/`+v.Hash, v.Hash)
			if !ok {
				d := downinfo{
					url:      `https://resources.download.minecraft.net/` + v.Hash[:2] + `/` + v.Hash,
					path:     launcher.Minecraft + `/assets/objects/` + v.Hash[:2] + `/` + v.Hash,
					e:        e,
					Sha1:     v.Hash,
					done:     done,
					ch:       ch,
					cxt:      cxt,
					randurls: l.randurls,
				}
				select {
				case ch <- struct{}{}:
					go d.down()
				case <-cxt.Done():
					return
				}
			} else {
				select {
				case done <- struct{}{}:
				case <-cxt.Done():
					return
				}
			}
		}
	}()
	n := 0
	for {
		select {
		case <-done:
			n++
			c <- len(l.assetIndex.Objects) - n
			if n == len(l.assetIndex.Objects) {
				close(c)
				return nil
			}
		case err := <-e:
			return fmt.Errorf("Downassets: %w", err)
		}
	}
}

func ver(path, hash string) bool {
	if hash != "" {
		file, err := os.Open(path)
		defer file.Close()
		if err != nil {
			return false
		}
		m := sha1.New()
		if _, err := io.Copy(m, file); err != nil {
			return false
		}
		h := hex.EncodeToString(m.Sum(nil))
		if strings.ToTitle(h) == strings.ToTitle(hash) {
			return true
		}
		return false
	}
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true

}

func (l Libraries) Downlibrarie(i int, c chan int) error {
	e, done, ch := creatch(len(l.librarie.Libraries), i)
	cxt, cancel := context.WithCancel(l.cxt)
	defer cancel()
	go func() {
		for _, v := range l.librarie.Libraries {
			v := v
			path := launcher.Minecraft + `/libraries/` + v.Downloads.Artifact.Path
			if v.Downloads.Artifact.URL == "" {
				select {
				case done <- struct{}{}:
				case <-cxt.Done():
					return
				}
				continue
			}
			if !ver(path, v.Downloads.Artifact.Sha1) {
				d := downinfo{
					url:      v.Downloads.Artifact.URL,
					path:     path,
					e:        e,
					Sha1:     v.Downloads.Artifact.Sha1,
					done:     done,
					ch:       ch,
					cxt:      cxt,
					randurls: l.randurls,
				}
				select {
				case ch <- struct{}{}:
					go d.down()
				case <-cxt.Done():
					return
				}
			} else {
				select {
				case done <- struct{}{}:
				case <-cxt.Done():
					return
				}
			}
		}
	}()
	n := 0
	for {
		select {
		case <-done:
			n++
			c <- len(l.librarie.Libraries) - n
			if n == len(l.librarie.Libraries) {
				close(c)
				return nil
			}
		case err := <-e:
			return fmt.Errorf("Downlibrarie: %w", err)
		}
	}
}

var FileDownLoadFail = errors.New("file download fail")

func (l Libraries) Downjar(version string) error {
	path := launcher.Minecraft + `/versions/` + version + "/" + version + ".jar"
	if ver(path, l.librarie.Downloads.Client.Sha1) {
		return nil
	}
	_, t := l.auto()
	for i := 0; i < 4; i++ {
		if i == 3 {
			return FileDownLoadFail
		}
		err := get(l.cxt, source(l.librarie.Downloads.Client.URL, t), path)
		if err != nil {
			fmt.Println(lang.Lang("weberr"), source(l.librarie.Downloads.Client.URL, t), fmt.Errorf("Downjar: %w", err))
			t = l.fail(t)
			continue
		}
		if !ver(path, l.librarie.Downloads.Client.Sha1) {
			fmt.Println(lang.Lang("filecheckerr"), source(l.librarie.Downloads.Client.URL, t))
			t = l.fail(t)
			continue
		}
		break
	}
	return nil
}

type downinfo struct {
	url  string
	path string
	e    chan error
	Sha1 string
	done chan struct{}
	ch   chan struct{}
	cxt  context.Context
	*randurls
}

func (d downinfo) down() {
	_, f := d.auto()
	for i := 0; i < 7; i++ {
		if i == 6 {
			select {
			case d.e <- FileDownLoadFail:
			case <-d.cxt.Done():
			}
			break
		}
		err := get(d.cxt, source(d.url, f), d.path)
		if err != nil {
			fmt.Println(lang.Lang("weberr"), source(d.url, f), err)
			f = d.fail(f)
			continue
		}
		if !ver(d.path, d.Sha1) {
			fmt.Println(lang.Lang("filecheckerr"), source(d.url, f))
			f = d.fail(f)
			continue
		}
		select {
		case d.done <- struct{}{}:
			<-d.ch
			d.add(f)
		case <-d.cxt.Done():
			return
		}
		break
	}

}

func creatch(a, i int) (e chan error, done, ch chan struct{}) {
	e = make(chan error, a)
	done = make(chan struct{}, a)
	ch = make(chan struct{}, i)
	return
}
