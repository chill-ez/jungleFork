package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/palantir/stacktrace"
)

// IPAddressReputationChecker checks the reputation of IP addresses
type IPAddressReputationChecker struct {
	log            *log.Logger
	reputation     map[string]float32
	reputationLock sync.RWMutex

	endpoint  string
	authToken string

	checkQueue chan string
}

// NewIPAddressReputationChecker initializes and returns a new IPAddressReputationChecker
func NewIPAddressReputationChecker(log *log.Logger, endpoint, authToken string) *IPAddressReputationChecker {
	return &IPAddressReputationChecker{
		log:        log,
		reputation: make(map[string]float32),
		checkQueue: make(chan string, 1000),

		endpoint:  endpoint,
		authToken: authToken,
	}
}

func (c *IPAddressReputationChecker) CanReceiveRewards(remoteAddress string) bool {
	c.reputationLock.RLock()
	defer c.reputationLock.RUnlock()
	badActorConfidence, present := c.reputation[remoteAddress]
	if !present {
		c.EnqueueAddressForChecking(remoteAddress)
		return true // let's be generous and reward until they're checked
	}
	return badActorConfidence < 0.95
}

func (c *IPAddressReputationChecker) EnqueueAddressForChecking(remoteAddress string) {
	c.reputationLock.RLock()
	defer c.reputationLock.RUnlock()
	if _, present := c.reputation[remoteAddress]; present || remoteAddress == "" {
		return
	}
	// make this function never block by simply dropping the request if the queue is full
	select {
	case c.checkQueue <- remoteAddress:
		c.log.Printf("Enqueued remote address %s for reputation checking", remoteAddress)
	default:
	}
}

func (c *IPAddressReputationChecker) Worker(ctx context.Context) {
	httpClient := http.Client{
		Timeout: 10 * time.Second,
	}
	for {
		select {
		case addressToCheck := <-c.checkQueue:
			time.Sleep(5 * time.Second) // TODO this rate limit might not be needed anymore
			url := fmt.Sprintf(c.endpoint, addressToCheck)
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				c.log.Println("error building http request:", stacktrace.Propagate(err, ""))
				continue
			}
			req.Header.Add("Authorization", "Bearer "+c.authToken)
			resp, err := httpClient.Do(req)
			if err != nil {
				c.log.Println("error checking IP reputation:", stacktrace.Propagate(err, ""))
				continue
			}
			if resp.StatusCode != http.StatusOK {
				c.log.Println("non-200 status code when checking IP reputation for address", addressToCheck)
				continue
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				c.log.Println("error reading response body:", stacktrace.Propagate(err, ""))
				continue
			}

			var response struct {
				Privacy struct {
					Proxy bool `json:"proxy"`
				} `json:"privacy"`
			}

			err = json.Unmarshal(body, &response)
			if err != nil {
				c.log.Println("error parsing response:", stacktrace.Propagate(err, ""))
				continue
			}

			if response.Privacy.Proxy {
				func() {
					c.reputationLock.Lock()
					defer c.reputationLock.Unlock()
					c.reputation[addressToCheck] = 1.0
				}()
				c.log.Printf("IP %v is bad actor", addressToCheck)
			}
		case <-ctx.Done():
			return
		}
	}

}
