package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/mit-dci/pooldetective/logging"
	"github.com/mit-dci/pooldetective/util"
	"github.com/mit-dci/pooldetective/wire"
)

var locationID int

func main() {
	logging.SetLogLevel(util.GetLoglevelFromEnv())
	logging.Debugf("Starting coordinator...")
	locationID64, err := strconv.ParseInt(os.Getenv("LOCATIONID"), 10, 32)
	if err != nil {
		panic(err)
	}
	locationID = int(locationID64)

	initDocker()
	err = initNetwork()
	if err != nil {
		panic(err)
	}

	err = pullLatestImages()
	if err != nil {
		panic(err)
	}

	go imageChangeDetector()

	refresh := make(chan bool, 1)
	restart := make(chan int, 10)
	go subscribe(refresh, restart)
	go func() {
		for poID := range restart {
			err := restartStratumClient(poID)
			if err != nil {
				logging.Errorf("Unable to restart pool observer %d: %v", poID, err)
			}
		}
	}()

	refresh <- true
	for {
		req, err := wire.NewClient(os.Getenv("HUBHOST"), wire.PortCoordinator)
		if err != nil {
			logging.Errorf("Could not connect to coordinator hub: %v", err)
			time.Sleep(time.Second * 5)
			continue
		}

		for {
			<-refresh
			logging.Infof("Refreshing config")
			err = refreshConfig(req)
			if err != nil {
				logging.Errorf("Could not refresh config: %v", err)
				time.Sleep(time.Second * 10)
				break
			}
		}
		req.Close()
		time.Sleep(time.Second)
	}
}

func imageChangeDetector() {
	images := []string{"blockobserver", "stratumclient", "stratumserver"}
	imageIDs := make([]string, len(images))
	for i, img := range images {
		imageIDs[i], _ = getImage(getPrefixedImageName(img))
	}

	for {
		reapply := false
		time.Sleep(time.Second * 60)
		pullLatestImages()
		for i, img := range images {
			newImageID, _ := getImage(getPrefixedImageName(img))
			if newImageID != imageIDs[i] {
				imageIDs[i] = newImageID
				reapply = true
			}

		}
		if reapply {
			logging.Debugf("Found changed images, reapplying config")
			err := reapplyConfig()
			if err != nil {
				logging.Errorf("Could not reapply config: %v", err)
			}
		}
		checkBlockObserver()
	}

}

func refreshConfig(req *wire.PoolDetectiveConn) error {

	err := req.Send(&wire.CoordinatorGetConfigRequestMsg{LocationID: locationID})
	if err != nil {
		return fmt.Errorf("Could not send CoordinatorGetConfigRequestMsg request to hub: %v", err)

	}

	msg, ok, err := req.Recv()
	if err != nil || !ok {
		return fmt.Errorf("Could not receive CoordinatorGetConfigResponseMsg response from hub: %v", err)

	}

	config, ok := msg.(*wire.CoordinatorGetConfigResponseMsg)
	if !ok {
		errMsg, ok := msg.(*wire.ErrorMsg)
		if ok {
			return fmt.Errorf("Error from hub: %v", errMsg.Error)
		}
		return fmt.Errorf("CoordinatorGetConfigResponseMsg response from hub is wrong type: %T", msg)
	}

	err = applyConfig(config)
	if err != nil {
		logging.Errorf("Could not apply config: %v", err)
	}
	return nil
}

var lastConfig *wire.CoordinatorGetConfigResponseMsg

func reapplyConfig() error {
	return applyConfig(lastConfig)
}

func applyConfig(config *wire.CoordinatorGetConfigResponseMsg) error {
	lastConfig = config
	logging.Debugf("Starting %d stratum servers and %d stratum clients\n", len(config.StratumServers), len(config.StratumClientPoolObserverIDs))

	err := checkBlockObserver()
	if err != nil {
		logging.Errorf("Failed to configure block observer: %v", err)
	}

	for _, i := range config.StratumClientPoolObserverIDs {
		logging.Debugf("Checking stratum observer %d", i)
		err := checkStratumClient(i)
		if err != nil {
			logging.Errorf("Failed to configure stratum client: %v", err)
		}
	}

	algorithmIDs := make([]int, len(config.StratumServers))
	for i, s := range config.StratumServers {
		algorithmIDs[i] = s.AlgorithmID
		logging.Debugf("Checking stratum server [Algorithm %d - Port %d - Protocol %d]", s.AlgorithmID, s.Port, s.Protocol)
		err := checkStratumServer(s.AlgorithmID, s.Port, s.Protocol)
		if err != nil {
			logging.Errorf("Failed to configure stratum server: %v", err)
		}
	}

	err = removeUnneededContainers("stratum-client", config.StratumClientPoolObserverIDs)
	if err != nil {
		logging.Errorf("Failed to remove unneeded stratum clients: %v", err)
	}

	err = removeUnneededContainers("stratum-server", algorithmIDs)
	if err != nil {
		logging.Errorf("Failed to remove unneeded stratum servers: %v", err)
	}

	return nil
}

func subscribe(refresh chan bool, restart chan int) {
	for {
		logging.Infof("Subscribing to Pub/Sub")
		sub, err := wire.NewSubscriber(os.Getenv("HUBHOST"), wire.PortPubSubSubscribers)
		if err != nil {
			logging.Errorf("Could not connect to subscriber hub: %v", err)
			time.Sleep(time.Second * 5)
			continue
		}
		logging.Infof("Subscribed to Pub/Sub")

		sub.Subscribe([]byte(fmt.Sprintf("co-%03d", locationID)))
		for {
			msg, _, _, err := sub.RecvSub()
			if err != nil {
				logging.Errorf("[Subscriber] Error receiving message: %v", err)
				break
			}
			logging.Infof("Received message from Pub/Sub")

			switch t := msg.(type) {
			case *wire.CoordinatorRefreshConfigMsg:
				logging.Infof("Received config refresh request from pubsub")
				refresh <- true
			case *wire.CoordinatorRestartPoolObserverMsg:
				logging.Infof("Received restart pool observer request from pubsub")
				restart <- t.PoolObserverID
			default:
				logging.Warnf("Received unknown message from Pub/Sub: %T", msg)
			}
		}
		sub.Close()
		time.Sleep(time.Second)
	}

}
