package pub_sub

type Node interface {
	Subscribe(topic string, ch chan interface{})
	Publish(topic string, msg interface{}) bool
	Unsubscribe(topic string, ch chan interface{})
}
