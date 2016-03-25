package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/codegangsta/cli"
	etcdClient "github.com/coreos/etcd/client"
	"github.com/monder/route53-etcd/utils"
	"golang.org/x/net/context"
	"os"
)

type HostConfig struct {
	zoneId string
	domain string
	key    string
}

type ServerState struct {
	etcdAPI etcdClient.KeysAPI
	route53 *route53.Route53

	monitoredHosts []*HostConfig

	activeWatcher *etcdClient.Watcher

	configReloadCh chan bool
	hostReloadCh   chan *HostConfig
}

func watchNewHostPath(s *ServerState, prefix string) {
	resp, err := s.etcdAPI.Get(context.Background(), prefix, &etcdClient.GetOptions{
		Recursive: true,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	var watchPrefix = ""
	s.monitoredHosts = make([]*HostConfig, 0)
	for _, zone := range resp.Node.Nodes {
		zoneId := zone.Key[len(prefix):]
		zonePrefix := zone.Key + "/"
		fmt.Printf("Zone %s\n", zoneId)
		for _, domain := range zone.Nodes {
			domainName := domain.Key[len(zonePrefix):]
			fmt.Printf("  Register domain: %s\n", domainName)
			s.monitoredHosts = append(s.monitoredHosts, &HostConfig{
				zoneId: zoneId,
				domain: domainName,
				key:    domain.Value,
			})
			if watchPrefix == "" {
				watchPrefix = domain.Value
			} else {
				watchPrefix = utils.CommonPrefixForPatterns(watchPrefix, domain.Value)
			}
		}
	}
	for _, host := range s.monitoredHosts {
		registerService(s, host)
	}
	fmt.Printf("Start watching prefix %s\n", watchPrefix)
	watcher := s.etcdAPI.Watcher(watchPrefix, &etcdClient.WatcherOptions{
		AfterIndex: 0,
		Recursive:  true,
	})
	s.activeWatcher = &watcher
	for {
		resp, err = watcher.Next(context.Background())
		if &watcher != s.activeWatcher {
			break
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			continue
		}

		for _, h := range s.monitoredHosts {
			// TODO expired?
			switch resp.Action {
			case "delete":
				if !utils.MatchPathPrefix(resp.Node.Key, h.key) {
					continue
				}
			case "set":
				if !utils.MatchPath(resp.Node.Key, h.key) {
					continue
				}
				if resp.PrevNode != nil && resp.Node.Value == resp.PrevNode.Value {
					fmt.Printf("Nothing changed\n")
					continue
				}
			default:
				continue
			}
			fmt.Printf("Matched update %s%s/%s\n", prefix, h.zoneId, h.domain)
			s.hostReloadCh <- h
		}
	}
}

func watchConfig(s *ServerState, prefix string) {
	watcher := s.etcdAPI.Watcher(prefix, &etcdClient.WatcherOptions{
		AfterIndex: 0,
		Recursive:  true,
	})
	fmt.Println("Loading initial config")
	s.configReloadCh <- true
	for {
		_, err := watcher.Next(context.Background())
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			continue
		}
		s.configReloadCh <- true
	}
}

func runServer(c *cli.Context) {
	state := &ServerState{
		etcdAPI: utils.GetEtcdKeysAPI(c),
		route53: route53.New(session.New()),

		monitoredHosts: make([]*HostConfig, 0),

		configReloadCh: make(chan bool),
		hostReloadCh:   make(chan *HostConfig),
	}

	go watchConfig(state, c.GlobalString("etcd-prefix"))

	for {
		select {
		case <-state.configReloadCh:
			go watchNewHostPath(state, c.GlobalString("etcd-prefix"))
		case host := <-state.hostReloadCh:
			registerService(state, host)
		}
	}
}

func main() {
	app := cli.NewApp()
	app.Version = "0.3.0"
	app.Usage = "Exposing IPs registred in etcd to route53"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "etcd-endpoints",
			Value:  "http://127.0.0.1:4001,http://127.0.0.1:2379",
			Usage:  "a comma-delimited list of etcd endpoints",
			EnvVar: "ETCDCTL_ENDPOINT",
		},
		cli.StringFlag{
			Name:  "etcd-prefix",
			Value: "/hosts/",
			Usage: "a keyspace for host configuration data in etcd",
		},
	}
	app.ArgsUsage = " "
	app.Action = runServer

	app.Run(os.Args)
}

func registerService(s *ServerState, hostConfig *HostConfig) {
	extractedIPs := make([]*route53.ResourceRecord, 0)
	var extractIPs func(*etcdClient.Node)
	extractIPs = func(node *etcdClient.Node) {
		if node.Dir {
			if utils.MatchPathPrefix(node.Key, hostConfig.key) {
				for _, n := range node.Nodes {
					extractIPs(n)
				}
			}
		} else {
			if utils.MatchPath(node.Key, hostConfig.key) {
				extractedIPs = append(extractedIPs, &route53.ResourceRecord{
					Value: aws.String(node.Value),
				})
			}
		}
	}
	resp, err := s.etcdAPI.Get(context.Background(), utils.PrefixForPattern(hostConfig.key), &etcdClient.GetOptions{
		Recursive: true,
	})
	if err != nil {
		fmt.Printf("Key not found %s\n", utils.PrefixForPattern(hostConfig.key))
		return
	}
	extractIPs(resp.Node)

	_, err = s.route53.ChangeResourceRecordSets(&route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String(route53.ChangeActionUpsert),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name:            aws.String(hostConfig.domain),
						Type:            aws.String(route53.RRTypeA),
						TTL:             aws.Int64(1),
						ResourceRecords: extractedIPs,
					},
				},
			},
		},
		HostedZoneId: aws.String(hostConfig.zoneId),
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
}
