package download

import (
	"math/rand"
	"strings"
	"sync"
	"time"
)

func (r *randurls) fail(typee string) string {
	if v, ok := r.typeweight.Load(typee); ok {
		i := v.(int)
		i--
		if i <= 0 {
			r.typeweight.Store(typee, 0)
		} else {
			r.typeweight.Store(typee, i)
		}
		for {
			lenmap, t := r.auto()
			if lenmap == 1 {
				break
			}
			if t != typee {
				return t
			}
		}
	}
	return typee
}

type arand struct {
	*rand.Rand
	*sync.Mutex
}

type randurls struct {
	typeweight sync.Map
	arand
	atype string
}

func newrandurls(typee string) *randurls {
	r := &randurls{}
	if typee == "" {
		r.typeweight.Store("vanilla", 5)
		r.typeweight.Store("bmclapi", 6)
		r.typeweight.Store("mcbbs", 9)
	} else {
		s := strings.Split(typee, "|")
		for _, v := range s {
			r.typeweight.Store(v, 5)
		}
	}
	r.Rand = rand.New(rand.NewSource(time.Now().Unix()))
	r.Mutex = &sync.Mutex{}
	r.atype = typee
	return r
}

func (r *randurls) auto() (int, string) {
	if r.atype != "" && !strings.Contains(r.atype, "|") {
		return 1, r.atype
	}
	i := 0
	lenmap := 0
	t := make([]string, 0, 4)
	b := make([]int, 0, 4)
	r.typeweight.Range(
		func(k, v interface{}) bool {
			lenmap++
			if v.(int) != 0 {
				t = append(t, k.(string))
				b = append(b, v.(int))
				i += v.(int)
			}
			return true
		})
	r.Lock()
	a := r.Intn(i) + 1
	r.Unlock()
	for i, v := range b {
		a = a - v
		if a <= 0 {
			return lenmap, t[i]
		}
	}
	panic(a)
}

func (r *randurls) add(typee string) {
	if v, ok := r.typeweight.Load(typee); ok {
		i := v.(int)
		i++
		r.typeweight.Store(typee, i)
	}
}
