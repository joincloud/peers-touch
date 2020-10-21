package pubsub

import (
	"context"
	"fmt"

	"github.com/ipfs/go-ipfs/core/coreapi"
	"github.com/ipfs/interface-go-ipfs-core"
	"github.com/joincloud/peers-touch-go/codec"
	"github.com/joincloud/peers-touch-go/logger"
	"github.com/joincloud/peers-touch-go/peer"
	"github.com/pkg/errors"
)

var (
	DefaultCodecName = "json"
)

type Subscriber interface {
	Topic() string
	Unsubscribe() error
}

type subscriber struct {
	opts       *SubOptions
	ipfs       coreapi.CoreAPI
	ipfsPubSub iface.PubSubSubscription
	peerID     peer.PeerID
	handler    Handler
	codec      codec.Codec
}

func (s *subscriber) Topic() string {
	return s.opts.Topic
}

func (s *subscriber) Unsubscribe() error {
	return s.ipfsPubSub.Close()
}

func (s *subscriber) start(ctx context.Context) {
	for {
		msg, err := s.ipfsPubSub.Next(ctx)
		if err != nil {
			fmt.Printf("msg err: %s", err.Error())
			continue
		}

		logger.Debugf("receive msg from %s, %s", msg.From().String(), s.peerID.String())
		// ignore self msg
		if msg.From() == s.peerID {
			logger.Infof("ignore self msg %s", msg.From().String())
			continue
		}

		topic := msg.Topics()[0]
		logger.Debugf("receive msg topic %s, %s", topic, s.opts.Topic)
		if topic != s.opts.Topic {
			logger.Infof("ignore topic %s", topic)
			continue
		}

		evt := event{}
		err = s.codec.Unmarshal(msg.Data(), &evt)
		if err != nil {
			logger.Errorf("err: %s, content: %s", err, string(msg.Data()))
		}

		s.handler(&evt)
	}
}

func NewSubscriber(ctx context.Context, opts ...SubOption) (sub Subscriber, err error) {
	options := &SubOptions{}
	for _, opt := range opts {
		opt(options)
	}

	if options.Topic == "" {
		return nil, fmt.Errorf("wrong empty topic")
	}

	if options.codec == nil {
		options.codec = codec.Codecs[DefaultCodecName]()
	}

	pubSubSub, err := options.coreAPI.PubSub().Subscribe(ctx, options.Topic)
	if err != nil {
		return nil, err
	}

	id, err := options.coreAPI.Key().Self(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get id for user")
	}

	logger.Infof("sub id: %s", id.ID().String())

	s := &subscriber{
		opts:       options,
		ipfs:       coreapi.CoreAPI{},
		ipfsPubSub: pubSubSub,
		peerID:     id.ID(),
		codec:      options.codec,
		handler:    options.Handler,
	}

	go s.start(ctx)

	sub = s
	return
}
