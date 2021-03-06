package graph

import (
	"booking/models"
	"booking/util"
	"context"
	"fmt"
	"strconv"
	"time"
)

type subscriptionResolver struct{ *Resolver }

// Updated returns a object each 5 seconds
func (r *subscriptionResolver) MessageAdded(ctx context.Context, canteenUUID string, adminId int) (<-chan *models.Message, error) {

	//如果通道还没有启用，就启动通道
	if r.tunnels == nil {
		r.tunnels = make(map[string]*Tunnel)
	}

	if _, ok := r.tunnels[canteenUUID]; !ok {
		r.tunnels[canteenUUID] = &Tunnel{Name: canteenUUID, Observers: map[string]chan *models.Message{}}
		r.tunnels[canteenUUID].Observers = make(map[string]chan *models.Message)
	}

	tunnel := r.tunnels[canteenUUID]

	events := make(chan *models.Message)

	id := strconv.Itoa(adminId)
	if _, ok := tunnel.Observers[id]; !ok {
		fmt.Println(" create new observer", id)
		tunnel.Observers[id] = events
	}

	go func() {
		success := &models.Message{}
		success.Text = "connect success"
		success.Error = false
		success.CreatedAt = time.Now()
		events <- success

		for {
			select {
			case msg := <-tunnel.Observers[id]:
				fmt.Println("msg", util.PrettyJson(msg))
				if msg.ID != 0 {
					events <- msg
				}

			case <-ctx.Done():
				fmt.Println("close tunnel")
				close(events)
				delete(tunnel.Observers, id)
				//events = nil
				return
			}
		}
	}()

	return events, nil

}


func (r *subscriptionResolver) SubComment(ctx context.Context, roomName string, userId int) (<-chan *models.Comment, error) {
	fmt.Println("tunnel", roomName,userId)
	//如果通道还没有启用，就启动通道
	if r.commentTunnels == nil {
		r.commentTunnels = make(map[string]*CommentTunnel)
	}

	if _, ok := r.commentTunnels[roomName]; !ok {
		r.commentTunnels[roomName] = &CommentTunnel{Name: roomName, Observers: map[string]chan *models.Comment{}}
		r.commentTunnels[roomName].Observers = make(map[string]chan *models.Comment)
	}

	tunnel := r.commentTunnels[roomName]

	events := make(chan *models.Comment)

	id := strconv.Itoa(userId)
	if _, ok := tunnel.Observers[id]; !ok {
		fmt.Println(" create new observer", id)
		tunnel.Observers[id] = events
	}

	go func() {
		for {
			select {
			case msg := <-tunnel.Observers[id]:
				fmt.Println("msg", util.PrettyJson(msg))
				if msg.ID != 0 {
					events <- msg
				}

			case <-ctx.Done():
				fmt.Println("close tunnel")
				close(events)
				delete(tunnel.Observers, id)
				//events = nil
				return
			}
		}
		//<-ctx.Done()
	}()

	return events, nil

	//object := make(chan models.Message, 1)

	//ticker := time.NewTicker(5 * time.Second)
	//if r.Tunnel == nil {
	//	r.Tunnel = make(chan models.Message)
	//}

	//go func() {
	//	for {
	//		select {
	//		//case t := <-ticker.C:
	//		//	s := fmt.Sprintf("%s", t)
	//		//	fmt.Println("s", s)
	//		//	object <- s
	//		//case msg := <-r.Tunnel:
	//		//	object <- msg
	//		case msg := <-r.tunnels[canteenUUID].Observers[]:
	//
	//		case <-ctx.Done():
	//			//ticker.Stop()
	//			//close(object)
	//			return
	//		}
	//	}
	//}()

	//return object, nil

}

