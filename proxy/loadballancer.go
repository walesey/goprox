package proxy

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// A round robin load ballancer that maintains a list of urls and their active status
type LoadBallancer struct {
	items           []*loadBallancerItem
	index           int
	InnactiveTimout time.Duration
}

type loadBallancerItem struct {
	url         string
	active      bool
	innactiveAt time.Time
}

// NewLoadBallancer - create new round robin loadballancer
func NewLoadBallancer(urls ...string) *LoadBallancer {
	items := make([]*loadBallancerItem, len(urls))
	for index, url := range urls {
		items[index] = &loadBallancerItem{
			url:         url,
			active:      true,
			innactiveAt: time.Now(),
		}
	}
	return &LoadBallancer{
		items:           items,
		InnactiveTimout: 30 * time.Second,
	}
}

func (lb *LoadBallancer) cycleIndex() {
	lb.index = lb.index + 1
	if lb.index >= len(lb.items) {
		lb.index = 0
	}
}

// MakeRequest - loops through each url
func (lb *LoadBallancer) MakeRequest(url string, w http.ResponseWriter, r *http.Request) error {
	req, err := copyRequest(r)
	if err != nil {
		log.Printf("Error copying proxy request: %v", err)
		return err
	}

	for i := 0; i < len(lb.items); i = i + 1 {
		item := lb.items[lb.index]
		lb.cycleIndex()
		if item.active || time.Since(item.innactiveAt) >= lb.InnactiveTimout {
			item.active = true
			err := proxyHttpRequest(fmt.Sprintf("%v%v", item.url, url), req, w)
			if err != nil {
				log.Printf("Error making Http Request: %v", err)
				item.active = false
				item.innactiveAt = time.Now()
			} else {
				return nil
			}
		}
	}
	return fmt.Errorf("All URLs are innactive")
}
