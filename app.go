package main

import (
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	etcdClient "github.com/coreos/etcd/client"
	"github.com/davecgh/go-spew/spew"
	"github.com/monder/route53-etcd/utils"
	"golang.org/x/net/context"
	"log"
	"os"
	"strings"
)

var opts struct {
	etcdEndpoints string
	etcdPrefix    string
	help          bool
}

type HostConfig struct {
	zoneId string
	domain string
	key    string
}

var config struct {
	route53     *route53.Route53
	watchPrefix string
	keys        []*HostConfig
}

func init() {
	flag.StringVar(&opts.etcdEndpoints, "etcd-endpoints", "http://127.0.0.1:4001,http://127.0.0.1:2379", "a comma-delimited list of etcd endpoints")
	flag.StringVar(&opts.etcdPrefix, "etcd-prefix", "/hosts/", "etcd prefix")
	flag.BoolVar(&opts.help, "help", false, "print this message")
}

func main() {
	flag.Set("logtostderr", "true")

	flag.Parse()

	if flag.NArg() > 0 || opts.help {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTION]...\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(0)
	}

	configEtcd, err := etcdClient.New(etcdClient.Config{
		Endpoints: strings.Split(opts.etcdEndpoints, ","),
	})
	if err != nil {
		log.Fatal(err)
	}
	etcdKeys := etcdClient.NewKeysAPI(configEtcd)

	resp, err := etcdKeys.Get(context.Background(), opts.etcdPrefix, &etcdClient.GetOptions{
		Recursive: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	config.watchPrefix = ""
	config.keys = make([]*HostConfig, 0)
	config.route53 = route53.New(session.New())

	for _, zone := range resp.Node.Nodes {
		zoneId := zone.Key[len(opts.etcdPrefix):]
		zonePrefix := zone.Key + "/"
		fmt.Printf("Zone %s\n", zoneId)
		for _, domain := range zone.Nodes {
			domainName := domain.Key[len(zonePrefix):]
			fmt.Printf("  Register domain: %s\n", domainName)
			config.keys = append(config.keys, &HostConfig{
				zoneId: zoneId,
				domain: domainName,
				key:    domain.Value,
			})
			if config.watchPrefix == "" {
				config.watchPrefix = domain.Value
			} else {
				config.watchPrefix = utils.CommonPrefixForPatterns(config.watchPrefix, domain.Value)
			}
		}
	}

	watcher := etcdKeys.Watcher(config.watchPrefix, &etcdClient.WatcherOptions{
		AfterIndex: 0,
		Recursive:  true,
	})
	for {
		response, err := watcher.Next(context.Background())
		if err != nil {
			log.Fatal("Error occurred", err)
		}
		for _, h := range config.keys {
			// TODO expired?
			switch response.Action {
			case "delete":
				if !utils.MatchPathPrefix(response.Node.Key, h.key) {
					continue
				}
			case "set":
				if !utils.MatchPath(response.Node.Key, h.key) {
					continue
				}
				if response.PrevNode != nil && response.Node.Value == response.PrevNode.Value {
					fmt.Printf("nothing changed\n")
					continue
				}
			default:
				continue
			}
			fmt.Printf("Matched update %s%s/%s\n", opts.etcdPrefix, h.zoneId, h.domain)
			spew.Dump(response)
			registerService(etcdKeys, h)
		}
	}
}

func registerService(etcdKeys etcdClient.KeysAPI, hostConfig *HostConfig) {
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
	resp, err := etcdKeys.Get(context.Background(), utils.PrefixForPattern(hostConfig.key), &etcdClient.GetOptions{
		Recursive: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	extractIPs(resp.Node)

	resp2, err := config.route53.ChangeResourceRecordSets(&route53.ChangeResourceRecordSetsInput{
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
	spew.Dump(resp2)
}
