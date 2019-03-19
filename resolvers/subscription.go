package resolvers

import (
	"context"
	"fmt"
	"time"
)

type subscriptionResolver struct{}


// Updated returns a fake object each 5 seconds
func (sr *subscriptionResolver) MessageAdded(ctx context.Context, roomName string) (<-chan string, error){
	object := make(chan string, 1)

	ticker := time.NewTicker(5 * time.Second)

	go func() {
		for {
			select {
			case t := <-ticker.C:
				s := fmt.Sprintf("%s", t)
				fmt.Println("s",s )
				object <- s
			case <-ctx.Done():
				ticker.Stop()
				close(object)
				return
			}
		}
	}()

	return object, nil
}